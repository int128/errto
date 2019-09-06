# migerr [![CircleCI](https://circleci.com/gh/int128/migerr.svg?style=shield)](https://circleci.com/gh/int128/migerr)

This is a command line tool to migrate error handling in Go files.

**Status**: Alpha, proof of concept.


## Getting Started

```
% migerr migrate --dry-run ./testdata
testdata/hello.go:7:2: rewriting the import with golang.org/x/xerrors
testdata/hello.go:15:10: rewriting the function call with xerrors.Errorf()
testdata/hello.go:18:10: rewriting the function call with xerrors.Errorf()
testdata/hello.go: total 3 change(s)
```


## Migration rules

Currently the following migration rules are supported.

- `github.com/pkg/errors` to `xerrors`
  - `errors.Errorf` to `xerrors.Errorf`
  - `errors.New` to `xerrors.New`
  - `errors.Wrapf` to `xerrors.Errorf`

### `github.com/pkg/errors` to `xerrors`

This rule set performs the following migration:

| From | To |
|------|----|
| `import "github.com/pkg/errors"` | `import "golang.org/x/xerrors"` |
| `errors.Errorf("message %s", msg)` | `xerrors.Errorf("message %s", msg)` |
| `errors.New("message")` | `xerrors.New("message")` |
| `errors.Wrapf(err, "message %s", msg)` | `xerrors.Errorf("message %s: %w", msg, err)` |


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
