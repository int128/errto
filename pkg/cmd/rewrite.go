package cmd

import (
	"github.com/int128/errto/pkg/rewrite"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func newRewriteCmd() *cobra.Command {
	var o rewriteOption
	c := &cobra.Command{
		Use:   "rewrite [flags] --to=METHOD PACKAGE...",
		Short: "Rewrite error handling code",
		RunE: func(c *cobra.Command, args []string) error {
			to, err := o.parseTo()
			if err != nil {
				return xerrors.Errorf("rewrite: %w", err)
			}
			in := rewrite.Input{
				PkgNames: args,
				Target:   to,
				DryRun:   o.dryRun,
			}
			if err := rewrite.Do(c.Context(), in); err != nil {
				return xerrors.Errorf("rewrite: %w", err)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
	c.Flags().StringVar(&o.to, "to", "", "Target error handling method (go-errors|xerrors|pkg-errors)")
	return c
}

type rewriteOption struct {
	dryRun bool
	to     string
}

func (o *rewriteOption) parseTo() (rewrite.Method, error) {
	switch o.to {
	case "":
		return 0, xerrors.New("you need to set --to flag")
	case "go-errors":
		return rewrite.GoErrors, nil
	case "xerrors":
		return rewrite.Xerrors, nil
	case "pkg-errors":
		return rewrite.PkgErrors, nil
	}
	return 0, xerrors.Errorf("unknown package %s", o.to)
}
