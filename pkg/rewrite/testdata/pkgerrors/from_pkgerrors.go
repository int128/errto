package main

import (
	"github.com/pkg/errors"
)

func syntaxSupportedOnlyFromPkgErrors(x int, y string, err error) {
	// wrap an error without format
	errors.Wrap(err, "MESSAGE")

	// wrap an error with the stack trace
	errors.WithStack(err)

	// wrap an error with a message
	errors.WithMessage(err, "MESSAGE")

	// wrap an error with a message
	errors.WithMessagef(err, "FORMAT")
	errors.WithMessagef(err, "FORMAT %d", x)
	errors.WithMessagef(err, "FORMAT %d, %s", x, y)
}
