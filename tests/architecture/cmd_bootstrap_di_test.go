package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCmdBootstrapDI verifies that cmd/ files do not instantiate repositories
// or services directly. All wiring/DI must go through internal/bootstrap/.
//
// Per ADR-0019:
//
//	"cmd/ deve usar bootstrap/ para wiring de dependências."
//	Instantiating NewRepository() or NewService() directly in cmd/ bypasses
//	the DI container and makes testing, mocking, and lifecycle management
//	impossible.
//
// Allowed in cmd/:
//   - Calling bootstrap.Initialize() or similar DI setup functions.
//   - Calling cmd-layer helpers that delegate to bootstrap.
//
// Prohibited in cmd/:
//   - Direct calls to module.NewRepository()
//   - Direct calls to module.NewService()
//   - Direct calls to any constructor that should be managed by bootstrap.
func TestCmdBootstrapDI(t *testing.T) {
	cmdDir := "../../cmd"
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatalf("cannot read cmd directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cmdPath := filepath.Join(cmdDir, entry.Name())

		err := filepath.Walk(cmdPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if perr != nil {
				return nil // skip unparseable
			}

			for _, imp := range f.Imports {
				impPath := strings.Trim(imp.Path.Value, `"`)
				// Check for direct imports of internal/modules/* in cmd/
				if strings.HasPrefix(impPath, "github.com/levygit837-cyber/OrchestraOS/internal/modules/") {
					modName := filepath.Base(impPath)
					t.Errorf(
						"cmd file %s imports module %q directly — "+
							"cmd/ must NOT import modules directly. Use internal/bootstrap/ for DI wiring (ADR-0019).",
						path, modName,
					)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", cmdPath, err)
		}
	}
}
