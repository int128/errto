package main

import (
	"golang.org/x/xerrors"
	"log"
	"os"
	"strconv"
)

// check returns nil if s is a positive number.
func check(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return xerrors.Errorf("invalid number: %w", err)
	}
	if n < 0 {
		// comment should be kept
		return xerrors.Errorf("number is negative: %d", n)
	}
	if n == 0 {
		return xerrors.New("number is zero")
	}
	return nil
}

// main is an entry point.
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s NUMBER", os.Args[0])
	}
	if err := check(os.Args[1]); err != nil {
		log.Printf("error: %+v", err)
	}
}
