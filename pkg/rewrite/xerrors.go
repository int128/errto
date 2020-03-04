package rewrite

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/int128/transerr/pkg/log"
	"golang.org/x/xerrors"
)

type toXerrorsVisitor struct {
	changes int
}

func (v *toXerrorsVisitor) Changes() int {
	return v.changes
}

func (v *toXerrorsVisitor) Import(p token.Position, spec *ast.ImportSpec) error {
	packagePath := strings.Trim(spec.Path.Value, `"`)
	switch packagePath {
	case pkgErrorsImportPath:
		log.Printf("%s: rewriting the import with %s", p, xerrorsImportPath)
		spec.Path.Value = fmt.Sprintf(`"%s"`, xerrorsImportPath)
		v.changes++
		return nil
	}
	return nil
}

func (v *toXerrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := resolvedPkgName.Imported().Path()
	switch packagePath {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(p, call, pkg, fun)
	}
	return nil
}

func (v *toXerrorsVisitor) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("%s: rewriting the function call with xerrors.Errorf()", p)
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
		v.changes++
		return nil

	case "Errorf", "New":
		log.Printf("%s: rewriting the function call with xerrors.%s()", p, functionName)
		pkg.Name = "xerrors"
		v.changes++
		return nil

	case "Cause":
		log.Printf("%s: rewriting the function call with xerrors.%s()", p, functionName)
		pkg.Name = "xerrors"
		fun.Sel.Name = "Unwrap"
		v.changes++
		return nil

	default:
		log.Printf("%s: NOTE: you need to manually rewrite errors.%s()", p, functionName)
		return nil
	}
}
