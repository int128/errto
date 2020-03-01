package cmd

import (
	"github.com/int128/transerr/pkg/transform"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func newTransformCmd() *cobra.Command {
	var o struct {
		dryRun bool
	}
	c := &cobra.Command{
		Use:   "transform PACKAGE...",
		Short: "Transform error handling",
		RunE: func(c *cobra.Command, args []string) error {
			in := transform.Input{
				PkgNames: args,
				DryRun:   o.dryRun,
			}
			if err := transform.Do(c.Context(), in); err != nil {
				return xerrors.Errorf("could not transform the packages: %w", err)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
	return c
}
