package rewrite

import (
	"context"

	"github.com/int128/errto/pkg/astio"
	"github.com/int128/errto/pkg/log"
	"golang.org/x/xerrors"
)

type Method int

const (
	_ = Method(iota)
	GoErrors
	Xerrors
	PkgErrors
)

const (
	pkgErrorsImportPath = "github.com/pkg/errors"
	xerrorsImportPath   = "golang.org/x/xerrors"
)

type Input struct {
	PkgNames []string
	Target   Method
	DryRun   bool
}

func Do(ctx context.Context, in Input) error {
	pkgs, err := astio.Load(ctx, in.PkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	if len(pkgs) == 0 {
		return xerrors.New("no package found")
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			t := newTransformer(in.Target)
			if t == nil {
				return xerrors.Errorf("unknown target method %v", in.Target)
			}
			n, err := t.Transform(pkg, file)
			if err != nil {
				return xerrors.Errorf("could not rewrite the file: %w", err)
			}
			if n == 0 {
				continue
			}
			if !in.DryRun {
				p := astio.Position(pkg, file)
				log.Printf("writing %d change(s) to %s", n, p.Filename)
				if err := astio.Write(pkg, file); err != nil {
					return xerrors.Errorf("could not write the file: %w", err)
				}
			}
		}
	}
	return nil
}
