package form_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// self is this component (any major version of it). The other components below must stay unknown to it.
const self = "github.com/donseba/go-form"

// components are the sibling libraries an application combines. None of them
// may import or require another: each one stands on its own, and only the
// application knows them all.
var components = []string{
	"github.com/donseba/go-partial",
	"github.com/donseba/go-router",
	"github.com/donseba/go-form",
	"github.com/donseba/go-importmap",
	"github.com/donseba/go-translator",
}

// TestKnowsNoOtherComponent checks that no Go file and no go.mod in this
// repository refers to another component.
func TestKnowsNoOtherComponent(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			name := entry.Name()
			if path != "." && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor") {
				return filepath.SkipDir
			}

			return nil
		}

		switch {
		case entry.Name() == "go.mod":
			checkGoMod(t, path)
		case strings.HasSuffix(path, ".go"):
			checkImports(t, path)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func checkImports(t *testing.T, path string) {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Errorf("%s: %v", path, err)
		return
	}

	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}

		if other, ok := otherComponent(imported); ok {
			t.Errorf("%s imports %s: components must not know each other", path, other)
		}
	}
}

func checkGoMod(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s: %v", path, err)
		return
	}

	for _, field := range strings.Fields(string(data)) {
		if other, ok := otherComponent(field); ok {
			t.Errorf("%s refers to %s: components must not know each other", path, other)
		}
	}
}

// otherComponent reports the component a module or package path belongs to,
// when that is not this module.
func otherComponent(path string) (string, bool) {
	for _, component := range components {
		if component == self {
			continue
		}

		if path == component || strings.HasPrefix(path, component+"/") {
			return component, true
		}
	}

	return "", false
}
