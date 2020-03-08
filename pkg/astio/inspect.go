package astio

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Visitor interface {
	PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, resolvedPkgName *types.PkgName, fun *ast.SelectorExpr) error
}

func Inspect(pkg *packages.Package, file *ast.File, v Visitor) error {
	var lastErr error
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			p := Position(pkg, node)
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						if err := v.PackageFunctionCall(p, node, x, o, fun); err != nil {
							lastErr = err
							return false
						}
					}
				}
			}
		}
		return true
	})
	return lastErr
}
