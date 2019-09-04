package scan

import (
	"context"
	"log"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"github.com/int128/migerr/pkg/domain/inst"
	"golang.org/x/xerrors"
)

type Interface interface {
	Do(ctx context.Context, pkgNames ...string) error
}

type UseCase struct {
	Loader inspector.Loader
}

func (uc *UseCase) Do(ctx context.Context, pkgNames ...string) error {
	ins, err := uc.Loader.Load(ctx, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	if err := ins.MutatePackageFunctionCalls(func(m inst.PackageFunctionCallMutator) error {
		log.Printf("(%s).%s", m.PackagePath(), m.FunctionName())
		return nil
	}); err != nil {
		return xerrors.Errorf("could not scan the packages: %w", err)
	}
	return nil
}
