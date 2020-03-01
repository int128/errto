package transform

import (
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

type Input struct {
	PkgNames []string
	DryRun   bool
}

func Do(ctx context.Context, in Input) error {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, in.PkgNames...)
	if err != nil {
		return xerrors.Errorf("could not load the packages: %w", err)
	}
	if n := packages.PrintErrors(pkgs); n > 0 {
		return xerrors.Errorf("could not load the packages with %d errors", n)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			p := position(pkg, file)
			v := &pkgErrorsToXerrorsVisitor{}
			if err := inspect(pkg, file, v); err != nil {
				return xerrors.Errorf("could not inspect the file: %w", err)
			}
			if v.changes == 0 {
				continue
			}
			log.Printf("%s: total %d change(s)", p.Filename, v.changes)
			if !in.DryRun {
				if err := write(pkg, file); err != nil {
					return xerrors.Errorf("could not write the file: %w", err)
				}
			}
		}
	}
	return nil
}

func write(pkg *packages.Package, file *ast.File) error {
	p := position(pkg, file)
	if p.Filename == "" {
		return xerrors.Errorf("could not determine filename of file %s", file)
	}
	f, err := os.Create(p.Filename)
	if err != nil {
		return xerrors.Errorf("could not open file %s: %w", p.Filename, err)
	}
	defer f.Close()
	if err := printer.Fprint(f, pkg.Fset, file); err != nil {
		return xerrors.Errorf("could not write to file %s: %w", p.Filename, err)
	}
	return nil
}

func position(pkg *packages.Package, node ast.Node) token.Position {
	p := pkg.Fset.Position(node.Pos())
	p.Filename = relative(p.Filename)
	return p
}

func relative(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return name
	}
	return strings.TrimPrefix(name, wd+"/")
}

type Visitor interface {
	Import(p token.Position, spec *ast.ImportSpec) error
	PackageFunctionCall(p token.Position, call *ast.CallExpr, pkg *ast.Ident, pkgName *types.PkgName, fun *ast.SelectorExpr) error
}

func inspect(pkg *packages.Package, file *ast.File, v Visitor) error {
	var lastErr error
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			p := position(pkg, node)
			switch node.Tok {
			case token.IMPORT:
				for _, spec := range node.Specs {
					switch spec := spec.(type) {
					case *ast.ImportSpec:
						if err := v.Import(p, spec); err != nil {
							lastErr = err
							return false
						}
					default:
						lastErr = xerrors.Errorf("spec wants *ast.ImportSpec but was %T", spec)
						return false
					}
				}
			}
		case *ast.CallExpr:
			p := position(pkg, node)
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				switch x := fun.X.(type) {
				case *ast.Ident:
					switch o := pkg.TypesInfo.ObjectOf(x).(type) {
					case *types.PkgName:
						if err := v.PackageFunctionCall(p, node, x, o, fun); err != nil {
							lastErr = err
							return false
						}
					}
				}
			}
		}
		return true
	})
	return lastErr
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
