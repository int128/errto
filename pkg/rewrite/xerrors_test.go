package rewrite

import (
	"testing"

	"github.com/int128/errto/pkg/log"
)

func TestToXerrors_Transform(t *testing.T) {
	log.Printf = t.Logf
	var tr toXerrors

	t.Run("common syntax from go-errors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/goerrors/common.go",
			"testdata/xerrors/common.go")
	})
	t.Run("common syntax from pkg-errors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/pkgerrors/common.go",
			"testdata/xerrors/common.go")
	})
	t.Run("special syntax from pkg-errors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/pkgerrors/from_pkgerrors.go",
			"testdata/xerrors/from_pkgerrors.go")
	})
}
