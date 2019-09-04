package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/usecases/scan"
	"golang.org/x/xerrors"
)

type Scan struct {
	UseCase scan.UseCase
}

func (s *Scan) Run(ctx context.Context, pkgNames ...string) error {
	if err := s.UseCase.Do(ctx, pkgNames...); err != nil {
		return xerrors.Errorf("could not scan the packages: %w", err)
	}
	return nil
}
