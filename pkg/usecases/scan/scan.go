package scan

import (
	"context"
	"log"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"golang.org/x/xerrors"
)

type Interface interface {
	Do(ctx context.Context, pkgNames ...string) error
}

type UseCase struct {
	Inspector inspector.Inspector
}

func (uc *UseCase) Do(ctx context.Context, pkgNames ...string) error {
	pkgs, err := uc.Inspector.Load(ctx, pkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if err := uc.Inspector.Inspect(pkg, file, &visitor{}); err != nil {
				return xerrors.Errorf("could not scan the file %s: %w", file, err)
			}
		}
	}
	return nil
}

type packageFunction struct {
	PackagePath  string
	FunctionName string
}

var filter = []packageFunction{
	{PackagePath: "github.com/pkg/errors", FunctionName: "Errorf"},
	{PackagePath: "github.com/pkg/errors", FunctionName: "Wrapf"},
	{PackagePath: "github.com/pkg/errors", FunctionName: "New"},
}

type visitor struct{}

func (v *visitor) Import(imp inspector.Import) error {
	if imp.PackagePath() == "github.com/pkg/errors" {
		log.Printf("%s: import: %s", imp.Position(), imp.PackagePath())
	}
	return nil
}

func (v *visitor) PackageFunctionCall(c inspector.PackageFunctionCall) error {
	target := packageFunction{PackagePath: c.PackagePath(), FunctionName: c.FunctionName()}
	for _, f := range filter {
		if target == f {
			log.Printf("%s: call: (%s).%s", c.Position(), c.PackagePath(), c.FunctionName())
		}
	}
	return nil
}
