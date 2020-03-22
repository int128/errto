package astio

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Visitor interface {
	PackageFunctionCall(call PackageFunctionCall) error
}

type PackageFunctionCall struct {
	Position      token.Position
	Call          *ast.CallExpr
	TargetPkg     *ast.Ident
	TargetPkgName *types.PkgName
	TargetFun     *ast.SelectorExpr
	TypesInfo     *types.Info
}

func (call *PackageFunctionCall) PackagePath() string {
	return call.TargetPkgName.Imported().Path()
}

func (call *PackageFunctionCall) FunctionName() string {
	return call.TargetFun.Sel.Name
}

func (call *PackageFunctionCall) Args() []ast.Expr {
	return call.Call.Args
}

func (call *PackageFunctionCall) SetArgs(args []ast.Expr) {
	call.Call.Args = args
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
						if err := v.PackageFunctionCall(PackageFunctionCall{
							Position:      p,
							Call:          node,
							TargetPkg:     x,
							TargetPkgName: o,
							TargetFun:     fun,
							TypesInfo:     pkg.TypesInfo,
						}); err != nil {
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
