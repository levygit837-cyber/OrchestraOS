package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// excludedFromModuleBoundaryCheck lists packages that are allowed to import
// multiple modules. These are coordination/orchestration layers, not business
// logic modules.
//
// Per ADR-0030 Pilar 2:
//   - Only orchestrator/ and bootstrap/ may import multiple modules.
//   - All other modules MUST be independent.
var excludedFromModuleBoundaryCheck = map[string]bool{
	"orchestrator": true,
}

// TestModuleBoundaries verifies that no module under internal/modules/
// imports another module, with the sole exceptions of orchestrator/ and
// bootstrap/ (which act as composition roots).
//
// Per ADR-0030:
//
//	"Módulos NÃO importam outros módulos. Ponto final."
//	Any cross-module dependency must be resolved through the orchestrator
//	or through shared types in internal/domain/.
func TestModuleBoundaries(t *testing.T) {
	modulesDir := "../../internal/modules"
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatalf("cannot read modules directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modName := entry.Name()

		// orchestrator and bootstrap are exempt — they compose modules
		if excludedFromModuleBoundaryCheck[modName] {
			continue
		}

		modPath := filepath.Join(modulesDir, modName)
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, modPath, func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("cannot parse module %s: %v", modName, err)
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)

					// Check for cross-module imports
					if strings.HasPrefix(path, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
						importedMod := filepath.Base(path)
						if importedMod == modName {
							continue // self-import (should not happen, but harmless)
						}
						t.Errorf(
							"module %q imports %q — modules must NOT import other modules (ADR-0030). "+
								"Only orchestrator/ and bootstrap/ may import multiple modules. "+
								"Move shared types to internal/domain/ or resolve the dependency in the orchestrator.",
							modName, importedMod,
						)
					}

					// Check for legacy coordination package imports
					if path == "github.com/levygit837-cyber/OrchestraOS/internal/core/coordination" {
						t.Errorf(
							"module %q imports internal/core/coordination — this package was removed per ADR-0028. "+
								"Use internal/core/transition/ for shared types.",
							modName,
						)
					}
				}
			}
		}
	}
}
