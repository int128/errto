package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/usecases/migrate"
	"golang.org/x/xerrors"
)

type Migrate struct {
	UseCase migrate.UseCase
}

func (c *Migrate) Run(ctx context.Context, pkgNames ...string) error {
	if err := c.UseCase.Do(ctx, pkgNames...); err != nil {
		return xerrors.Errorf("could not migrate the packages: %w", err)
	}
	return nil
}
