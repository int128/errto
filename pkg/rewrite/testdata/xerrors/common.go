package main

import (
	"golang.org/x/xerrors"
)

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
}
