package main

import (
	"context"
	"os"

	"github.com/int128/migerr/pkg/cmd"
)

func main() {
	ctx := context.Background()
	var c cmd.Cmd
	os.Exit(c.Run(ctx, os.Args))
}
