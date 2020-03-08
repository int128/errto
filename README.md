# transerr [![CircleCI](https://circleci.com/gh/int128/transerr.svg?style=shield)](https://circleci.com/gh/int128/transerr)

This is a command to rewrite Go error handling code between `errors`, `xerrors` and `github.com/pkg/errors`.

**Status**: Proof of Concept.


## Getting Started

Install the latest release.

```sh
go get github.com/int128/transerr
```

Run.

```
% transerr rewrite ./testdata
testdata/hello.go:4:2: rewriting the import with golang.org/x/xerrors
testdata/hello.go:11:9: rewriting the function call with xerrors.New()
testdata/hello.go:12:6: rewriting the function call with xerrors.Errorf()
testdata/hello.go:13:6: rewriting the function call with xerrors.Errorf()
testdata/hello.go: total 4 change(s)
```

You can see the following changes.

```patch
index eb32e7e..5ed2504 100644
--- a/testdata/hello.go
+++ b/testdata/hello.go
@@ -1,15 +1,15 @@
 package testdata

 import (
-       "github.com/pkg/errors"
+       "golang.org/x/xerrors"
 )

 var msg = "foo"

 // Hello says hello world!
 func Hello() error {
-       err := errors.New("message")
-       _ = errors.Wrapf(err, "message %s", msg)
-       _ = errors.Errorf("message %s", msg)
+       err := xerrors.New("message")
+       _ = xerrors.Errorf("message %s: %w", msg, err)
+       _ = xerrors.Errorf("message %s", msg)
        return nil
 }
```


## Rewrite rules

Currently the following rules are supported.

1. `github.com/pkg/errors` to `golang.org/x/xerrors`
2. `github.com/pkg/errors` to `xerrors` (Go 1.13+)

### 1. `github.com/pkg/errors` to `golang.org/x/xerrors`

| From | To |
|------|----|
| `import "github.com/pkg/errors"` | `import "golang.org/x/xerrors"` |
| `errors.Errorf("message %s", msg)` | `xerrors.Errorf("message %s", msg)` |
| `errors.New("message")` | `xerrors.New("message")` |
| `errors.Wrapf(err, "message %s", msg)` | `xerrors.Errorf("message %s: %w", msg, err)` |
| `errors.Cause(err)` | `xerrors.Unwrap(err)` |
| `errors.Wrap()` | TODO |
| `errors.WithStack()` | TODO |
| `errors.WithMessage()` | TODO |
| `errors.WithMessagef()` | TODO |

### 2. `github.com/pkg/errors` to `xerrors` (Go 1.13+)

| From | To |
|------|----|
| `import "github.com/pkg/errors"` | `import "errors"` |
| `errors.Errorf("message %s", msg)` | `fmt.Errorf("message %s", msg)` |
| `errors.New("message")` | `errors.New("message")` |
| `errors.Wrapf(err, "message %s", msg)` | `fmt.Errorf("message %s: %w", msg, err)` |
| `errors.Cause(err)` | `errors.Unwrap(err)` |
| `errors.Wrap()` | TODO |
| `errors.WithStack()` | TODO |
| `errors.WithMessage()` | TODO |
| `errors.WithMessagef()` | TODO |


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
