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

// TestContractsSync verifies that CONTRACTS.md files stay in sync with code.
func TestContractsSync(t *testing.T) {
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

		contractsPath := filepath.Join(modPath, "CONTRACTS.md")
		contractsBytes, err := os.ReadFile(contractsPath)
		if err != nil {
			t.Errorf("module %q: cannot read CONTRACTS.md: %v", modName, err)
			continue
		}
		contractsText := string(contractsBytes)

		// Every module must have an Invariants section.
		if !strings.Contains(contractsText, "## Invariants") {
			t.Errorf("module %q: CONTRACTS.md missing '## Invariants' section", modName)
		}

		// Extract status constants from models.go.
		statuses := extractStatusConstants(t, modPath)

		if len(statuses) > 0 {
			// Module has statuses → must have State Machine section.
			if !strings.Contains(contractsText, "## State Machine") {
				t.Errorf("module %q: has status constants but CONTRACTS.md missing '## State Machine' section", modName)
			}

			// Each status should be mentioned in the CONTRACTS.md.
			for _, status := range statuses {
				if !strings.Contains(contractsText, status) {
					t.Errorf("module %q: status constant %q found in models.go but not mentioned in CONTRACTS.md", modName, status)
				}
			}
		}

		// Verify that allowed dependencies in README.md reflect actual imports.
		readmePath := filepath.Join(modPath, "README.md")
		readmeBytes, err := os.ReadFile(readmePath)
		if err != nil {
			t.Errorf("module %q: cannot read README.md: %v", modName, err)
			continue
		}
		readmeText := string(readmeBytes)

		actualImports := getModuleImports(t, modPath)
		for _, imported := range actualImports {
			if imported == modName {
				continue // self
			}
			if !strings.Contains(readmeText, imported) {
				t.Errorf("module %q imports %q but README.md does not mention it in Allowed Dependencies", modName, imported)
			}
		}
	}
}

// extractStatusConstants parses models.go and returns all constant **values**
// that look like status strings (e.g., "created", "running").
// It looks for Status* constants and extracts their underlying string value.
func extractStatusConstants(t *testing.T, modPath string) []string {
	modelsPath := filepath.Join(modPath, "models.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, modelsPath, nil, parser.AllErrors)
	if err != nil {
		// models.go may not exist in some modules; that's okay.
		return nil
	}

	var statuses []string
	ast.Inspect(f, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}
		for _, spec := range decl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "Status") {
					continue
				}
				if i < len(vs.Values) {
					if bl, ok := vs.Values[i].(*ast.BasicLit); ok && bl.Kind == token.STRING {
						statuses = append(statuses, strings.Trim(bl.Value, `"`))
					}
				}
			}
		}
		return true
	})
	return statuses
}

// getModuleImports returns the names of other modules imported by the given module.
func getModuleImports(t *testing.T, modPath string) []string {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, modPath, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("cannot parse module %s: %v", modPath, err)
	}

	seen := make(map[string]bool)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
					mod := filepath.Base(path)
					if !seen[mod] {
						seen[mod] = true
					}
				}
			}
		}
	}

	var result []string
	for mod := range seen {
		result = append(result, mod)
	}
	return result
}
