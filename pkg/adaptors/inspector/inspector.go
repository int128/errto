package inspector

import (
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"os"

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

type PackageFunctionCall struct {
	PackagePath  string
	FunctionName string
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
						call := PackageFunctionCall{
							//Position:     pkg.Fset.Position(fun.Pos()),
							PackagePath:  o.Imported().Path(),
							FunctionName: fun.Sel.Name,
							//Args:         node.Args,
						}
						f(call)
					}
				}
			}
		}
	})
}

//TODO
func (ins *Inspector) MutateImport(fromPath, toPath string) {
	ins.Inspect(func(pkg *packages.Package, node ast.Node) {
		switch node := node.(type) {
		case *ast.GenDecl:
			if node.Tok.String() == "import" {
				for _, spec := range node.Specs {
					if imp, ok := spec.(*ast.ImportSpec); ok {
						if imp.Path.Value == fmt.Sprintf(`"%s"`, fromPath) {
							imp.Path.Value = fmt.Sprintf(`"%s"`, toPath)
						}
					}
				}
			}
		}
	})
}

//TODO
func (ins *Inspector) MutatePackageFunctionCalls(f func(PackageFunctionCallMutator)) {
	ins.Inspect(func(pkg *packages.Package, node ast.Node) {
		switch node := node.(type) {
		case *ast.CallExpr:
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						m := PackageFunctionCallMutator{
							c: PackageFunctionCall{
								PackagePath:  o.Imported().Path(),
								FunctionName: fun.Sel.Name,
							},
							call: node,
							x:    x,
							sel:  fun.Sel,
						}
						f(m)
					}
				}
			}
		}
	})
}

type PackageFunctionCallMutator struct {
	c    PackageFunctionCall
	call *ast.CallExpr
	x    *ast.Ident
	sel  *ast.Ident
}

func (m *PackageFunctionCallMutator) Target() PackageFunctionCall {
	return m.c
}

func (m *PackageFunctionCallMutator) SetTarget(pkgName, functionName string) {
	m.x.Name = pkgName
	m.sel.Name = functionName
}

type FunctionCallArg struct {
	expr ast.Expr
}

func (a *FunctionCallArg) StringLiteral() string {
	if l, ok := a.expr.(*ast.BasicLit); ok {
		return l.Value
	}
	return ""
}

func NewFunctionCallArgStringLiteral(s string) *FunctionCallArg {
	return &FunctionCallArg{&ast.BasicLit{
		Kind:  token.STRING,
		Value: s,
	}}
}

func (m *PackageFunctionCallMutator) Args() []*FunctionCallArg {
	args := make([]*FunctionCallArg, len(m.call.Args))
	for i, arg := range m.call.Args {
		args[i] = &FunctionCallArg{arg}
	}
	return args
}

func (m *PackageFunctionCallMutator) SetArgs(args []*FunctionCallArg) {
	m.call.Args = make([]ast.Expr, len(args))
	for i, arg := range args {
		m.call.Args[i] = arg.expr
	}
}
