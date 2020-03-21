package main

import (
	"github.com/pkg/errors"
)

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
	errors.Cause(err)
}
