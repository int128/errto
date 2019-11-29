package dump

import (
	"context"

	"github.com/google/wire"
	"github.com/int128/transerr/pkg/adaptors/inspector"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*UseCase)),
	wire.Struct(new(UseCase), "*"),
)

type Interface interface {
	Do(ctx context.Context, pkgNames ...string) error
}

type UseCase struct {
	Inspector inspector.Interface
}

func (uc *UseCase) Do(ctx context.Context, pkgNames ...string) error {
	pkgs, err := uc.Inspector.Load(ctx, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if err := uc.Inspector.Dump(pkg, file); err != nil {
				return xerrors.Errorf("could not dump the file %s: %w", file, err)
			}
		}
	}
	return nil
}
