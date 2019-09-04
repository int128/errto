package migrate

import (
	"context"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"github.com/int128/migerr/pkg/domain/eh"
	"golang.org/x/xerrors"
)

type UseCase struct {
	Loader inspector.Loader
}

func (uc *UseCase) Do(ctx context.Context, pkgNames ...string) error {
	ins, err := uc.Loader.Load(ctx, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	if err := ins.MutatePackageFunctionCalls(eh.PkgErrorsToXerrors); err != nil {
		return xerrors.Errorf("could not mutate the packages: %w", err)
	}
	if err := ins.Print(); err != nil {
		return xerrors.Errorf("could not print the packages: %w", err)
	}
	return nil
}
