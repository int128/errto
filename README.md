# migerr

This is a tiny tool to migrate error handling in Go.

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

Currently the following migrations are supported.

- `github.com/pkg/errors.Errorf` to `xerrors.Errorf`
- `github.com/pkg/errors.New` to `xerrors.New`
- `github.com/pkg/errors.Wrapf` to `xerrors.Errorf`
