package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCoordinationRemoved verifies that:
//  1. internal/core/coordination directory no longer exists (ADR-0028).
//  2. No Go file imports the coordination package.
//  3. No markdown document still references coordination as a valid package,
//     except for historical ADR files and documentation.
func TestCoordinationRemoved(t *testing.T) {
	coordinationDir := "../../internal/core/coordination"
	if _, err := os.Stat(coordinationDir); !os.IsNotExist(err) {
		t.Errorf("internal/core/coordination directory still exists — it must be removed per ADR-0028. Its logic was distributed to owner modules.")
	}

	// Walk the entire repo looking for imports of the coordination package
	// or stale documentation references.
	repoRoot := "../.."
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip .git, vendor, build artifacts, and local-only directories
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" ||
				name == "plans" || name == ".ultraplan" || name == "temp" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check Go files for imports
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(bytes)
			if strings.Contains(content, `"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"`) {
				t.Errorf("Go file %s still imports the removed coordination package — update to use the owner module or internal/core/transition (ADR-0028).", path)
			}
		}

		// Check markdown/docs for stale references that present coordination as valid.
		// Skip ADR files (historical context) and architecture docs that discuss the removal.
		if strings.HasSuffix(path, ".md") {
			// Skip historical ADR files and analysis docs
			base := filepath.Base(path)
			dir := filepath.Dir(path)
			if strings.Contains(base, "adr") || strings.Contains(base, "ADR") ||
				strings.Contains(dir, "adr") || strings.Contains(dir, "analysis") ||
				strings.Contains(dir, "architecture") || strings.Contains(dir, "implementation") ||
				strings.Contains(dir, "templates") {
				return nil
			}
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(bytes)
			// Only flag if the reference presents coordination as CURRENTLY valid
			if strings.Contains(content, "internal/core/coordination") {
				// Check if it's explaining the removal (contains "removed", "deprecated", etc.)
				lower := strings.ToLower(content)
				if !strings.Contains(lower, "removed") && !strings.Contains(lower, "deprecated") &&
					!strings.Contains(lower, "moved to") && !strings.Contains(lower, "no longer") {
					t.Errorf("Markdown file %s still references internal/core/coordination as a valid package — update documentation to reflect ADR-0028 removal.", path)
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}
