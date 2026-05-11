package architecture

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestModuleContract verifies that every module has a contract.go that acts as
// a gateway to README.md and CONTRACTS.md. The contract.go must NOT contain
// actual rules (that was the failure mode we observed: agents read contract.go
// and ignored the markdown files because the Go file was "good enough").
func TestModuleContract(t *testing.T) {
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

		contractPath := filepath.Join(modPath, "contract.go")
		contractBytes, err := os.ReadFile(contractPath)
		if err != nil {
			t.Errorf("module %q: missing contract.go: %v", modName, err)
			continue
		}
		contractText := string(contractBytes)

		// Must declare ModuleContract variable.
		if !strings.Contains(contractText, "var ModuleContract") {
			t.Errorf("module %q: contract.go must declare 'var ModuleContract'", modName)
		}

		// Name must match directory name.
		nameRe := regexp.MustCompile(`Name:\s*"([^"]+)"`)
		nameMatch := nameRe.FindStringSubmatch(contractText)
		if len(nameMatch) < 2 {
			t.Errorf("module %q: contract.go missing Name field", modName)
		} else if nameMatch[1] != modName {
			t.Errorf("module %q: contract.go Name=%q does not match directory name", modName, nameMatch[1])
		}

		// Must embed README.md and CONTRACTS.md.
		if !strings.Contains(contractText, "//go:embed README.md") {
			t.Errorf("module %q: contract.go must contain '//go:embed README.md'", modName)
		}
		if !strings.Contains(contractText, "//go:embed CONTRACTS.md") {
			t.Errorf("module %q: contract.go must contain '//go:embed CONTRACTS.md'", modName)
		}

		// Must contain a CRITICAL RULES header so LLM agents see the cheat sheet.
		if !strings.Contains(contractText, "CRITICAL RULES") {
			t.Errorf("module %q: contract.go must contain 'CRITICAL RULES' header with the essential rules", modName)
		}

		// Must contain an explicit instruction to read the markdown files.
		if !strings.Contains(contractText, "README.md") || !strings.Contains(contractText, "CONTRACTS.md") {
			t.Errorf("module %q: contract.go must reference both README.md and CONTRACTS.md", modName)
		}
	}
}
