package main

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
)

func check1(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return errors.Wrap(err, "invalid number")
	}
	return nil
}

func check2(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func check3(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return errors.WithMessage(err, "invalid number")
	}
	return nil
}

func check4(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return errors.WithMessagef(err, "invalid number: %s", s)
	}
	return nil
}

// main is an entry point.
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s NUMBER", os.Args[0])
	}

	log.Printf("err=%s", check1(os.Args[1]))
	log.Printf("err=%s", check2(os.Args[1]))
	log.Printf("err=%s", check3(os.Args[1]))
	log.Printf("err=%s", check4(os.Args[1]))

	log.Printf("stacktrace=\n%+v", check1(os.Args[1]))
	log.Printf("stacktrace=\n%+v", check2(os.Args[1]))
	log.Printf("stacktrace=\n%+v", check3(os.Args[1]))
	log.Printf("stacktrace=\n%+v", check4(os.Args[1]))
}
