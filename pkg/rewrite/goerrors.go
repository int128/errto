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

type toGoErrors struct {
	changes int
}

func (t *toGoErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	if err := t.transformImports(pkg, file); err != nil {
		return 0, xerrors.Errorf("could not rewrite the imports: %w", err)
	}
	if err := astio.Inspect(pkg, file, t); err != nil {
		return 0, xerrors.Errorf("could not inspect the file: %w", err)
	}
	return t.changes, nil
}

func (t *toGoErrors) addChange() {
	t.changes++
}

func (t *toGoErrors) transformImports(pkg *packages.Package, file *ast.File) error {
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
						case pkgErrorsImportPath, xerrorsImportPath:
							log.Printf("rewrite: %s: import %s -> errors, fmt", p, path)
							specs = append(specs,
								&ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote("errors")}},
								&ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote("fmt")}},
							)
							t.addChange()
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

func (t *toGoErrors) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return t.pkgErrorsFunctionCall(p, call, pkg, fun)
	case xerrorsImportPath:
		return t.xerrorsFunctionCall(p, pkg, fun)
	}
	return nil
}

func (t *toGoErrors) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("rewrite: %s: pkg/errors.Wrapf() -> fmt.Errorf()", p)
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
		t.addChange()
		return nil

	case "Errorf":
		log.Printf("rewrite: %s: pkg/errors.Errorf() -> fmt.Errorf()", p)
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"
		t.addChange()
		return nil

	case "New":
		log.Printf("rewrite: %s: pkg/errors.%s() -> errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		t.addChange()
		return nil

	case "Cause":
		log.Printf("rewrite: %s: pkg/errors.Cause() -> errors.Unwrap()", p)
		pkg.Name = "errors"
		fun.Sel.Name = "Unwrap"
		t.addChange()
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite pkg/errors.%s() -> errors", p, functionName)
	pkg.Name = "errors"
	t.addChange()
	return nil
}

func (t *toGoErrors) xerrorsFunctionCall(p token.Position, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		log.Printf("rewrite: %s: xerrors.Errorf() -> fmt.Errorf()", p)
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"
		t.addChange()
		return nil

	case "New", "Unwrap", "As", "Is":
		log.Printf("rewrite: %s: xerrors.%s() -> errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		t.addChange()
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite xerrors.%s() -> errors", p, functionName)
	pkg.Name = "errors"
	t.addChange()
	return nil
}
