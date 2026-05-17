package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestTransitionPackageIsLeaf(t *testing.T) {
	transitionDir := "../../internal/core/transition"
	entries, err := os.ReadDir(transitionDir)
	if err != nil {
		t.Fatalf("cannot read transition directory: %v", err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := transitionDir + "/" + entry.Name()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("cannot parse %s: %v", path, err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(path, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
				t.Errorf("internal/core/transition imports module %q. transition must not import any module.", path)
			}
			if path == "github.com/levygit837-cyber/OrchestraOS/internal/core/coordination" {
				t.Errorf("internal/core/transition imports internal/core/coordination. coordination was removed per ADR-0028.")
			}
		}
	}
}
