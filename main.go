package main

import (
	"context"
	"log"
	"os"

	"github.com/int128/migerr/pkg/cmd"
)

func init() {
	log.SetFlags(0)
}

func main() {
	ctx := context.Background()
	var c cmd.Cmd
	os.Exit(c.Run(ctx, os.Args))
}
