package testdata

import (
	"github.com/pkg/errors"
)

var msg = "foo"

// Hello says hello world!
func Hello() error {
	err := errors.New("message")
	_ = errors.Wrapf(err, "message %s", msg)
	_ = errors.Errorf("message %s", msg)
	return nil
}
