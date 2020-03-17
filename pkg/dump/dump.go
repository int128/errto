package dump

import (
	"context"
	"go/ast"

	"github.com/int128/errto/pkg/astio"
	"golang.org/x/xerrors"
)

func Do(ctx context.Context, pkgNames ...string) error {
	pkgs, err := astio.Load(ctx, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if err := ast.Print(pkg.Fset, file); err != nil {
				p := astio.Position(pkg, file)
				return xerrors.Errorf("could not dump file %s: %w", p.Filename, err)
			}
		}
	}
	return nil
}
