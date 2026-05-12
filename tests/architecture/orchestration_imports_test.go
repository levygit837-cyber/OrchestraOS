package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// allowedOrchestrationImporters lists packages that are allowed to import modules.
// All other packages under internal/ must not import internal/modules/* directly.
var allowedOrchestrationImporters = []string{
	"internal/core/orchestration",
	"internal/services",
	"internal/bootstrap",
	"cmd",
}

func TestOnlyOrchestrationImportsModules(t *testing.T) {
	internalDir := "../../internal"
	allowed := map[string]bool{}
	for _, p := range allowedOrchestrationImporters {
		allowed[p] = true
	}

	// Walk internal/ to find all Go packages
	var packages []string
	err := filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip modules directory — they have their own rules
			if strings.HasPrefix(path, filepath.Join(internalDir, "modules")) {
				return filepath.SkipDir
			}
			// Skip vendor-like dirs
			if strings.Contains(path, "/vendor/") || strings.Contains(path, "/.git/") {
				return filepath.SkipDir
			}
			packages = append(packages, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("cannot walk internal directory: %v", err)
	}

	for _, pkgPath := range packages {
		relPath, _ := filepath.Rel("../..", pkgPath)
		relPath = filepath.ToSlash(relPath)
		if allowed[relPath] || strings.HasPrefix(relPath, "cmd/") {
			continue
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, pkgPath, func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, parser.ImportsOnly)
		if err != nil {
			// Not a Go package — skip
			continue
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)
					if strings.HasPrefix(path, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
						t.Errorf("package %q imports module %q. Only %v may import modules directly.", relPath, path, allowedOrchestrationImporters)
					}
				}
			}
		}
	}
}
