package rewrite

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/int128/transerr/pkg/astio"
	"github.com/int128/transerr/pkg/log"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type toXerrors struct {
	changes int
}

func (t *toXerrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	if astutil.AddImport(pkg.Fset, file, xerrorsImportPath) {
		log.Printf("rewrite: added import %s", xerrorsImportPath)
		t.addChange()
	}
	if astutil.DeleteImport(pkg.Fset, file, pkgErrorsImportPath) {
		log.Printf("rewrite: deleted import %s", pkgErrorsImportPath)
		t.addChange()
	}
	if astutil.DeleteImport(pkg.Fset, file, "errors") {
		log.Printf("rewrite: deleted import %s", "errors")
		t.addChange()
	}
	if err := astio.Inspect(pkg, file, t); err != nil {
		return 0, xerrors.Errorf("could not inspect the file: %w", err)
	}
	if !astutil.UsesImport(file, "fmt") {
		if astutil.DeleteImport(pkg.Fset, file, "fmt") {
			log.Printf("rewrite: deleted import %s", "fmt")
			t.addChange()
		}
	}
	ast.SortImports(pkg.Fset, file)
	return t.changes, nil
}

func (t *toXerrors) addChange() {
	t.changes++
}

func (t *toXerrors) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return t.pkgErrorsFunctionCall(p, call, pkg, fun)
	case "errors":
		return t.goErrorsFunctionCall(p, call, pkg, fun)
	case "fmt":
		return t.goFmtFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (t *toXerrors) goErrorsFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New", "Unwrap", "As", "Is":
		log.Printf("rewrite: %s: errors.%s() -> xerrors.%s()", p, functionName, functionName)
		pkg.Name = "xerrors"
		t.addChange()
		return nil
	}
	return nil
}

func (t *toXerrors) goFmtFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		log.Printf("rewrite: %s: fmt.Errorf() -> xerrors.Errorf()", p)
		pkg.Name = "xerrors"
		t.addChange()
		return nil
	}
	return nil
}

func (t *toXerrors) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("rewrite: %s: pkg/errors.Wrapf() -> xerrors.Errorf()", p)
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
		log.Printf("rewrite: %s: pkg/errors.%s() -> xerrors.%s()", p, functionName, functionName)
		pkg.Name = "xerrors"
		t.addChange()
		return nil

	case "Cause":
		log.Printf("rewrite: %s: pkg/errors.Cause() -> xerrors.Unwrap()", p)
		pkg.Name = "xerrors"
		fun.Sel.Name = "Unwrap"
		t.addChange()
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite pkg/errors.%s() -> xerrors", p, functionName)
	pkg.Name = "xerrors"
	t.addChange()
	return nil
}
