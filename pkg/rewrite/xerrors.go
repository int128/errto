package rewrite

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/int128/errto/pkg/astio"
	"github.com/int128/errto/pkg/log"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type toXerrors struct{}

func (t *toXerrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toXerrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, fmt.Errorf("could not inspect the file: %w", err)
	}
	if v.needImport == 0 {
		return 0, nil
	}
	n := t.replaceImports(pkg, file)
	return v.needImport + n, nil
}

func (*toXerrors) replaceImports(pkg *packages.Package, file *ast.File) int {
	var n int
	if astutil.AddImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("%s: + import %s", astio.Filename(pkg, file), xerrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("%s: - import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, "errors") {
		n++
		log.Printf("%s: - import %s", astio.Filename(pkg, file), "errors")
	}
	if !astutil.UsesImport(file, "fmt") {
		if astutil.DeleteImport(pkg.Fset, file, "fmt") {
			n++
			log.Printf("%s: - import %s", astio.Filename(pkg, file), "fmt")
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
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "")
		v.needImport++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "xerrors"
	v.needImport++
	return nil
}

func (v *toXerrorsVisitor) goFmtFunctionCall(p token.Position, _ *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "")
		v.needImport++
		return nil
	}
	return nil
}

func (v *toXerrorsVisitor) pkgErrorsFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		// append %w to the format arg
		b, ok := call.Args[1].(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("%s: 2nd argument of Wrapf must be a literal but was %T", p, call.Args[1])
		}
		if b.Kind != token.STRING {
			return fmt.Errorf("%s: 2nd argument of Wrapf must be a string but was %s", p, b.Kind)
		}
		b.Value = strings.TrimSuffix(b.Value, `"`) + `: %w"`

		// reorder the args
		var args []ast.Expr
		args = append(args, call.Args[1])
		args = append(args, call.Args[2:]...)
		args = append(args, call.Args[0])
		call.Args = args

		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "Errorf", "New":
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "")
		v.needImport++
		return nil

	case "Cause":
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Unwrap")
		v.needImport++
		return nil

	case "Wrap":
		if len(call.Args) != 2 {
			return fmt.Errorf("%s: errors.Wrap expects 2 arguments but has %d arguments", p, len(call.Args))
		}
		call.Args = []ast.Expr{
			&ast.BasicLit{Value: `"%s: %w"`},
			call.Args[1],
			call.Args[0],
		}
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithStack":
		if len(call.Args) != 1 {
			return fmt.Errorf("%s: errors.WithStack expects 1 argument but has %d arguments", p, len(call.Args))
		}
		call.Args = []ast.Expr{
			&ast.BasicLit{Value: `"%w"`},
			call.Args[0],
		}
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithMessage":
		if len(call.Args) != 2 {
			return fmt.Errorf("%s: errors.WithMessage expects 2 arguments but has %d arguments", p, len(call.Args))
		}
		call.Args = []ast.Expr{
			&ast.BasicLit{Value: `"%s: %s"`},
			call.Args[1],
			call.Args[0],
		}
		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithMessagef":
		// append %s to the format arg
		b, ok := call.Args[1].(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("%s: 2nd argument of WithMessagef must be a literal but %T", p, call.Args[1])
		}
		if b.Kind != token.STRING {
			return fmt.Errorf("%s: 2nd argument of WithMessagef must be a string but %s", p, b.Kind)
		}
		b.Value = strings.TrimSuffix(b.Value, `"`) + `: %s"`

		// reorder the args
		var args []ast.Expr
		args = append(args, call.Args[1])
		args = append(args, call.Args[2:]...)
		args = append(args, call.Args[0])
		call.Args = args

		replacePackageFunctionCall(p, pkg, fun, "xerrors", "Errorf")
		v.needImport++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "xerrors"
	v.needImport++
	return nil
}
