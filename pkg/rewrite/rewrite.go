package rewrite

import (
	"context"
	"go/ast"

	"github.com/int128/transerr/pkg/astio"
	"github.com/int128/transerr/pkg/log"
	"golang.org/x/tools/go/packages"
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

type Transformer interface {
	Transform(pkg *packages.Package, file *ast.File) (int, error)
}

func newTransformer(m Method) Transformer {
	switch m {
	case Xerrors:
		return &toXerrors{}
	case GoErrors:
		return &toGoErrors{}
	case PkgErrors:
		return &toPkgErrors{}
	}
	return nil
}
