package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

// check returns nil if s is a positive number.
func check(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid number: %w", err)
	}
	if n < 0 {
		// comment should be kept
		return fmt.Errorf("number is negative: %d", n)
	}
	if n == 0 {
		return errors.New("number is zero")
	}
	return nil
}

// main is an entry point.
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s NUMBER", os.Args[0])
	}

	err := check(os.Args[1])
	log.Printf("err=%+v", err)
	if err != nil {
		log.Printf("Unwrap(err)=%+v", errors.Unwrap(err))
	}
}
