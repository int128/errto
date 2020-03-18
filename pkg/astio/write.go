package astio

import (
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"os"

	"golang.org/x/tools/go/packages"
)

func Write(pkg *packages.Package, file *ast.File) error {
	p := Position(pkg, file)
	if p.Filename == "" {
		return errors.New("could not determine filename")
	}
	f, err := os.Create(p.Filename)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", p.Filename, err)
	}
	defer f.Close()
	if err := printer.Fprint(f, pkg.Fset, file); err != nil {
		return fmt.Errorf("could not write to file %s: %w", p.Filename, err)
	}
	return nil
}
