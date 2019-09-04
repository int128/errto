package inspector

import (
	"context"
	"go/ast"
	"go/printer"
	"go/types"
	"os"

	"github.com/int128/migerr/pkg/domain/inst"
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

func (ins *Inspector) Inspect(f func(*packages.Package, *ast.File, ast.Node) error) error {
	var lastErr error
	for _, pkg := range ins.Pkgs {
		for _, syntax := range pkg.Syntax {
			ast.Inspect(syntax, func(node ast.Node) bool {
				if err := f(pkg, syntax, node); err != nil {
					lastErr = xerrors.Errorf("inspection error: %w", err)
					return false
				}
				return true
			})
		}
	}
	return lastErr
}

func (ins *Inspector) Print() error {
	for _, pkg := range ins.Pkgs {
		for _, syntax := range pkg.Syntax {
			if err := printer.Fprint(os.Stdout, pkg.Fset, syntax); err != nil {
				return xerrors.Errorf("could not print %s: %w", pkg, err)
			}
		}
	}
	return nil
}

func (ins *Inspector) Dump() error {
	for _, pkg := range ins.Pkgs {
		for _, syntax := range pkg.Syntax {
			if err := ast.Print(pkg.Fset, syntax); err != nil {
				return xerrors.Errorf("could not dump %s: %w", pkg, err)
			}
		}
	}
	return nil
}

func (ins *Inspector) MutatePackageFunctionCalls(f func(inst.PackageFunctionCallMutator) error) error {
	if err := ins.Inspect(func(pkg *packages.Package, file *ast.File, node ast.Node) error {
		switch node := node.(type) {
		case *ast.CallExpr:
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						m := &PackageFunctionCallMutator{
							call:    node,
							pkgName: o,
							f:       fun,
							x:       x,
						}
						if err := f(m); err != nil {
							return xerrors.Errorf("mutation error: %w", err)
						}
					}
				}
			}
		}
		return nil
	}); err != nil {
		return xerrors.Errorf("error while mutation: %w", err)
	}
	return nil
}

type PackageFunctionCallMutator struct {
	call    *ast.CallExpr
	pkgName *types.PkgName
	f       *ast.SelectorExpr
	x       *ast.Ident
}

func (m *PackageFunctionCallMutator) PackagePath() string {
	return m.pkgName.Imported().Path()
}

func (m *PackageFunctionCallMutator) PackageName() string {
	return m.x.Name
}

func (m *PackageFunctionCallMutator) FunctionName() string {
	return m.f.Sel.Name
}

func (m *PackageFunctionCallMutator) Args() []ast.Expr {
	return m.call.Args
}

func (m *PackageFunctionCallMutator) SetPackageName(pkgName string) {
	m.x.Name = pkgName
}

func (m *PackageFunctionCallMutator) SetFunctionName(name string) {
	m.f.Sel.Name = name
}

func (m *PackageFunctionCallMutator) SetArgs(args []ast.Expr) {
	m.call.Args = args
}
