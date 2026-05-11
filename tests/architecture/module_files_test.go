package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

// requiredModuleFiles lists files that every module directory must contain.
var requiredModuleFiles = []string{
	"doc.go",
	"README.md",
	"CONTRACTS.md",
	"contract.go",
}

// modulesWithoutQueries are modules that do not interact with the database directly
// and therefore do not need a queries.go file.
var modulesWithoutQueries = map[string]bool{
	"agent": true,
	"event": true,
}

func TestModuleRequiredFiles(t *testing.T) {
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

		for _, file := range requiredModuleFiles {
			fullPath := filepath.Join(modPath, file)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("module %q is missing required file %q", modName, file)
			}
		}
		if !modulesWithoutQueries[modName] {
			queriesPath := filepath.Join(modPath, "queries.go")
			if _, err := os.Stat(queriesPath); os.IsNotExist(err) {
				t.Errorf("module %q is missing queries.go (add it or exempt the module in modulesWithoutQueries)", modName)
			}
		}
	}
}

// requiredCoreFiles lists files that every core package directory must contain.
var requiredCoreFiles = []string{
	"doc.go",
}

func TestCoreRequiredFiles(t *testing.T) {
	coreDir := "../../internal/core"
	entries, err := os.ReadDir(coreDir)
	if err != nil {
		t.Fatalf("cannot read core directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()
		pkgPath := filepath.Join(coreDir, pkgName)

		for _, file := range requiredCoreFiles {
			fullPath := filepath.Join(pkgPath, file)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("core package %q is missing required file %q", pkgName, file)
			}
		}
	}
}
