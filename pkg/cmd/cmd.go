package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

type Cmd struct {
	Transform Transform
	Dump      Dump
}

func (c *Cmd) Run(ctx context.Context, osArgs []string) int {
	root := &cobra.Command{
		Use: "transerr",
	}
	root.SilenceErrors = true
	root.SilenceUsage = true

	root.AddCommand(
		c.Transform.New(ctx),
		c.Dump.New(ctx),
	)

	if err := root.Execute(); err != nil {
		log.Printf("error: %s", err)
		return 1
	}
	return 0
}
