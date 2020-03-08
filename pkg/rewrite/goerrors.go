package rewrite

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/int128/transerr/pkg/astio"
	"github.com/int128/transerr/pkg/log"
	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type toGoErrorsVisitor struct {
	changes int
}

func (v *toGoErrorsVisitor) Changes() int {
	return v.changes
}

func (v *toGoErrorsVisitor) addChange() {
	v.changes++
}

func (v *toGoErrorsVisitor) RewriteImports(pkg *packages.Package, file *ast.File) error {
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.IMPORT:
				specs := make([]ast.Spec, 0)
				for _, spec := range decl.Specs {
					p := astio.Position(pkg, spec)
					switch spec := spec.(type) {
					case *ast.ImportSpec:
						path, err := strconv.Unquote(spec.Path.Value)
						if err != nil {
							return xerrors.Errorf("%s: import expects a quoted string: %w", p, err)
						}
						switch path {
						case pkgErrorsImportPath:
							log.Printf("%s: rewrite: import pkg/errors -> errors, fmt", p)
							specs = append(specs,
								&ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote("errors")}},
								&ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote("fmt")}},
							)
						default:
							specs = append(specs, spec)
						}
					}
				}
				decl.Specs = specs
			}
		}
	}
	ast.SortImports(pkg.Fset, file)
	return nil
}

func (v *toGoErrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (v *toGoErrorsVisitor) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("%s: rewrite: pkg/errors.Wrapf() -> fmt.Errorf()", p)
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"

		// reorder the args
		a := call.Args
		args := make([]ast.Expr, 0)
		args = append(args, a[1])
		args = append(args, a[2:]...)
		args = append(args, a[0])
		call.Args = args

		// append %w to the format arg
		b, ok := a[1].(*ast.BasicLit)
		if !ok {
			return xerrors.Errorf("2nd argument of Wrapf must be a literal but %T", a[1])
		}
		if b.Kind != token.STRING {
			return xerrors.Errorf("2nd argument of Wrapf must be a string but %s", b.Kind)
		}
		b.Value = fmt.Sprintf(`"%s: %%w"`, strings.Trim(b.Value, `"`))
		v.addChange()
		return nil

	case "Errorf":
		log.Printf("%s: rewrite: pkg/errors.Errorf() -> fmt.Errorf()", p)
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"
		v.addChange()
		return nil

	case "New":
		log.Printf("%s: rewrite: pkg/errors.New() -> errors.New()", p)
		pkg.Name = "errors"
		fun.Sel.Name = "New"
		v.addChange()
		return nil

	case "Cause":
		log.Printf("%s: rewrite: pkg/errors.Cause() -> errors.Unwrap()", p)
		pkg.Name = "errors"
		fun.Sel.Name = "Unwrap"
		v.addChange()
		return nil

	default:
		log.Printf("%s: NOTE: you need to manually rewrite pkg/errors.%s() -> errors", p, functionName)
		pkg.Name = "errors"
		v.addChange()
		return nil
	}
}
