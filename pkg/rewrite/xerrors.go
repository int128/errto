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

type toXerrors struct {
	changes int
}

func (t *toXerrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	if err := t.transformImports(pkg, file); err != nil {
		return 0, xerrors.Errorf("could not rewrite the imports: %w", err)
	}
	if err := astio.Inspect(pkg, file, t); err != nil {
		return 0, xerrors.Errorf("could not inspect the file: %w", err)
	}
	return t.changes, nil
}

func (t *toXerrors) addChange() {
	t.changes++
}

func (t *toXerrors) transformImports(pkg *packages.Package, file *ast.File) error {
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.IMPORT:
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
							log.Printf("%s: rewrite: import pkg/errors -> xerrors", p)
							spec.Path.Value = strconv.Quote(xerrorsImportPath)
							t.addChange()
						}
					}
				}
			}
		}
	}
	ast.SortImports(pkg.Fset, file)
	return nil
}

func (t *toXerrors) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return t.pkgErrorsFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (t *toXerrors) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("%s: rewrite: pkg/errors.Wrapf() -> xerrors.Errorf()", p)
		pkg.Name = "xerrors"
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
		t.addChange()
		return nil

	case "Errorf", "New":
		log.Printf("%s: rewrite: pkg/errors.%s() -> xerrors.%s()", p, functionName, functionName)
		pkg.Name = "xerrors"
		t.addChange()
		return nil

	case "Cause":
		log.Printf("%s: rewrite: pkg/errors.Cause() -> xerrors.Unwrap()", p)
		pkg.Name = "xerrors"
		fun.Sel.Name = "Unwrap"
		t.addChange()
		return nil

	default:
		log.Printf("%s: NOTE: you need to manually rewrite pkg/errors.%s() -> xerrors", p, functionName)
		pkg.Name = "xerrors"
		t.addChange()
		return nil
	}
}
