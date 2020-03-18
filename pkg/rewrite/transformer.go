package rewrite

import (
	"go/ast"
	"go/token"

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

func replacePackageFunctionCall(p token.Position, pkg *ast.Ident, fun *ast.SelectorExpr, newPkgName, newFunName string) {
	if newFunName == "" {
		newFunName = fun.Sel.Name
	}
	log.Printf("rewrite: %s: %s.%s() -> %s.%s()", p, pkg.Name, fun.Sel.Name, newPkgName, newFunName)
	pkg.Name = newPkgName
	fun.Sel.Name = newFunName
}
