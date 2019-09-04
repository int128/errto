package cmd

import (
	"context"
	"log"
)

type Cmd struct {
	Scan Scan
}

func (c *Cmd) Run(ctx context.Context, osArgs []string) int {
	if err := c.Scan.Run(ctx, osArgs[1:]...); err != nil {
		log.Printf("error: %s", err)
		return 1
	}
	return 0
}
