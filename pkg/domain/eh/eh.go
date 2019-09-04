package eh

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/int128/migerr/pkg/domain/inst"
	"golang.org/x/xerrors"
)

func PkgErrorsToXerrors(m inst.PackageFunctionCallMutator) error {
	if m.PackagePath() != "github.com/pkg/errors" {
		return nil
	}
	switch m.FunctionName() {
	case "Wrapf":
		m.SetPackageName("golang.org/x/xerrors")
		m.SetFunctionName("Errorf")

		a := m.Args()
		args := make([]ast.Expr, 0)
		args = append(args, a[1])
		args = append(args, a[2:]...)
		args = append(args, a[0])
		m.SetArgs(args)

		b, ok := a[1].(*ast.BasicLit)
		if !ok {
			return xerrors.Errorf("2nd argument of Wrapf must be a literal but %T", a[1])
		}
		if b.Kind != token.STRING {
			return xerrors.Errorf("2nd argument of Wrapf must be a string but %s", b.Kind)
		}
		b.Value = fmt.Sprintf(`"%s: %%w"`, strings.Trim(b.Value, `"`))
		return nil

	case "Errorf":
		m.SetPackageName("golang.org/x/xerrors")
		m.SetFunctionName("Errorf")
		return nil
	}
	return nil
}
