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

type toGoErrors struct{}

func (t *toGoErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toGoErrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, fmt.Errorf("could not inspect the file: %w", err)
	}
	if v.needImportFmt == 0 && v.needImportErrors == 0 {
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
			log.Printf("%s: + import %s", astio.Filename(pkg, file), "fmt")
		}
	}
	if needImportErrors > 0 {
		if astutil.AddImport(pkg.Fset, file, "errors") {
			n++
			log.Printf("%s: + import %s", astio.Filename(pkg, file), "errors")
		}
	}
	if astutil.DeleteImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("%s: - import %s", astio.Filename(pkg, file), xerrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("%s: - import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
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

func (v *toGoErrorsVisitor) PackageFunctionCall(call astio.PackageFunctionCall) error {
	switch call.PackagePath() {
	case pkgErrorsImportPath:
		return v.pkgErrorsFunctionCall(call)
	case xerrorsImportPath:
		return v.xerrorsFunctionCall(call)
	}
	return nil
}

func (v *toGoErrorsVisitor) pkgErrorsFunctionCall(call astio.PackageFunctionCall) error {
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

		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "Errorf":
		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "New":
		replacePackageFunctionCall(call, "errors", "")
		v.needImportErrors++
		return nil

	case "Cause":
		replacePackageFunctionCall(call, "errors", "Unwrap")
		v.needImportErrors++
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
		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
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
		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
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
		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
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

		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "errors"
	v.needImportErrors++
	return nil
}

func (v *toGoErrorsVisitor) xerrorsFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "Errorf":
		replacePackageFunctionCall(call, "fmt", "Errorf")
		v.needImportFmt++
		return nil

	case "New", "Unwrap", "As", "Is":
		replacePackageFunctionCall(call, "errors", "")
		v.needImportErrors++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "errors"
	v.needImportErrors++
	return nil
}
