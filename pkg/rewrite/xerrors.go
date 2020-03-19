package rewrite

import (
	"fmt"
	"go/ast"
	"go/token"
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

func (v *toXerrorsVisitor) PackageFunctionCall(call astio.PackageFunctionCall) error {
	switch call.PackagePath() {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(call)
	case "errors":
		return v.goErrorsFunctionCall(call)
	case "fmt":
		return v.goFmtFunctionCall(call)
	}
	return nil
}

func (v *toXerrorsVisitor) goErrorsFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "New", "Unwrap", "As", "Is":
		replacePackageFunctionCall(call, "xerrors", "")
		v.needImport++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "xerrors"
	v.needImport++
	return nil
}

func (v *toXerrorsVisitor) goFmtFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "Errorf":
		replacePackageFunctionCall(call, "xerrors", "")
		v.needImport++
		return nil
	}
	return nil
}

func (v *toXerrorsVisitor) pkgErrorsFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "Wrapf":
		args := call.Args()
		// append %w to the format arg
		b, ok := args[1].(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("%s: 2nd argument of Wrapf must be a literal but was %T", call.Position, args[1])
		}
		if b.Kind != token.STRING {
			return fmt.Errorf("%s: 2nd argument of Wrapf must be a string but was %s", call.Position, b.Kind)
		}
		b.Value = strings.TrimSuffix(b.Value, `"`) + `: %w"`

		// reorder the args
		var newArgs []ast.Expr
		newArgs = append(newArgs, args[1])
		newArgs = append(newArgs, args[2:]...)
		newArgs = append(newArgs, args[0])
		call.SetArgs(newArgs)

		replacePackageFunctionCall(call, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "Errorf", "New":
		replacePackageFunctionCall(call, "xerrors", "")
		v.needImport++
		return nil

	case "Cause":
		replacePackageFunctionCall(call, "xerrors", "Unwrap")
		v.needImport++
		return nil

	case "Wrap":
		args := call.Args()
		if len(args) != 2 {
			return fmt.Errorf("%s: errors.Wrap expects 2 arguments but has %d arguments", call.Position, len(args))
		}
		call.SetArgs([]ast.Expr{
			&ast.BasicLit{Value: `"%s: %w"`},
			args[1],
			args[0],
		})
		replacePackageFunctionCall(call, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithStack":
		args := call.Args()
		if len(args) != 1 {
			return fmt.Errorf("%s: errors.WithStack expects 1 argument but has %d arguments", call.Position, len(args))
		}
		call.SetArgs([]ast.Expr{
			&ast.BasicLit{Value: `"%w"`},
			args[0],
		})
		replacePackageFunctionCall(call, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithMessage":
		args := call.Args()
		if len(args) != 2 {
			return fmt.Errorf("%s: errors.WithMessage expects 2 arguments but has %d arguments", call.Position, len(args))
		}
		call.SetArgs([]ast.Expr{
			&ast.BasicLit{Value: `"%s: %s"`},
			args[1],
			args[0],
		})
		replacePackageFunctionCall(call, "xerrors", "Errorf")
		v.needImport++
		return nil

	case "WithMessagef":
		args := call.Args()
		// append %s to the format arg
		b, ok := args[1].(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("%s: 2nd argument of WithMessagef must be a literal but %T", call.Position, args[1])
		}
		if b.Kind != token.STRING {
			return fmt.Errorf("%s: 2nd argument of WithMessagef must be a string but %s", call.Position, b.Kind)
		}
		b.Value = strings.TrimSuffix(b.Value, `"`) + `: %s"`

		// reorder the args
		var newArgs []ast.Expr
		newArgs = append(newArgs, args[1])
		newArgs = append(newArgs, args[2:]...)
		newArgs = append(newArgs, args[0])
		call.SetArgs(newArgs)

		replacePackageFunctionCall(call, "xerrors", "Errorf")
		v.needImport++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "xerrors"
	v.needImport++
	return nil
}
