package cmd

import (
	"github.com/int128/transerr/pkg/rewrite"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func newRewriteCmd() *cobra.Command {
	var o struct {
		dryRun bool
	}
	c := &cobra.Command{
		Use:   "rewrite PACKAGE...",
		Short: "Rewrite error handling code",
		RunE: func(c *cobra.Command, args []string) error {
			in := rewrite.Input{
				PkgNames: args,
				DryRun:   o.dryRun,
			}
			if err := rewrite.Do(c.Context(), in); err != nil {
				return xerrors.Errorf("rewrite: %w", err)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
	return c
}
