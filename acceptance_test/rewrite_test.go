package acceptance_test

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/int128/errto/pkg/log"
	"github.com/int128/errto/pkg/rewrite"
)

func TestRewrite(t *testing.T) {
	log.Printf = t.Logf
	ctx := context.TODO()

	t.Run("pkg-errors to xerrors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.Xerrors, "testdata/pkgerrors/main.go", "testdata/xerrors/main.go")
	})
	t.Run("go-errors to xerrors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.Xerrors, "testdata/goerrors/main.go", "testdata/xerrors/main.go")
	})
	t.Run("pkg-errors to go-errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.GoErrors, "testdata/pkgerrors/main.go", "testdata/goerrors/main.go")
	})
	t.Run("xerrors to go-errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.GoErrors, "testdata/xerrors/main.go", "testdata/goerrors/main.go")
	})
	t.Run("go-errors to pkg-errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.PkgErrors, "testdata/goerrors/main.go", "testdata/pkgerrors/main.go")
	})
	t.Run("xerrors to pkg-errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		testRewrite(t, ctx, rewrite.PkgErrors, "testdata/xerrors/main.go", "testdata/pkgerrors/main.go")
	})
}

func testRewrite(t *testing.T, ctx context.Context, target rewrite.Method, fixtureFilename, wantFilename string) {
	tempDir, err := ioutil.TempDir(".", "fixture")
	if err != nil {
		t.Fatalf("could not create a temp dir: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("could not remove the temp dir: %s", err)
		}
	}()
	tempFile, err := os.Create(filepath.Join(tempDir, "main.go"))
	if err != nil {
		t.Fatalf("could not create a temp file: %s", err)
	}
	defer tempFile.Close()
	fixtureFile, err := os.Open(fixtureFilename)
	if err != nil {
		t.Fatalf("could not open the fixture file: %s", err)
	}
	defer fixtureFile.Close()
	if _, err := io.Copy(tempFile, fixtureFile); err != nil {
		t.Fatalf("could not copy the fixture: %s", err)
	}

	if err := rewrite.Do(ctx, rewrite.Input{Target: target, PkgNames: []string{"./" + tempDir}}); err != nil {
		t.Errorf("error: %+v", err)
	}

	wantContent, err := ioutil.ReadFile(wantFilename)
	if err != nil {
		t.Fatalf("could not read the want file: %s", err)
	}
	gotContent, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("could not read the fixture file: %s", err)
	}
	if diff := diffLines(wantContent, gotContent); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func diffLines(a []byte, b []byte) string {
	sa := strings.Split(string(a), "\n")
	sb := strings.Split(string(b), "\n")
	return cmp.Diff(sa, sb)
}
