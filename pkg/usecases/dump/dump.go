package dump

import (
	"context"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

func Do(ctx context.Context, pkgNames ...string) error {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return xerrors.New("error while loading the packages")
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if err := ast.Print(pkg.Fset, file); err != nil {
				p := position(pkg, file)
				return xerrors.Errorf("could not dump file %s: %w", p.Filename, err)
			}
		}
	}
	return nil
}

func position(pkg *packages.Package, file *ast.File) token.Position {
	p := pkg.Fset.Position(file.Pos())
	p.Filename = relative(p.Filename)
	return p
}

func relative(name string) string {
	wd, _ := os.Getwd()
	if wd != "" {
		wd += "/"
	}
	return strings.TrimPrefix(name, wd)
}
