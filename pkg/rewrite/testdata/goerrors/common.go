package main

import (
	"errors"
	"fmt"
)

func commonSyntax(x int, y string, err error) {
	// create an error
	errors.New("MESSAGE")

	// format an error
	fmt.Errorf("FORMAT")
	fmt.Errorf("FORMAT %d", x)
	fmt.Errorf("FORMAT %d, %s", x, y)

	// wrap an error
	fmt.Errorf("FORMAT: %w", err)
	fmt.Errorf("FORMAT %d: %w", x, err)
	fmt.Errorf("FORMAT %d, %s: %w", x, y, err)

	// unwrap an error
	errors.Unwrap(err)
}
