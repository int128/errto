package migrate

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/int128/migerr/pkg/adaptors/inspector"
	"golang.org/x/xerrors"
)

type Interface interface {
	Do(ctx context.Context, cfg Config) error
}

type Config struct {
	PkgNames []string
	DryRun   bool
}

type UseCase struct {
	Inspector inspector.Inspector
}

func (uc *UseCase) Do(ctx context.Context, cfg Config) error {
	pkgs, err := uc.Inspector.Load(ctx, cfg.PkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			filename := uc.Inspector.Filename(pkg, file)
			v := &pkgErrorsToXerrorsMigration{}
			if err := uc.Inspector.Inspect(pkg, file, v); err != nil {
				return xerrors.Errorf("could not migrate the file: %w", err)
			}
			if v.changes == 0 {
				continue
			}
			log.Printf("%s: total %d change(s)", filename, v.changes)
			if cfg.DryRun {
				if err := uc.Inspector.Print(pkg, file); err != nil {
					return xerrors.Errorf("could not print the file: %w", err)
				}
			} else {
				if err := uc.Inspector.Write(pkg, file); err != nil {
					return xerrors.Errorf("could not write the file: %w", err)
				}
			}
		}
	}
	return nil
}

const (
	pkgErrorsPkgPath = "github.com/pkg/errors"
	xerrorsPkgPath   = "golang.org/x/xerrors"
)

type pkgErrorsToXerrorsMigration struct {
	changes int
}

func (v *pkgErrorsToXerrorsMigration) Import(imp inspector.Import) error {
	if imp.PackagePath() != pkgErrorsPkgPath {
		return nil
	}
	log.Printf("%s: rewriting the import with %s", imp.Position(), xerrorsPkgPath)
	imp.SetPackagePath(xerrorsPkgPath, "")
	v.changes++
	return nil
}

func (v *pkgErrorsToXerrorsMigration) PackageFunctionCall(c inspector.PackageFunctionCall) error {
	if c.PackagePath() != pkgErrorsPkgPath {
		return nil
	}

	switch c.FunctionName() {
	case "Wrapf":
		log.Printf("%s: rewriting the function call with xerrors.Errorf()", c.Position())
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
		v.changes++
		return nil

	case "Errorf":
		log.Printf("%s: rewriting the function call with xerrors.Errorf()", c.Position())
		c.SetPackageName("xerrors")
		c.SetFunctionName("Errorf")
		v.changes++
		return nil

	default:
		log.Printf("%s: you need to manually migrate the function call", c.Position())
		return nil
	}
}
