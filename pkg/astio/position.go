package astio

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func Position(pkg *packages.Package, node ast.Node) token.Position {
	p := pkg.Fset.Position(node.Pos())
	p.Filename = relative(p.Filename)
	return p
}

func Filename(pkg *packages.Package, file *ast.File) string {
	p := pkg.Fset.Position(file.Pos())
	return relative(p.Filename)
}

func relative(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return name
	}
	return strings.TrimPrefix(name, wd+"/")
}
