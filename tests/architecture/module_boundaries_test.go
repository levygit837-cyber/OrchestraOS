package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// allowedModuleImports define which cross-module imports are legitimate.
// Keys are the importing module; values are the imported modules.
// TODO[ADR-0022]: run imports task only for TaskReader DI interface (returns *task.Task).
// Remove when task types are fully decoupled or when TaskReader uses a local struct.
var allowedModuleImports = map[string]map[string]bool{
	"run": {"task": true},
}

// leafModules must not import any other module under internal/modules/.
var leafModules = map[string]bool{
	"agent": true,
}

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
					if !strings.HasPrefix(path, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
						continue
					}
					importedMod := filepath.Base(path)

					if leafModules[modName] {
						t.Errorf("leaf module %q must not import any other module, but imports %q", modName, importedMod)
						continue
					}

					if modName == importedMod {
						continue // self-import (should not happen, but harmless)
					}

					allowed, ok := allowedModuleImports[modName]
					if !ok || !allowed[importedMod] {
						t.Errorf("module %q imports %q, which is not in the allowed list. Cross-module imports are forbidden. Use internal/core/coordination/ or internal/services/ for cross-module coordination", modName, importedMod)
					}
				}
			}
		}
	}
}

func TestModulesDoNotImportCoordination(t *testing.T) {
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
					if path == "github.com/levygit837-cyber/OrchestraOS/internal/core/coordination" {
						// The orchestrator module is allowed to import coordination as it is the
						// central coordinator that consumes RuntimeEventRelay and PromptOrchestrator.
						if modName != "orchestrator" {
							t.Errorf("module %q imports internal/core/coordination. Modules must not import coordination directly. Use internal/core/transition/ for shared types.", modName)
						}
					}
				}
			}
		}
	}
}
