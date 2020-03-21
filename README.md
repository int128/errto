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

To rewrite the packages in the current working directory:

```sh
# rewrite with Go 1.13+ errors
errto go-errors ./...

# rewrite with golang.org/x/xerrors
errto xerrors ./...

# rewrite with github.com/pkg/errors
errto pkg-errors ./...
```

For example, to rewrite package `acceptance_test/testdata/basic/pkgerrors` in this repository:

```console
% errto xerrors ./acceptance_test/testdata/basic/pkgerrors
acceptance_test/testdata/basic/pkgerrors/main.go:14:10: errors.Wrapf() -> xerrors.Errorf()
acceptance_test/testdata/basic/pkgerrors/main.go:18:10: errors.Errorf() -> xerrors.Errorf()
acceptance_test/testdata/basic/pkgerrors/main.go:21:10: errors.New() -> xerrors.New()
acceptance_test/testdata/basic/pkgerrors/main.go:35:33: errors.Cause() -> xerrors.Unwrap()
acceptance_test/testdata/basic/pkgerrors/main.go: + import golang.org/x/xerrors
acceptance_test/testdata/basic/pkgerrors/main.go: - import github.com/pkg/errors
--- writing 6 change(s) to acceptance_test/testdata/basic/pkgerrors/main.go
```

It will rewrite [`main.go`](acceptance_test/testdata/basic/pkgerrors/main.go) as follows:

```patch
--- a/acceptance_test/testdata/basic/pkgerrors/main.go
+++ b/acceptance_test/testdata/basic/pkgerrors/main.go
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

You can see changes without actually writing files by `--dry-run` flag.

```sh
errto go-errors --dry-run ./...
```

It is recommended to commit files into a Git repository before running the command.


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

The following syntax is supported.

| go-errors | xerrors | pkg-errors |
|-----------|---------|------------|
| `New("MESSAGE")` | `New("MESSAGE")` | `New("MESSAGE")` |
| `Errorf("FORMAT", ...)` | `Errorf("FORMAT", ...)` | `Errorf("FORMAT", ...)` |
| `Errorf("FORMAT: %w", ..., err)` | `Errorf("FORMAT: %w", ..., err)` | `Wrapf(err, "FORMAT", ...)` |
| `Unwrap(err)` | `Unwrap(err)` | `Cause(err)` <sup>1</sup> |
| `As(err, v)` | `As(err, v)` | - |
| `Is(err, v)` | `Is(err, v)` | - |
| `Errorf("%s: %w", "MSG", err)` | `Errorf("%s: %w", "MSG", err)` | `Wrap(err, "MSG")` <sup>2</sup> |
| `Errorf("%w", err)` | `Errorf("%w", err)` | `WithStack(err)` <sup>2</sup> |
| `Errorf("%s: %s", "MSG", err)` | `Errorf("%s: %s", "MSG", err)` | `WithMessage(err, "MSG")` <sup>2</sup> |
| `Errorf("FORMAT: %s", ..., err)` | `Errorf("FORMAT: %s", ..., err)` | `WithMessagef(err, "FORMAT", ...)` <sup>2</sup> |

<sup>1</sup> Supported only if rewriting from pkg-errors to go-errors or xerrors.

<sup>2</sup> Compatible with Go errors since `github.com/pkg/errors@v0.9.0`. See [the release note](https://github.com/pkg/errors/releases/tag/v0.9.0) for details.


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
