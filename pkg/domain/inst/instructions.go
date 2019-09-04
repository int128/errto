package inst

import "go/ast"

type PackageFunctionCallMutator interface {
	PackagePath() string
	PackageName() string
	FunctionName() string
	Args() []ast.Expr
	SetPackageName(pkgName string)
	SetFunctionName(name string)
	SetArgs(args []ast.Expr)
}
