package main

import (
	"golang.org/x/xerrors"
)

type SomeError struct{}

func (err SomeError) Error() string {
	return "hello"
}

func commonSyntax(x int, y string, err error) {
	// create an error
	xerrors.New("MESSAGE")

	// format an error
	xerrors.Errorf("FORMAT")
	xerrors.Errorf("FORMAT %d", x)
	xerrors.Errorf("FORMAT %d, %s", x, y)

	// wrap an error
	xerrors.Errorf("FORMAT: %w", err)
	xerrors.Errorf("FORMAT %d: %w", x, err)
	xerrors.Errorf("FORMAT %d, %s: %w", x, y, err)

	// unwrap an error
	xerrors.Unwrap(err)

	// cast an error
	var targetErr SomeError
	xerrors.As(err, &targetErr)

	// test an error
	xerrors.Is(err, &targetErr)

	// wrap an error without format
	xerrors.Errorf("%s: %w", "MESSAGE", err)

	// wrap an error with the stack trace
	xerrors.Errorf("%w", err)

	// wrap an error with a message
	xerrors.Errorf("%s: %s", "MESSAGE", err)

	// wrap an error with a message
	xerrors.Errorf("FORMAT: %s", err)
	xerrors.Errorf("FORMAT %d: %s", x, err)
	xerrors.Errorf("FORMAT %d, %s: %s", x, y, err)
}
