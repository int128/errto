package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/usecases/migrate"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type Migrate struct {
	UseCase migrate.UseCase
}

func (m *Migrate) New(ctx context.Context) *cobra.Command {
	var o struct {
		dryRun bool
	}
	c := &cobra.Command{
		Use:   "migrate PACKAGE...",
		Short: "Migrate error handling",
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := migrate.Config{
				PkgNames: args,
				DryRun:   o.dryRun,
			}
			if err := m.UseCase.Do(ctx, cfg); err != nil {
				return xerrors.Errorf("could not migrate the packages: %w", err)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
	return c
}
