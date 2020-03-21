package rewrite

import (
	"testing"

	"github.com/int128/errto/pkg/log"
)

func TestToPkgErrors_Transform(t *testing.T) {
	log.Printf = t.Logf
	var tr toPkgErrors

	t.Run("common syntax from go-errors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/goerrors/common.go",
			"testdata/pkgerrors/common.go")
	})
	t.Run("common syntax from xerrors", func(t *testing.T) {
		transform(t, &tr,
			"testdata/xerrors/common.go",
			"testdata/pkgerrors/common.go")
	})
}
