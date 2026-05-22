package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sharedEntityTypes lists the entity types that MUST be defined in
// internal/domain/ per ADR-0030 Pilar 1.
//
// These types are shared across multiple modules and must not be duplicated
// or defined inside individual modules.
var sharedEntityTypes = []string{
	// Task domain
	"Task",
	"TaskStatus",
	"TaskPriority",
	"TaskRiskLevel",
	// Run domain
	"Run",
	"RunStatus",
	"RunResult",
	// WorkUnit domain
	"WorkUnit",
	"WorkUnitStatus",
	// Agent domain
	"Agent",
	"AgentRuntimeType",
	"AgentStatus",
	// AgentSession domain
	"AgentSession",
	"AgentSessionStatus",
	// TaskGraph domain
	"TaskGraph",
	"TaskGraphStatus",
	// Prompt domain
	"PromptFragment",
	"PromptSnapshot",
	"ToolsetSnapshot",
	"ComposedPrompt",
	// Review domain
	"Review",
	"ReviewStatus",
	"ReviewValidationGate",
	// Trigger domain
	"Trigger",
	"TriggerStatus",
	"TriggerType",
}

// TestDomainImportIntegrity verifies that:
//  1. All shared entity types are defined in internal/domain/.
//  2. Modules import internal/domain/ to use these types (not define their own).
//
// Per ADR-0030 Pilar 1:
//   "internal/domain/ centraliza TODOS os tipos compartilhados."
//   Modules must NOT define their own versions of these types.
func TestDomainImportIntegrity(t *testing.T) {
	// Phase 1: Verify all shared types exist in internal/domain/
	domainDir := "../../internal/domain"
	foundInDomain := make(map[string]bool)

	entries, err := os.ReadDir(domainDir)
	if err != nil {
		t.Fatalf("cannot read domain directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(domainDir, entry.Name())
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			continue // skip unparseable files
		}

		ast.Inspect(f, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				return true
			}
			for _, spec := range decl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				for _, entityType := range sharedEntityTypes {
					if ts.Name.Name == entityType {
						foundInDomain[entityType] = true
					}
				}
			}
			return true
		})
	}

	// Report missing types
	var missing []string
	for _, entityType := range sharedEntityTypes {
		if !foundInDomain[entityType] {
			missing = append(missing, entityType)
		}
	}
	if len(missing) > 0 {
		t.Errorf(
			"%d shared entity type(s) missing from internal/domain/ — they must be defined there per ADR-0030: %v",
			len(missing), missing,
		)
	}

	// Phase 2: Verify modules import internal/domain/
	// (This will be enforced once types are migrated; for now we check presence)
	modulesDir := "../../internal/modules"
	modEntries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatalf("cannot read modules directory: %v", err)
	}

	for _, entry := range modEntries {
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
			continue
		}

		importsDomain := false
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)
					if path == "github.com/levygit837-cyber/OrchestraOS/internal/domain" {
						importsDomain = true
					}
				}
			}
		}

		if !importsDomain {
			// Only flag if the module has models (it likely needs domain types)
			modelsPath := filepath.Join(modPath, "models.go")
			if _, err := os.Stat(modelsPath); err == nil {
				t.Logf("module %q does not import internal/domain/ — it should import domain types once migrated (ADR-0030)", modName)
			}
		}
	}
}
