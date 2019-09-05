package inspector

import (
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type Inspector struct{}

func (*Inspector) Load(ctx context.Context, pkgNames ...string) ([]*packages.Package, error) {
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
	return pkgs, nil
}

func (ins *Inspector) Print(pkg *packages.Package, file *ast.File) error {
	if err := printer.Fprint(os.Stdout, pkg.Fset, file); err != nil {
		return xerrors.Errorf("could not print the file: %w", err)
	}
	return nil
}

func (ins *Inspector) Dump(pkg *packages.Package, file *ast.File) error {
	if err := ast.Print(pkg.Fset, file); err != nil {
		return xerrors.Errorf("could not dump the file: %w", err)
	}
	return nil
}

type Visitor interface {
	Import(Import) error
	PackageFunctionCall(PackageFunctionCall) error
}

func (ins *Inspector) Inspect(pkg *packages.Package, file *ast.File, v Visitor) error {
	var lastErr error
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			switch node.Tok {
			case token.IMPORT:
				for _, spec := range node.Specs {
					switch spec := spec.(type) {
					case *ast.ImportSpec:
						imp := Import{
							position: pkg.Fset.Position(spec.Pos()),
							spec:     spec,
						}
						if err := v.Import(imp); err != nil {
							lastErr = err
							return false
						}
					default:
						lastErr = xerrors.Errorf("spec wants *ast.ImportSpec but was %T", spec)
						return false
					}
				}
			}
		case *ast.CallExpr:
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						c := PackageFunctionCall{
							pkgPath:  o.Imported().Path(),
							position: pkg.Fset.Position(node.Pos()),
							call:     node,
							f:        fun,
							x:        x,
						}
						if err := v.PackageFunctionCall(c); err != nil {
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

type Import struct {
	position token.Position
	spec     *ast.ImportSpec
}

func (imp *Import) PackagePath() string {
	return strings.Trim(imp.spec.Path.Value, `"`)
}

func (imp *Import) SetPackagePath(path string, name string) {
	imp.spec.Path.Value = fmt.Sprintf(`"%s"`, path)
	if name != "" {
		imp.spec.Name = &ast.Ident{Name: name}
	} else {
		imp.spec.Name = nil
	}
}

func (imp *Import) Position() token.Position {
	return imp.position
}

type PackageFunctionCall struct {
	pkgPath  string
	position token.Position

	call *ast.CallExpr
	f    *ast.SelectorExpr
	x    *ast.Ident
}

func (c *PackageFunctionCall) PackagePath() string {
	return c.pkgPath
}

func (c *PackageFunctionCall) PackageName() string {
	return c.x.Name
}

func (c *PackageFunctionCall) FunctionName() string {
	return c.f.Sel.Name
}

func (c *PackageFunctionCall) Args() []ast.Expr {
	return c.call.Args
}

func (c *PackageFunctionCall) Position() token.Position {
	return c.position
}

func (c *PackageFunctionCall) SetPackageName(pkgName string) {
	c.x.Name = pkgName
}

func (c *PackageFunctionCall) SetFunctionName(name string) {
	c.f.Sel.Name = name
}

func (c *PackageFunctionCall) SetArgs(args []ast.Expr) {
	c.call.Args = args
}
