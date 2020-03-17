package rewrite

import (
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

type toGoErrors struct{}

func (t *toGoErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toGoErrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, xerrors.Errorf("could not inspect the file: %w", err)
	}
	if v.needImportFmt == 0 && v.needImportErrors == 0 {
		log.Printf("rewrite: %s: no change", astio.Filename(pkg, file))
		return 0, nil
	}
	n := t.replaceImports(pkg, file, v.needImportFmt, v.needImportErrors)
	return v.needImportFmt + v.needImportErrors + n, nil
}

func (*toGoErrors) replaceImports(pkg *packages.Package, file *ast.File, needImportFmt, needImportErrors int) int {
	var n int
	if needImportFmt > 0 {
		if astutil.AddImport(pkg.Fset, file, "fmt") {
			n++
			log.Printf("rewrite: %s: + import %s", astio.Filename(pkg, file), "fmt")
		}
	}
	if needImportErrors > 0 {
		if astutil.AddImport(pkg.Fset, file, "errors") {
			n++
			log.Printf("rewrite: %s: + import %s", astio.Filename(pkg, file), "errors")
		}
	}
	if astutil.DeleteImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), xerrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
	}
	if n > 0 {
		ast.SortImports(pkg.Fset, file)
	}
	return n
}

type toGoErrorsVisitor struct {
	needImportFmt    int
	needImportErrors int
}

func (v *toGoErrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(p, call, pkg, fun)
	case xerrorsImportPath:
		return v.xerrorsFunctionCall(p, pkg, fun)
	}
	return nil
}

func (v *toGoErrorsVisitor) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
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
		b.Value = strings.TrimSuffix(b.Value, `"`) + `: %w"`

		log.Printf("rewrite: %s: pkg/errors.Wrapf() -> fmt.Errorf()", p)
		v.needImportFmt++
		return nil

	case "Errorf":
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"
		log.Printf("rewrite: %s: pkg/errors.Errorf() -> fmt.Errorf()", p)
		v.needImportFmt++
		return nil

	case "New":
		pkg.Name = "errors"
		log.Printf("rewrite: %s: pkg/errors.%s() -> errors.%s()", p, functionName, functionName)
		v.needImportErrors++
		return nil

	case "Cause":
		pkg.Name = "errors"
		fun.Sel.Name = "Unwrap"
		log.Printf("rewrite: %s: pkg/errors.Cause() -> errors.Unwrap()", p)
		v.needImportErrors++
		return nil
	}

	pkg.Name = "errors"
	log.Printf("rewrite: %s: NOTE: you need to manually rewrite pkg/errors.%s() -> errors", p, functionName)
	v.needImportErrors++
	return nil
}

func (v *toGoErrorsVisitor) xerrorsFunctionCall(p token.Position, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		pkg.Name = "fmt"
		fun.Sel.Name = "Errorf"
		log.Printf("rewrite: %s: xerrors.Errorf() -> fmt.Errorf()", p)
		v.needImportFmt++
		return nil

	case "New", "Unwrap", "As", "Is":
		pkg.Name = "errors"
		log.Printf("rewrite: %s: xerrors.%s() -> errors.%s()", p, functionName, functionName)
		v.needImportErrors++
		return nil
	}

	pkg.Name = "errors"
	log.Printf("rewrite: %s: NOTE: you need to manually rewrite xerrors.%s() -> errors", p, functionName)
	v.needImportErrors++
	return nil
}
