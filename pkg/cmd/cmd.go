package cmd

import (
	"context"
	"log"

	"github.com/google/wire"
	"github.com/spf13/cobra"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Cmd)),
	wire.Struct(new(Cmd), "*"),
	wire.Struct(new(Transform), "*"),
	wire.Struct(new(Dump), "*"),
)

type Interface interface {
	Run(ctx context.Context, osArgs []string) int
}

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
