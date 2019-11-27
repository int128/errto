# transerr [![CircleCI](https://circleci.com/gh/int128/transerr.svg?style=shield)](https://circleci.com/gh/int128/transerr)

This is a command line tool to transform Go error handling between `errors`, `xerrors` and `github.com/pkg/errors`.

**Status**: Proof of Concept.


## Getting Started

Install the latest release.

```sh
go get github.com/int128/transerr
```

Run the transerr.

```
% transerr migrate ./testdata
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


## Transformation rules

Currently the following rules are supported.

- `github.com/pkg/errors` to `golang.org/x/xerrors`
  - `errors.Errorf` to `xerrors.Errorf`
  - `errors.New` to `xerrors.New`
  - `errors.Wrapf` to `xerrors.Errorf`

### 1. `github.com/pkg/errors` to `golang.org/x/xerrors`

| From | To |
|------|----|
| `import "github.com/pkg/errors"` | `import "golang.org/x/xerrors"` |
| `errors.Errorf("message %s", msg)` | `xerrors.Errorf("message %s", msg)` |
| `errors.New("message")` | `xerrors.New("message")` |
| `errors.Wrapf(err, "message %s", msg)` | `xerrors.Errorf("message %s: %w", msg, err)` |
| `errors.Wrap()` | TODO |
| `errors.Cause()` | TODO |
| `errors.WithStack()` | TODO |
| `errors.WithMessage()` | TODO |
| `errors.WithMessagef()` | TODO |

### 2. `xerrors` to `errors` (Go 1.13)

TODO


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
