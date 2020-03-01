package cmd

import (
	"github.com/int128/transerr/pkg/dump"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func newDumpCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "dump PACKAGE...",
		Short: "Dump AST of packages",
		RunE: func(c *cobra.Command, args []string) error {
			if err := dump.Do(c.Context(), args...); err != nil {
				return xerrors.Errorf("could not dump the packages: %w", err)
			}
			return nil
		},
	}
	return c
}
