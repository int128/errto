package inspector

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type Loader struct{}

func (*Loader) Load(ctx context.Context, pkgNames ...string) (*Inspector, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		return nil, xerrors.Errorf("could not load the packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, xerrors.New("error while loading the packages")
	}
	return &Inspector{Pkgs: pkgs}, nil
}

type Inspector struct {
	Pkgs []*packages.Package
}

func (ins *Inspector) Inspect(f func(*packages.Package, ast.Node)) {
	for _, pkg := range ins.Pkgs {
		for _, syntax := range pkg.Syntax {
			ast.Inspect(syntax, func(node ast.Node) bool {
				f(pkg, node)
				return true
			})
		}
	}
}

type PackageFunctionCall struct {
	Position     token.Position
	PackagePath  string
	FunctionName string
	Args         []ast.Expr
}

func (ins *Inspector) FindPackageFunctionCalls(f func(PackageFunctionCall)) {
	ins.Inspect(func(pkg *packages.Package, node ast.Node) {
		switch node := node.(type) {
		case *ast.CallExpr:
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						f(PackageFunctionCall{
							Position:     pkg.Fset.Position(fun.Pos()),
							PackagePath:  o.Imported().Path(),
							FunctionName: fun.Sel.Name,
							Args:         node.Args,
						})
					}
				}
			}
		}
	})
}
