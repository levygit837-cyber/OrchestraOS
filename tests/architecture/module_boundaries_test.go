package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// allowedModuleImports defines which cross-module type imports are legitimate
// for DI (Dependency Injection) interfaces.
//
// Policy (ADR-0026):
//   - Modules MAY import types (structs, enums) from another module ONLY for
//     DI interface return types (e.g., TaskReader.GetByID() -> *task.Task).
//   - Modules MUST NEVER import services, repositories, or business logic
//     from another module.
//   - This map tracks the ALLOWED compile-time dependencies.
//   - When a migration adds a new DI interface that returns a type from module B,
//     add the entry here and document it in the migration plan.
//
// Current allowed imports:
//
//	run -> task: run.TaskReader returns *task.Task
//	workunit -> task: workunit.TaskReader returns *task.Task
//	orchestrator -> review: orchestrator.ReviewManager returns *review.Review (ADR-0022 migration)
//	prompt -> run/task/workunit/agentsession: PrepareAndPersistInput uses *run.Run, *task.Task, *workunit.WorkUnit, *agentsession.AgentSession
//	orchestrator -> prompt: PreparedPrompt uses *prompt.PromptSnapshot and *prompt.ToolsetSnapshot
var allowedModuleImports = map[string]map[string]bool{
	"run":          {"task": true, "workunit": true},
	"workunit":     {"task": true, "taskgraph": true},
	"taskgraph":    {"task": true, "workunit": true},
	"agentsession": {"agent": true},
	"prompt":       {"run": true, "task": true, "workunit": true, "agentsession": true},
	"orchestrator": {"review": true, "taskgraph": true, "workunit": true, "trigger": true, "prompt": true},
	"trigger":      {"agentsession": true, "run": true, "workunit": true},
}

// leafModules must not import any other module under internal/modules/.
// These modules have no DI dependencies on other domain modules.
var leafModules = map[string]bool{
	"agent": true,
	"task":  true,
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
						t.Errorf(
							"module %q imports %q, which is not in the allowed list. "+
								"Cross-module imports are allowed ONLY for DI interface return types (ADR-0026). "+
								"If this import is for a DI interface type, add it to allowedModuleImports and document in the migration plan. "+
								"If this import is for a service/repository/business logic, refactor to use DI or move to core/coordination.",
							modName, importedMod,
						)
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
