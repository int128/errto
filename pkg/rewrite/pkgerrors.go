package rewrite

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/int128/errto/pkg/astio"
	"github.com/int128/errto/pkg/log"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type toPkgErrors struct{}

func (t *toPkgErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toPkgErrorsVisitor
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

func (*toPkgErrors) replaceImports(pkg *packages.Package, file *ast.File) int {
	var n int
	if astutil.AddImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("rewrite: %s: + import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("rewrite: %s: - import %s", astio.Filename(pkg, file), xerrorsImportPath)
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
	ast.SortImports(pkg.Fset, file)
	return n
}

type toPkgErrorsVisitor struct {
	needImport int
}

func (v *toPkgErrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case xerrorsImportPath:
		return v.xerrorsFunctionCall(p, call, pkg, fun)
	case "errors":
		return v.goErrorsFunctionCall(p, call, pkg, fun)
	case "fmt":
		return v.goFmtFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (v *toPkgErrorsVisitor) goErrorsFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		replacePackageFunctionCall(p, pkg, fun, "errors", "")
		v.needImport++
		return nil

	case "Unwrap":
		replacePackageFunctionCall(p, pkg, fun, "errors", "Cause")
		v.needImport++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "errors"
	v.needImport++
	return nil
}

func (v *toPkgErrorsVisitor) goFmtFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		v.needImport++
		return replaceErrorfWithPkgErrors(p, call, pkg, fun)
	}
	return nil
}

func (v *toPkgErrorsVisitor) xerrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		replacePackageFunctionCall(p, pkg, fun, "errors", "New")
		v.needImport++
		return nil

	case "Errorf":
		v.needImport++
		return replaceErrorfWithPkgErrors(p, call, pkg, fun)

	case "Unwrap":
		replacePackageFunctionCall(p, pkg, fun, "errors", "Cause")
		v.needImport++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "errors"
	v.needImport++
	return nil
}

// replaceErrorfWithPkgErrors rewrites the Errorf function call
// from fmt.Errorf() or xerrors.Errorf() to pkg/errors.Errorf().
// If the Errorf wraps an error, it rewrites to pkg/errors.Wrapf().
func replaceErrorfWithPkgErrors(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	if len(call.Args) < 2 {
		replacePackageFunctionCall(p, pkg, fun, "errors", "Errorf")
		return nil
	}
	b, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return xerrors.Errorf("%s: 1st argument of Errorf must be a literal but %T", p, call.Args[0])
	}
	if b.Kind != token.STRING {
		return xerrors.Errorf("%s: 1st argument of Errorf must be a string but %s", p, b.Kind)
	}
	if !strings.HasSuffix(b.Value, `: %w"`) {
		replacePackageFunctionCall(p, pkg, fun, "errors", "Errorf")
		return nil
	}

	// trim the suffix `: %w`
	b.Value = strings.TrimSuffix(b.Value, `: %w"`) + `"`

	// reorder the args
	var args []ast.Expr
	args = append(args, call.Args[len(call.Args)-1])
	args = append(args, call.Args[0])
	if len(call.Args) > 2 {
		args = append(args, call.Args[1:len(call.Args)-2]...)
	}
	call.Args = args

	replacePackageFunctionCall(p, pkg, fun, "errors", "Wrapf")
	return nil
}
