# migerr

This is a tiny tool to migrate error handling in Go.

Currently the following migrations are implemented.

- `github.com/pkg/errors.Errorf` to `xerrors.Errorf`
- `github.com/pkg/errors.Wrapf` to `xerrors.Errorf`

**Status**: Alpha, proof of concept.


## Getting Started

```
% migerr migrate --dry-run ./testdata
/src/migerr/testdata/hello.go:7:2: rewriting the import with golang.org/x/xerrors
/src/migerr/testdata/hello.go:15:10: rewriting the function call with xerrors.Errorf()
/src/migerr/testdata/hello.go:18:10: rewriting the function call with xerrors.Errorf()
/src/migerr/testdata/hello.go: total 3 change(s)
package testdata

import (
	"fmt"
	"os"

	"golang.org/x/xerrors"
)

var msg = "foo"

// Hello says hello world!
func Hello() error {
	if _, err := fmt.Fprintf(os.Stderr, "hello world"); err != nil {
		return xerrors.Errorf("error %s: %w", msg, err)
	}
	if _, err := fmt.Fprintf(os.Stderr, "hello world"); err != nil {
		return xerrors.Errorf("error %s", msg)
	}
	return nil
}
```
