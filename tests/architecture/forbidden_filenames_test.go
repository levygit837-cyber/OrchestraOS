package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenNames lists file names that are strictly prohibited anywhere in
// internal/modules/* and internal/core/* per ADR-0028 Section 2.2.2.
// These names communicate nothing about content and become dumping grounds.
var forbiddenNames = []string{
	"helpers.go",
	"utils.go",
	"common.go",
	"base.go",
	"misc.go",
	"kit.go",
	"txkit.go",
	"ops.go",
	"eventops.go",
	"stuff.go",
	"things.go",
}

// TestForbiddenFilenames verifies that no forbidden filename exists under
// internal/modules/* or internal/core/*.
func TestForbiddenFilenames(t *testing.T) {
	roots := []string{"../../internal/modules", "../../internal/core"}

	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := info.Name()
			for _, forbidden := range forbiddenNames {
				if strings.EqualFold(name, forbidden) {
					t.Errorf("forbidden filename %q found at %s — %q is prohibited by ADR-0028. Use a descriptive name that communicates what the file does.", name, path, forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", root, err)
		}
	}
}
