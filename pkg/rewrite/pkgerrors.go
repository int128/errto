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

func (v *toPkgErrorsVisitor) goErrorsFunctionCall(p token.Position, _ *ast.CallExpr, _ *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		log.Printf("rewrite: %s: errors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		v.needImport++
		return nil

	case "Unwrap":
		log.Printf("rewrite: %s: errors.Unwrap() -> pkg/errors.Cause()", p)
		fun.Sel.Name = "Cause"
		v.needImport++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite errors.%s() -> pkg/errors", p, functionName)
	v.needImport++
	return nil
}

func (v *toPkgErrorsVisitor) goFmtFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		pkg.Name = "errors"
		v.needImport++
		return errorfFromGoErrorsToPkgErrors(p, call, pkg, fun)
	}
	return nil
}

func (v *toPkgErrorsVisitor) xerrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		log.Printf("rewrite: %s: xerrors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		v.needImport++
		return nil

	case "Errorf":
		log.Printf("rewrite: %s: xerrors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		v.needImport++
		return errorfFromGoErrorsToPkgErrors(p, call, pkg, fun)

	case "Unwrap":
		log.Printf("rewrite: %s: xerrors.Unwrap() -> pkg/errors.Cause()", p)
		pkg.Name = "errors"
		fun.Sel.Name = "Cause"
		v.needImport++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite xerrors.%s() -> pkg/errors", p, functionName)
	pkg.Name = "errors"
	v.needImport++
	return nil
}

// errorfFromGoErrorsToPkgErrors rewrites the Errorf function call
// from fmt.Errorf() or xerrors.Errorf() to pkg/errors.Errorf().
// If the Errorf wraps an error, it rewrites to pkg/errors.Wrapf().
func errorfFromGoErrorsToPkgErrors(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	a := call.Args
	if len(a) < 2 {
		log.Printf("rewrite: %s: %s.Errorf() -> pkg/errors.Errorf()", p, pkg.Name)
		return nil
	}

	// check if Errorf wraps an error
	b, ok := a[0].(*ast.BasicLit)
	if !ok {
		return xerrors.Errorf("1st argument of Errorf must be a literal but %T", a[0])
	}
	if b.Kind != token.STRING {
		return xerrors.Errorf("1st argument of Errorf must be a string but %s", b.Kind)
	}
	if !strings.HasSuffix(b.Value, `: %w"`) {
		log.Printf("rewrite: %s: %s.Errorf(...) -> pkg/errors.Errorf(...)", p, pkg.Name)
		return nil
	}

	log.Printf("rewrite: %s: %s.Errorf(..., err) -> pkg/errors.Wrapf(err, ...)", p, pkg.Name)
	fun.Sel.Name = "Wrapf"

	// trim the suffix
	b.Value = strings.TrimSuffix(b.Value, `: %w"`) + `"`

	// reorder args
	args := make([]ast.Expr, 0)
	args = append(args, a[len(a)-1])
	args = append(args, a[0])
	if len(a) > 2 {
		args = append(args, a[1:len(a)-2]...)
	}
	call.Args = args
	return nil
}
