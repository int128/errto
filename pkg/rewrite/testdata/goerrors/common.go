package main

import (
	"errors"
	"fmt"
)

type SomeError struct{}

func (err SomeError) Error() string {
	return "hello"
}

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

	// cast an error
	var targetErr SomeError
	errors.As(err, &targetErr)

	// test an error
	errors.Is(err, &targetErr)
}
