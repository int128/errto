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

type toPkgErrors struct{}

func (t *toPkgErrors) Transform(pkg *packages.Package, file *ast.File) (int, error) {
	var v toPkgErrorsVisitor
	if err := astio.Inspect(pkg, file, &v); err != nil {
		return 0, fmt.Errorf("could not inspect the file: %w", err)
	}
	if v.needImport == 0 {
		return 0, nil
	}
	n := t.replaceImports(pkg, file)
	return v.needImport + n, nil
}

func (*toPkgErrors) replaceImports(pkg *packages.Package, file *ast.File) int {
	var n int
	if astutil.AddImport(pkg.Fset, file, pkgErrorsImportPath) {
		n++
		log.Printf("%s: + import %s", astio.Filename(pkg, file), pkgErrorsImportPath)
	}
	if astutil.DeleteImport(pkg.Fset, file, xerrorsImportPath) {
		n++
		log.Printf("%s: - import %s", astio.Filename(pkg, file), xerrorsImportPath)
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
	ast.SortImports(pkg.Fset, file)
	return n
}

type toPkgErrorsVisitor struct {
	needImport int
}

func (v *toPkgErrorsVisitor) PackageFunctionCall(call astio.PackageFunctionCall) error {
	switch call.PackagePath() {
	case xerrorsImportPath:
		return v.xerrorsFunctionCall(call)
	case "errors":
		return v.goErrorsFunctionCall(call)
	case "fmt":
		return v.goFmtFunctionCall(call)
	}
	return nil
}

func (v *toPkgErrorsVisitor) goErrorsFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "New", "Unwrap", "As", "Is":
		replacePackageFunctionCall(call, "errors", "")
		v.needImport++
		return nil
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "errors"
	v.needImport++
	return nil
}

func (v *toPkgErrorsVisitor) goFmtFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "Errorf":
		v.needImport++
		return replaceErrorfWithPkgErrors(call)
	}
	return nil
}

func (v *toPkgErrorsVisitor) xerrorsFunctionCall(call astio.PackageFunctionCall) error {
	switch call.FunctionName() {
	case "New", "Unwrap", "As", "Is":
		replacePackageFunctionCall(call, "errors", "")
		v.needImport++
		return nil

	case "Errorf":
		v.needImport++
		return replaceErrorfWithPkgErrors(call)
	}

	log.Printf("%s: NOTE: you need to manually rewrite %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName())
	call.TargetPkg.Name = "errors"
	v.needImport++
	return nil
}

// replaceErrorfWithPkgErrors rewrites the Errorf function call
// from fmt.Errorf() or xerrors.Errorf() to pkg/errors.Errorf().
// If the Errorf wraps an error, it rewrites to pkg/errors.Wrapf().
func replaceErrorfWithPkgErrors(call astio.PackageFunctionCall) error {
	args := call.Args()
	if len(args) < 2 {
		replacePackageFunctionCall(call, "errors", "Errorf")
		return nil
	}

	// if the last argument is not an error, just rewrite the function call
	lastArgType := call.TypesInfo.TypeOf(args[len(args)-1])
	if lastArgType == nil || lastArgType.String() != "error" {
		replacePackageFunctionCall(call, "errors", "Errorf")
		return nil
	}

	// rewrite to pkg/errors specific functions if the format argument matched
	firstArg, ok := args[0].(*ast.BasicLit)
	if !ok {
		return fmt.Errorf("%s: 1st argument of Errorf must be a literal but %T", call.Position, args[0])
	}
	if firstArg.Kind != token.STRING {
		return fmt.Errorf("%s: 1st argument of Errorf must be a string but %s", call.Position, firstArg.Kind)
	}
	if firstArg.Value == `"%s: %w"` {
		call.SetArgs([]ast.Expr{args[2], args[1]})
		replacePackageFunctionCall(call, "errors", "Wrap")
		return nil
	}
	if firstArg.Value == `"%w"` {
		call.SetArgs([]ast.Expr{args[1]})
		replacePackageFunctionCall(call, "errors", "WithStack")
		return nil
	}
	if firstArg.Value == `"%s: %s"` {
		call.SetArgs([]ast.Expr{args[2], args[1]})
		replacePackageFunctionCall(call, "errors", "WithMessage")
		return nil
	}
	if strings.HasSuffix(firstArg.Value, `: %s"`) {
		firstArg.Value = strings.TrimSuffix(firstArg.Value, `: %s"`) + `"`
		var newArgs []ast.Expr
		newArgs = append(newArgs, args[len(args)-1])
		newArgs = append(newArgs, args[0])
		if len(args) > 2 {
			newArgs = append(newArgs, args[1:len(args)-1]...)
		}
		call.SetArgs(newArgs)
		replacePackageFunctionCall(call, "errors", "WithMessagef")
		return nil
	}
	if strings.HasSuffix(firstArg.Value, `: %w"`) {
		firstArg.Value = strings.TrimSuffix(firstArg.Value, `: %w"`) + `"`
		var newArgs []ast.Expr
		newArgs = append(newArgs, args[len(args)-1])
		newArgs = append(newArgs, args[0])
		if len(args) > 2 {
			newArgs = append(newArgs, args[1:len(args)-1]...)
		}
		call.SetArgs(newArgs)
		replacePackageFunctionCall(call, "errors", "Wrapf")
		return nil
	}

	replacePackageFunctionCall(call, "errors", "Errorf")
	return nil
}
