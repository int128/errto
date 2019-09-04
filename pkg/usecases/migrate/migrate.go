package migrate

import (
	"context"
	"strings"

	"github.com/int128/migerr/pkg/adaptors/inspector"
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
	if err := ins.Dump(); err != nil {
		return xerrors.Errorf("could not print the packages: %w", err)
	}
	ins.MutatePackageFunctionCalls(func(m inspector.PackageFunctionCallMutator) {
		// pkg/errors.Wrapf() -> xerrors.Errorf()
		if m.Target() == (inspector.PackageFunctionCall{PackagePath: "github.com/pkg/errors", FunctionName: "Wrapf"}) {
			m.SetTarget("xerrors", "Errorf")

			format := m.Args()[1].StringLiteral()
			args := make([]*inspector.FunctionCallArg, 0)
			args = append(args, inspector.NewFunctionCallArgStringLiteral(strings.TrimRight(format, `"`)+`: %w"`))
			args = append(args, m.Args()[2:]...)
			args = append(args, m.Args()[0])
			m.SetArgs(args)
			return
		}
		// pkg/errors.Errorf() -> xerrors.Errorf()
		if m.Target() == (inspector.PackageFunctionCall{PackagePath: "github.com/pkg/errors", FunctionName: "Errorf"}) {
			m.SetTarget("xerrors", "Errorf")
			return
		}
	})
	if err := ins.Print(); err != nil {
		return xerrors.Errorf("could not print the packages: %w", err)
	}
	return nil
}
