# transerr [![CircleCI](https://circleci.com/gh/int128/transerr.svg?style=shield)](https://circleci.com/gh/int128/transerr)

This is a command to rewrite Go error handling code between the following packages:

- `errors` (1.13+)
- `golang.org/x/xerrors`
- `github.com/pkg/errors`

It rewrites the package imports and package function calls using AST transformation.
All whitespaces and comments are kept.


## Getting Started

Install the latest release.

```sh
go get github.com/int128/transerr
```

Run the following command.

```
% transerr rewrite --to=xerrors ./pkg/rewrite/testdata/basic/pkgerrors
pkg/rewrite/testdata/basic/pkgerrors/main.go:8:2: rewrite: import github.com/pkg/errors -> golang.org/x/xerrors
pkg/rewrite/testdata/basic/pkgerrors/main.go:15:10: rewrite: pkg/errors.Wrapf() -> xerrors.Errorf()
pkg/rewrite/testdata/basic/pkgerrors/main.go:19:10: rewrite: pkg/errors.Errorf() -> xerrors.Errorf()
pkg/rewrite/testdata/basic/pkgerrors/main.go:22:10: rewrite: pkg/errors.New() -> xerrors.New()
pkg/rewrite/testdata/basic/pkgerrors/main.go:36:33: rewrite: pkg/errors.Cause() -> xerrors.Unwrap()
pkg/rewrite/testdata/basic/pkgerrors/main.go: writing 5 change(s)
```

Then [`pkg/rewrite/testdata/basic/pkgerrors/main.go`](pkg/rewrite/testdata/basic/pkgerrors/main.go) will be rewrote as follows:

```patch
--- a/pkg/rewrite/testdata/basic/pkgerrors/main.go
+++ b/pkg/rewrite/testdata/basic/pkgerrors/main.go
@@ -5,21 +5,21 @@ import (
        "os"
        "strconv"

-       "github.com/pkg/errors"
+       "golang.org/x/xerrors"
 )

 // check returns nil if s is a positive number.
 func check(s string) error {
        n, err := strconv.Atoi(s)
        if err != nil {
-               return errors.Wrapf(err, "invalid number")
+               return xerrors.Errorf("invalid number: %w", err)
        }
        if n < 0 {
                // comment should be kept
-               return errors.Errorf("number is negative: %d", n)
+               return xerrors.Errorf("number is negative: %d", n)
        }
        if n == 0 {
-               return errors.New("number is zero")
+               return xerrors.New("number is zero")
        }
        return nil
 }
@@ -33,6 +33,6 @@ func main() {
        err := check(os.Args[1])
        log.Printf("err=%+v", err)
        if err != nil {
-               log.Printf("Unwrap(err)=%+v", errors.Cause(err))
+               log.Printf("Unwrap(err)=%+v", xerrors.Unwrap(err))
        }
 }
```


## Usage

### Rewrite command

```
Usage:
  transerr rewrite [flags] --to=METHOD PACKAGE...

Flags:
      --dry-run     Do not write files actually
  -h, --help        help for rewrite
      --to string   Target error handling method (go-errors|xerrors|pkg-errors)
```

It supports the following packages:

- `errors` (1.13+)
- `golang.org/x/xerrors`
- `github.com/pkg/errors`

The following rewrite rules are supported.

- Import the package.
  - `import "errors"`, `import "fmt"`
  - `import "golang.org/x/xerrors"`
  - `import "github.com/pkg/errors"`
- Create an error.
  - `errors.New("message")`
  - `golang.org/x/xerrors.New("message")`
  - `github.com/pkg/errors.New("message")`
- Format an error.
  - `fmt.Errorf("message %s", msg)`
  - `github.com/pkg/errors.Errorf("message %s", msg)`
  - `golang.org/x/xerrors.Errorf("message %s", msg)`
- Wrap an error.
  - `fmt.Errorf("message %s: %w", msg, err)`
  - `golang.org/x/xerrors.Errorf("message %s: %w", msg, err)`
  - `github.com/pkg/errors.Wrapf(err, "message %s", msg)`
- Unwrap an error.
  - `errors.Unwrap(err)`
  - `golang.org/x/xerrors.Unwrap(err)`
  - `github.com/pkg/errors.Cause(err)`
- Unwrap and cast an error.
  - `errors.As(err, v)`
  - `golang.org/x/xerrors.As(err, v)`
- Unwrap and test an error.
  - `errors.Is(err, v)`
  - `golang.org/x/xerrors.Is(err, v)`

Not implemented yet:

- `golang.org/x/xerrors.Opaque()`
- `github.com/pkg/errors.Wrap()`
- `github.com/pkg/errors.WithStack()`
- `github.com/pkg/errors.WithMessage()`
- `github.com/pkg/errors.WithMessagef()`


### Dump command

You can dump the AST for debug.

```
Usage:
  transerr dump PACKAGE... [flags]
```


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
