package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

func Run(ctx context.Context, osArgs []string) int {
	root := &cobra.Command{
		Use: "errto",
	}

	root.AddCommand(
		newRewriteCmd(),
		newDumpCmd(),
	)

	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs(osArgs[1:])
	if err := root.ExecuteContext(ctx); err != nil {
		log.Printf("error: %s", err)
		return 1
	}
	return 0
}
