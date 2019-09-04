package testdata

import (
	"golang.org/x/xerrors"
)

// Hello says hello world!
func Hello() error {
	return xerrors.New("hello world")
}
