package migrate

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"golang.org/x/xerrors"
)

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
				return xerrors.Errorf("could not migrate the file: %w", err)
			}
			if err := uc.Inspector.Print(pkg, file); err != nil {
				return xerrors.Errorf("could not write the file: %w", err)
			}
		}
	}
	return nil
}

type packageFunction struct {
	PackagePath  string
	FunctionName string
}

type visitor struct{}

func (v *visitor) Import(imp inspector.Import) error {
	if imp.PackagePath() == "github.com/pkg/errors" {
		imp.SetPackagePath("golang.org/x/xerrors", "")
	}
	return nil
}

func (v *visitor) PackageFunctionCall(c inspector.PackageFunctionCall) error {
	target := packageFunction{PackagePath: c.PackagePath(), FunctionName: c.FunctionName()}
	switch target {
	case packageFunction{PackagePath: "github.com/pkg/errors", FunctionName: "Wrapf"}:
		c.SetPackageName("xerrors")
		c.SetFunctionName("Errorf")

		// reorder the args
		a := c.Args()
		args := make([]ast.Expr, 0)
		args = append(args, a[1])
		args = append(args, a[2:]...)
		args = append(args, a[0])
		c.SetArgs(args)

		// append %w to the format
		b, ok := a[1].(*ast.BasicLit)
		if !ok {
			return xerrors.Errorf("2nd argument of Wrapf must be a literal but %T", a[1])
		}
		if b.Kind != token.STRING {
			return xerrors.Errorf("2nd argument of Wrapf must be a string but %s", b.Kind)
		}
		b.Value = fmt.Sprintf(`"%s: %%w"`, strings.Trim(b.Value, `"`))
		return nil

	case packageFunction{PackagePath: "github.com/pkg/errors", FunctionName: "Errorf"}:
		c.SetPackageName("xerrors")
		c.SetFunctionName("Errorf")
		return nil
	}
	return nil
}
