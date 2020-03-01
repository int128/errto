package transform

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"github.com/int128/transerr/pkg/astio"
	"golang.org/x/xerrors"
)

type Input struct {
	PkgNames []string
	DryRun   bool
}

func Do(ctx context.Context, in Input) error {
	pkgs, err := astio.Load(ctx, in.PkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			p := astio.Position(pkg, file)
			v := &pkgErrorsToXerrorsVisitor{}
			if err := astio.Inspect(pkg, file, v); err != nil {
				return xerrors.Errorf("could not inspect the file: %w", err)
			}
			if v.changes == 0 {
				continue
			}
			log.Printf("%s: total %d change(s)", p.Filename, v.changes)
			if !in.DryRun {
				if err := astio.Write(pkg, file); err != nil {
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

type pkgErrorsToXerrorsVisitor struct {
	changes int
}

func (v *pkgErrorsToXerrorsVisitor) Import(p token.Position, spec *ast.ImportSpec) error {
	name := strings.Trim(spec.Path.Value, `"`)
	if name != pkgErrorsPkgPath {
		return nil
	}
	log.Printf("%s: rewriting the import with %s", p, xerrorsPkgPath)
	spec.Path.Value = fmt.Sprintf(`"%s"`, xerrorsPkgPath)
	v.changes++
	return nil
}

func (v *pkgErrorsToXerrorsVisitor) PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, pkgName *types.PkgName, fun *ast.SelectorExpr) error {
	packagePath := pkgName.Imported().Path()
	if packagePath != pkgErrorsPkgPath {
		return nil
	}

	functionName := fun.Sel.Name
	switch functionName {
	case "Wrapf":
		log.Printf("%s: rewriting the function call with xerrors.Errorf()", p)
		pkg.Name = "xerrors"
		fun.Sel.Name = "Errorf"

		// reorder the args
		a := call.Args
		args := make([]ast.Expr, 0)
		args = append(args, a[1])
		args = append(args, a[2:]...)
		args = append(args, a[0])
		call.Args = a

		// append %w to the format arg
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

	case "Errorf", "New":
		log.Printf("%s: rewriting the function call with xerrors.%s()", p, functionName)
		pkg.Name = "xerrors"
		v.changes++
		return nil

	default:
		log.Printf("%s: NOTE: you need to manually rewrite errors.%s()", p, functionName)
		return nil
	}
}
