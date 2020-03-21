package rewrite

import (
	"context"
	"go/printer"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/int128/errto/pkg/astio"
)

func transform(t *testing.T, transformer Transformer, fixtureFilename, wantFilename string) {
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
	t.Logf("wrote %s", tempFile.Name())

	pkgs, err := astio.Load(context.TODO(), "./"+tempDir)
	if err != nil {
		t.Fatalf("could not load the fixture package: %s", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("len(pkgs) wants 1 but was %d", len(pkgs))
	}
	if len(pkgs[0].Syntax) != 1 {
		t.Fatalf("len(pkgs[0].Syntax) wants 1 but was %d", len(pkgs[0].Syntax))
	}
	n, err := transformer.Transform(pkgs[0], pkgs[0].Syntax[0])
	if err != nil {
		t.Fatalf("could not transform: %s", err)
	}
	t.Logf("%d change(s)", n)
	var w strings.Builder
	if err := printer.Fprint(&w, pkgs[0].Fset, pkgs[0].Syntax[0]); err != nil {
		t.Fatalf("could not print the AST: %s", err)
	}
	got := w.String()

	wantContent, err := ioutil.ReadFile(wantFilename)
	if err != nil {
		t.Fatalf("could not read the want file: %s", err)
	}
	want := string(wantContent)
	if diff := diffLines(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func diffLines(a string, b string) string {
	sa := strings.Split(a, "\n")
	sb := strings.Split(b, "\n")
	return cmp.Diff(sa, sb)
}
