package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type Dump struct {
	Loader inspector.Loader
}

func (d *Dump) New(ctx context.Context) *cobra.Command {
	c := &cobra.Command{
		Use:   "dump PACKAGE...",
		Short: "Dump AST of packages",
		RunE: func(_ *cobra.Command, args []string) error {
			ins, err := d.Loader.Load(ctx, args...)
			if err != nil {
				return xerrors.Errorf("could not load the packages: %w", err)
			}
			if err := ins.Dump(); err != nil {
				return xerrors.Errorf("could not dump the packages: %w", err)
			}
			return nil
		},
	}
	return c
}
