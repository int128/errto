# migerr

This is a tiny tool to migrate error handling in Go.

Currently the following migrations are implemented.

- `github.com/pkg/errors.Errorf` to `xerrors.Errorf`
- `github.com/pkg/errors.Wrapf` to `xerrors.Errorf`

**Status**: Alpha, proof of concept.


## Getting Started

```
% migerr migrate --dry-run ./testdata
testdata/hello.go:7:2: rewriting the import with golang.org/x/xerrors
testdata/hello.go:15:10: rewriting the function call with xerrors.Errorf()
testdata/hello.go:18:10: rewriting the function call with xerrors.Errorf()
testdata/hello.go: total 3 change(s)
```
