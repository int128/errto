package cmd

import (
	"fmt"

	"github.com/int128/errto/pkg/rewrite"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newRewriteToGoErrorsCmd() *cobra.Command {
	var o rewriteOption
	c := &cobra.Command{
		Use:   "go-errors [flags] PACKAGE...",
		Short: "Rewrite the packages with Go errors (fmt, errors)",
		RunE: func(c *cobra.Command, args []string) error {
			in := rewrite.Input{
				PkgNames: args,
				Target:   rewrite.GoErrors,
				DryRun:   o.dryRun,
			}
			if err := rewrite.Do(c.Context(), in); err != nil {
				return fmt.Errorf("rewrite: %w", err)
			}
			return nil
		},
	}
	o.register(c.Flags())
	return c
}

func newRewriteToXerrorsCmd() *cobra.Command {
	var o rewriteOption
	c := &cobra.Command{
		Use:   "xerrors [flags] PACKAGE...",
		Short: "Rewrite the packages with golang.org/x/xerrors",
		RunE: func(c *cobra.Command, args []string) error {
			in := rewrite.Input{
				PkgNames: args,
				Target:   rewrite.Xerrors,
				DryRun:   o.dryRun,
			}
			if err := rewrite.Do(c.Context(), in); err != nil {
				return fmt.Errorf("rewrite: %w", err)
			}
			return nil
		},
	}
	o.register(c.Flags())
	return c
}

func newRewriteToPkgErrorsCmd() *cobra.Command {
	var o rewriteOption
	c := &cobra.Command{
		Use:   "pkg-errors [flags] PACKAGE...",
		Short: "Rewrite the packages with github.com/pkg/errors",
		RunE: func(c *cobra.Command, args []string) error {
			in := rewrite.Input{
				PkgNames: args,
				Target:   rewrite.PkgErrors,
				DryRun:   o.dryRun,
			}
			if err := rewrite.Do(c.Context(), in); err != nil {
				return fmt.Errorf("rewrite: %w", err)
			}
			return nil
		},
	}
	o.register(c.Flags())
	return c
}

type rewriteOption struct {
	dryRun bool
}

func (o *rewriteOption) register(f *pflag.FlagSet) {
	f.BoolVar(&o.dryRun, "dry-run", false, "Do not write files actually")
}
