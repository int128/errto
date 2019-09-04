package cmd

import (
	"context"
	"log"
)

type Cmd struct {
	Scan    Scan
	Migrate Migrate
}

func (c *Cmd) Run(ctx context.Context, osArgs []string) int {
	//TODO: add sub-commands
	if err := c.Migrate.Run(ctx, osArgs[1:]...); err != nil {
		log.Printf("error: %s", err)
		return 1
	}
	return 0
}
