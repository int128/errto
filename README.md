# errto [![CircleCI](https://circleci.com/gh/int128/errto.svg?style=shield)](https://circleci.com/gh/int128/errto)

This is a command to rewrite Go error handling code between the following packages:

- `errors` (1.13+)
- `golang.org/x/xerrors`
- `github.com/pkg/errors`

It rewrites the package imports and function calls using AST transformation.
All whitespaces and comments are kept.


## Getting Started

Install the latest release.

```sh
go get github.com/int128/errto
```

To rewrite package(s) with `golang.org/x/xerrors`:

```
% errto xerrors ./pkg/rewrite/testdata/basic/pkgerrors
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go:14:10: pkg/errors.Wrapf() -> xerrors.Errorf()
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go:18:10: pkg/errors.Errorf() -> xerrors.Errorf()
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go:21:10: pkg/errors.New() -> xerrors.New()
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go:35:33: pkg/errors.Cause() -> xerrors.Unwrap()
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go: + import golang.org/x/xerrors
rewrite: pkg/rewrite/testdata/basic/pkgerrors/main.go: - import github.com/pkg/errors
writing 6 change(s) to pkg/rewrite/testdata/basic/pkgerrors/main.go
```

It will rewrite [`pkg/rewrite/testdata/basic/pkgerrors/main.go`](pkg/rewrite/testdata/basic/pkgerrors/main.go) as follows:

```patch
--- a/pkg/rewrite/testdata/basic/pkgerrors/main.go
+++ b/pkg/rewrite/testdata/basic/pkgerrors/main.go
@@ -1,7 +1,7 @@
 package main

 import (
-       "github.com/pkg/errors"
+       "golang.org/x/xerrors"
        "log"
        "os"
        "strconv"
@@ -11,14 +11,14 @@ import (
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
@@ -32,6 +32,6 @@ func main() {
        err := check(os.Args[1])
        log.Printf("err=%+v", err)
        if err != nil {
-               log.Printf("Unwrap(err)=%+v", errors.Cause(err))
+               log.Printf("Unwrap(err)=%+v", xerrors.Unwrap(err))
        }
 }
```


## Usage

```
Usage:
  errto [command]

Available Commands:
  dump        Dump AST of packages
  go-errors   Rewrite the packages with Go errors (fmt, errors)
  help        Help about any command
  pkg-errors  Rewrite the packages with github.com/pkg/errors
  xerrors     Rewrite the packages with golang.org/x/xerrors
```

### Rewrite commands

You can run the following commands.

- `go-errors`: rewrite with Go errors (1.13+)
- `xerrors`: rewrite with `golang.org/x/xerrors`
- `pkg-errors`: rewrite with `github.com/pkg/errors`

The following syntax is supported.

| go-errors | xerrors | pkg-errors |
|-----------|---------|------------|
| `errors.New("MESSAGE")` | `New("MESSAGE")` | `New("MESSAGE")` |
| `fmt.Errorf("FORMAT", ...)` | `Errorf("FORMAT", ...)` | `Errorf("FORMAT", ...)` |
| `fmt.Errorf("FORMAT: %w", ..., err)` | `Errorf("FORMAT: %w", ..., err)` | `Wrapf(err, "FORMAT", ...)` |
| `errors.Unwrap(err)` | `Unwrap(err)` | `Cause(err)` <sup>2</sup> |
| `errors.As(err, v)` | `As(err, v)` | - |
| `errors.Is(err, v)` | `Is(err, v)` | - |
| `fmt.Errorf("%s: %w", "MESSAGE", err)` | `Errorf("%s: %w", "MESSAGE", err)` | `Wrap(err, "MESSAGE")` <sup>1</sup> |
| `fmt.Errorf("%w", err)` | `Errorf("%w", err)` | `WithStack(err)` <sup>1</sup> |
| `fmt.Errorf("%s: %s", "MESSAGE", err)` | `Errorf("%s: %s", "MESSAGE", err)` | `WithMessage(err, "MESSAGE")` <sup>1</sup> |
| `fmt.Errorf("FORMAT: %s", ..., err)` | `Errorf("FORMAT: %s", ..., err)` | `WithMessagef(err, "FORMAT", ...)` <sup>1</sup> |

<sup>1</sup> Only rewriting from pkg-errors to go-errors or xerrors is supported. Opposite is not supported.

<sup>2</sup> Compatible with Go errors since v0.9.0. See [the release note](https://github.com/pkg/errors/releases/tag/v0.9.0) for details.


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
