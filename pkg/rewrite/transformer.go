package rewrite

import (
	"go/ast"

	"github.com/int128/errto/pkg/astio"
	"github.com/int128/errto/pkg/log"
	"golang.org/x/tools/go/packages"
)

type Transformer interface {
	Transform(pkg *packages.Package, file *ast.File) (int, error)
}

func newTransformer(m Method) Transformer {
	switch m {
	case Xerrors:
		return &toXerrors{}
	case GoErrors:
		return &toGoErrors{}
	case PkgErrors:
		return &toPkgErrors{}
	}
	return nil
}

func replacePackageFunctionCall(call astio.PackageFunctionCall, newPkgName, newFunName string) {
	if newFunName == "" {
		newFunName = call.FunctionName()
	}
	log.Printf("%s: %s.%s() -> %s.%s()", call.Position, call.TargetPkg.Name, call.FunctionName(), newPkgName, newFunName)
	call.TargetPkg.Name = newPkgName
	call.TargetFun.Sel.Name = newFunName
}
