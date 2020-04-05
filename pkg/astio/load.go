package astio

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"
)

func Load(ctx context.Context, pkgNames ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests:   true,
	}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}
	if n := packages.PrintErrors(pkgs); n > 0 {
		return nil, fmt.Errorf("load error")
	}
	return pkgs, nil
}
