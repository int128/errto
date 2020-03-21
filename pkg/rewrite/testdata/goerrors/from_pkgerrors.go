package main

import (
	"fmt"
)

func syntaxSupportedOnlyFromPkgErrors(x int, y string, err error) {
	// wrap an error without format
	fmt.Errorf("%s: %w", "MESSAGE", err)

	// wrap an error with the stack trace
	fmt.Errorf("%w", err)

	// wrap an error with a message
	fmt.Errorf("%s: %s", "MESSAGE", err)

	// wrap an error with a message
	fmt.Errorf("FORMAT: %s", err)
	fmt.Errorf("FORMAT %d: %s", x, err)
	fmt.Errorf("FORMAT %d, %s: %s", x, y, err)
}
