package rewrite

import (
	"testing"

	"github.com/int128/errto/pkg/log"
)

func TestToGoErrors_Transform(t *testing.T) {
	log.Printf = t.Logf
	var tr toGoErrors

	t.Run("common syntax from xerrors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/xerrors/common.go",
			"testdata/goerrors/common.go")
	})
	t.Run("common syntax from pkg-errors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/pkgerrors/common.go",
			"testdata/goerrors/common.go")
	})
}
