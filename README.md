# errto [![CircleCI](https://circleci.com/gh/int128/errto.svg?style=shield)](https://circleci.com/gh/int128/errto)

This is a command to rewrite Go error handling code between the following packages:

- `errors` (1.13+)
- `golang.org/x/xerrors`
- `github.com/pkg/errors`

It rewrites the package imports and package function calls using AST transformation.
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

You can run the following command.

- `go-errors`: rewrite with Go errors (1.13+)
- `xerrors`: rewrite with `golang.org/x/xerrors`
- `pkg-errors`: rewrite with `github.com/pkg/errors`

The following syntax is supported.

| go-errors | xerrors | pkg-errors |
|-----------|---------|------------|
| `errors.New("MESSAGE")` | `xerrors.New("MESSAGE")` | `errors.New("MESSAGE")` |
| `fmt.Errorf("%s", s)` | `xerrors.Errorf("%s", s)` | `errors.Errorf("%s", s)` |
| `fmt.Errorf("MESSAGE: %w", err)` | `xerrors.Errorf("MESSAGE: %w", err)` | `errors.Wrapf(err, "MESSAGE")` |
| `errors.Unwrap(err)` | `xerrors.Unwrap(err)` | `errors.Cause(err)` <sup>1</sup> |
| `errors.As(err, v)` | `xerrors.As(err, v)` | - |
| `errors.Is(err, v)` | `xerrors.Is(err, v)` | - |

<sup>1</sup> Incompatible behavior. You may need to rewrite code manually.


### NOTE: these are not implemented yet

Functions:

- `golang.org/x/xerrors.Opaque()`
- `github.com/pkg/errors.Wrap()`
- `github.com/pkg/errors.WithStack()`
- `github.com/pkg/errors.WithMessage()`
- `github.com/pkg/errors.WithMessagef()`


### Dump command

You can dump the AST for debug.

```
Usage:
  errto dump PACKAGE... [flags]
```


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
