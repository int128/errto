//+build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/transerr/pkg/adaptors/inspector"
	"github.com/int128/transerr/pkg/cmd"
	"github.com/int128/transerr/pkg/usecases/dump"
	"github.com/int128/transerr/pkg/usecases/transform"
)

func NewCmd() cmd.Interface {
	wire.Build(
		// adaptors
		cmd.Set,
		inspector.Set,

		// use-cases
		dump.Set,
		transform.Set,
	)
	return nil
}
