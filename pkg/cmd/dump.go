package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/usecases/dump"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type Dump struct {
	UseCase dump.UseCase
}

func (d *Dump) New(ctx context.Context) *cobra.Command {
	c := &cobra.Command{
		Use:   "dump PACKAGE...",
		Short: "Dump AST of packages",
		RunE: func(_ *cobra.Command, args []string) error {
			if err := d.UseCase.Do(ctx, args...); err != nil {
				return xerrors.Errorf("could not dump the packages: %w", err)
			}
			return nil
		},
	}
	return c
}
