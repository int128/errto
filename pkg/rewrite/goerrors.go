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

type toGoErrors struct{}

func (t *toGoErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toGoErrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, fmt.Errorf("could not inspect the file: %w", err)
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

		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "Errorf":
		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "New":
		replacePackageFunctionCall(p, pkg, fun, "errors", "")
		v.needImportErrors++
		return nil

	case "Cause":
		replacePackageFunctionCall(p, pkg, fun, "errors", "Unwrap")
		v.needImportErrors++
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
		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "WithStack":
		if len(call.Args) != 1 {
			return fmt.Errorf("%s: errors.WithStack expects 1 argument but has %d arguments", p, len(call.Args))
		}
		call.Args = []ast.Expr{
			&ast.BasicLit{Value: `"%w"`},
			call.Args[0],
		}
		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
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
		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
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

		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "errors"
	v.needImportErrors++
	return nil
}

func (v *toGoErrorsVisitor) xerrorsFunctionCall(p token.Position, pkg *ast.Ident, fun *ast.SelectorExpr) error {
	functionName := fun.Sel.Name
	switch functionName {
	case "Errorf":
		replacePackageFunctionCall(p, pkg, fun, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "New", "Unwrap", "As", "Is":
		replacePackageFunctionCall(p, pkg, fun, "errors", "")
		v.needImportErrors++
		return nil
	}

	log.Printf("rewrite: %s: NOTE: you need to manually rewrite %s.%s()", p, pkg.Name, functionName)
	pkg.Name = "errors"
	v.needImportErrors++
	return nil
}
