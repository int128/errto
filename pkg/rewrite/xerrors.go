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

type toXerrors struct{}

func (t *toXerrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toXerrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, xerrors.Errorf("could not inspect the file: %w", err)
	}
	if v.needImport == 0 {
		log.Printf("rewrite: %s: no change", astio.Filename(pkg, file))
		return 0, nil
	}
	n := t.replaceImports(pkg, file)
	return v.needImport + n, nil
}

func (*toXerrors) replaceImports(pkg *packages.Package, file *ast.File) int {
	var n int
	if astutil.AddImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("rewrite: %s: + import %s", astio.Filename(pkg, file), xerrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, "errors") {
		n++
		log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), "errors")
	}
	if !astutil.UsesImport(file, "fmt") {
		if astutil.DeleteImport(pkg.Fset, file, "fmt") {
			n++
			log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), "fmt")
		}
	}
	if n > 0 {
		ast.SortImports(pkg.Fset, file)
	}
	return n
}

type toXerrorsVisitor struct {
	needImport int
}

func (v *toXerrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(p, call, pkg, fun)
	case "errors":
		return v.goErrorsFunctionCall(p, call, pkg, fun)
	case "fmt":
		return v.goFmtFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (v *toXerrorsVisitor) goErrorsFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New", "Unwrap", "As", "Is":
		log.Printf("rewrite: %s: errors.%s() -> xerrors.%s()", p, functionName, functionName)
		pkg.Name = "xerrors"
		v.needImport++
		return nil
	}
	return nil
}

func (v *toXerrorsVisitor) goFmtFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		log.Printf("rewrite: %s: fmt.Errorf() -> xerrors.Errorf()", p)
		pkg.Name = "xerrors"
		v.needImport++
		return nil
	}
	return nil
}

func (v *toXerrorsVisitor) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
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

		log.Printf("rewrite: %s: pkg/errors.Wrapf() -> xerrors.Errorf()", p)
		v.needImport++
		return nil

	case "Errorf", "New":
		pkg.Name = "xerrors"
		log.Printf("rewrite: %s: pkg/errors.%s() -> xerrors.%s()", p, functionName, functionName)
		v.needImport++
		return nil

	case "Cause":
		pkg.Name = "xerrors"
		fun.Sel.Name = "Unwrap"
		log.Printf("rewrite: %s: pkg/errors.Cause() -> xerrors.Unwrap()", p)
		v.needImport++
		return nil
	}

	pkg.Name = "xerrors"
	log.Printf("rewrite: %s: NOTE: you need to manually rewrite pkg/errors.%s() -> xerrors", p, functionName)
	v.needImport++
	return nil
}
