package astio

import (
	"context"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

func Load(ctx context.Context, pkgNames ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		return nil, xerrors.Errorf("load error: %w", err)
	}
	if n := packages.PrintErrors(pkgs); n > 0 {
		return nil, xerrors.Errorf("load error")
	}
	return pkgs, nil
}
