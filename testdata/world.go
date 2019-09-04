package testdata

import "errors"

// World says hello world!
func World() (string, error) {
	return "hello", errors.New("world")
}
