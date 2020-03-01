package astio

import (
	"go/ast"
	"go/printer"
	"os"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

func Write(pkg *packages.Package, file *ast.File) error {
	p := Position(pkg, file)
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
