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
// These imports are typically for repository interfaces, event type helpers,
// or DI reader factories — never for service logic.
var allowedModuleImports = map[string]map[string]bool{
	"task":         {"run": true, "workunit": true},
	"run":          {"workunit": true},
	"agentsession": {"run": true},
	"taskgraph":    {"task": true, "workunit": true},
	"prompt":       {"task": true, "workunit": true, "run": true, "agentsession": true},
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
						t.Errorf("module %q imports %q, which is not in the allowed list. If this import is legitimate, update allowedModuleImports in module_boundaries_test.go", modName, importedMod)
					}
				}
			}
		}
	}
}
