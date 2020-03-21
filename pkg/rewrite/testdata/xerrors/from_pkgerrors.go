package main

import (
	"golang.org/x/xerrors"
)

func syntaxSupportedOnlyFromPkgErrors(x int, y string, err error) {
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
