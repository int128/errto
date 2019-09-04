package testdata

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var msg = "foo"

// Hello says hello world!
func Hello() error {
	if _, err := fmt.Fprintf(os.Stderr, "hello world"); err != nil {
		return errors.Wrapf(err, "error %s", msg)
	}
	if _, err := fmt.Fprintf(os.Stderr, "hello world"); err != nil {
		return errors.Errorf("error %s", msg)
	}
	return nil
}
