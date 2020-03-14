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

type toPkgErrors struct {
	changes int
}

func (t *toPkgErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	if astutil.AddImport(pkg.Fset, file, pkgErrorsImportPath) {
		log.Printf("rewrite: added import %s", pkgErrorsImportPath)
		t.addChange()
	}
	if astutil.DeleteImport(pkg.Fset, file, xerrorsImportPath) {
		log.Printf("rewrite: deleted import %s", xerrorsImportPath)
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

func (t *toPkgErrors) addChange() {
	t.changes++
}

func (t *toPkgErrors) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case xerrorsImportPath:
		return t.xerrorsFunctionCall(p, call, pkg, fun)
	case "errors":
		return t.goErrorsFunctionCall(p, call, pkg, fun)
	case "fmt":
		return t.goFmtFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (t *toPkgErrors) goErrorsFunctionCall(p token.Position, _ *ast.CallExpr, _ *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		log.Printf("rewrite: %s: errors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		return nil

	case "Unwrap":
		log.Printf("rewrite: %s: errors.Unwrap() -> pkg/errors.Cause()", p)
		fun.Sel.Name = "Cause"
		t.addChange()
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite errors.%s() -> pkg/errors", p, functionName)
	return nil
}

func (t *toPkgErrors) goFmtFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		pkg.Name = "errors"
		t.addChange()
		return errorfFromGoErrorsToPkgErrors(p, call, pkg, fun)
	}
	return nil
}

func (t *toPkgErrors) xerrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "New":
		log.Printf("rewrite: %s: xerrors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		t.addChange()
		return nil

	case "Errorf":
		log.Printf("rewrite: %s: xerrors.%s() -> pkg/errors.%s()", p, functionName, functionName)
		pkg.Name = "errors"
		t.addChange()
		return errorfFromGoErrorsToPkgErrors(p, call, pkg, fun)

	case "Unwrap":
		log.Printf("rewrite: %s: xerrors.Unwrap() -> pkg/errors.Cause()", p)
		pkg.Name = "errors"
		fun.Sel.Name = "Cause"
		t.addChange()
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite xerrors.%s() -> pkg/errors", p, functionName)
	pkg.Name = "errors"
	t.addChange()
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
