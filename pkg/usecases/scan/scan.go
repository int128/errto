package scan

import (
	"context"
	"log"

	"github.com/int128/migerr/pkg/adaptors/inspector"
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
	ins.FindPackageFunctionCalls(func(call inspector.PackageFunctionCall) {
		log.Printf("call: %+v", call)
	})
	return nil
}
