package cmd

import (
	"fmt"

	"github.com/int128/errto/pkg/dump"
	"github.com/spf13/cobra"
)

func newDumpCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "dump PACKAGE...",
		Short: "Dump AST of packages",
		RunE: func(c *cobra.Command, args []string) error {
			if err := dump.Do(c.Context(), args...); err != nil {
				return fmt.Errorf("could not dump the packages: %w", err)
			}
			return nil
		},
	}
	return c
}
