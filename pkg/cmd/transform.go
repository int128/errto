package cmd

import (
	"context"

	"github.com/int128/transerr/pkg/usecases/transform"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type Transform struct {
	UseCase transform.UseCase
}

func (m *Transform) New(ctx context.Context) *cobra.Command {
	var o struct {
		dryRun bool
	}
	c := &cobra.Command{
		Use:   "transform PACKAGE...",
		Short: "Transform error handling",
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := transform.Config{
				PkgNames: args,
				DryRun:   o.dryRun,
			}
			if err := m.UseCase.Do(ctx, cfg); err != nil {
				return xerrors.Errorf("could not transform the packages: %w", err)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
	return c
}
