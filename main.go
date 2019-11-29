package main

import (
	"context"
	"log"
	"os"

	"github.com/int128/transerr/pkg/di"
)

func init() {
	log.SetFlags(0)
}

func main() {
	os.Exit(di.NewCmd().Run(context.Background(), os.Args))
}
