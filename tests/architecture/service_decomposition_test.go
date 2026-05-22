package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestServiceDecomposition verifies that service_<sub>.go files only exist
// when the main service.go file exceeds 300 lines.
//
// Per ADR-0030 / CODING_STANDARDS.md:
//   "service_<sub>.go só é permitido se service.go tiver > 300 linhas."
//
// Rationale: Decomposition should be justified by file size. Small service.go
// files should not be split arbitrarily, as it fragments the business logic
// and makes the module harder to understand.
func TestServiceDecomposition(t *testing.T) {
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

		// Check if module has any service_<sub>.go files
		var subServiceFiles []string
		serviceGoPath := filepath.Join(modPath, "service.go")

		modEntries, err := os.ReadDir(modPath)
		if err != nil {
			continue
		}

		for _, file := range modEntries {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			if strings.HasPrefix(name, "service_") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
				subServiceFiles = append(subServiceFiles, name)
			}
		}

		if len(subServiceFiles) == 0 {
			continue // no sub-service files, no issue
		}

		// Count lines in service.go
		serviceGoBytes, err := os.ReadFile(serviceGoPath)
		if err != nil {
			// service.go doesn't exist but sub-service files do
			t.Errorf("module %q has %s but no service.go", modName, strings.Join(subServiceFiles, ", "))
			continue
		}

		lines := strings.Count(string(serviceGoBytes), "\n")
		if lines <= 300 {
			t.Errorf(
				"module %q has %s but service.go has only %d lines — "+
				"service_<sub>.go is only permitted when service.go has > 300 lines (CODING_STANDARDS.md). "+
				"Consider merging back into service.go or justifying the decomposition.",
				modName, strings.Join(subServiceFiles, ", "), lines,
			)
		}
	}
}
