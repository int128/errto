package main

import (
	"github.com/pkg/errors"
)

type SomeError struct{}

func (err SomeError) Error() string {
	return "hello"
}

func commonSyntax(x int, y string, err error) {
	// create an error
	errors.New("MESSAGE")

	// format an error
	errors.Errorf("FORMAT")
	errors.Errorf("FORMAT %d", x)
	errors.Errorf("FORMAT %d, %s", x, y)

	// wrap an error
	errors.Wrapf(err, "FORMAT")
	errors.Wrapf(err, "FORMAT %d", x)
	errors.Wrapf(err, "FORMAT %d, %s", x, y)

	// unwrap an error
	errors.Unwrap(err)

	// cast an error
	var targetErr SomeError
	errors.As(err, &targetErr)

	// test an error
	errors.Is(err, &targetErr)

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
