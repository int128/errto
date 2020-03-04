package rewrite

import (
	"context"

	"github.com/int128/transerr/pkg/astio"
	"github.com/int128/transerr/pkg/log"
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
			v := newVisitor(in.Target)
			if v == nil {
				return xerrors.Errorf("unknown target method %v", in.Target)
			}
			if err := astio.Inspect(pkg, file, v); err != nil {
				return xerrors.Errorf("could not inspect the file: %w", err)
			}
			if v.Changes() == 0 {
				continue
			}
			p := astio.Position(pkg, file)
			if !in.DryRun {
				log.Printf("%s: writing %d change(s)", p.Filename, v.Changes())
				if err := astio.Write(pkg, file); err != nil {
					return xerrors.Errorf("could not write the file: %w", err)
				}
			}
		}
	}
	return nil
}

type Visitor interface {
	astio.Visitor
	Changes() int
}

func newVisitor(m Method) Visitor {
	switch m {
	case Xerrors:
		return &toXerrorsVisitor{}
	}
	return nil
}
