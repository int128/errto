# migerr

A tiny tool to migrate error handling in Go.

Currently the following migration is implemented.

- `github.com/pkg/errors.Errorf` to `xerrors.Errorf`
- `github.com/pkg/errors.Wrapf` to `xerrors.Errorf`
