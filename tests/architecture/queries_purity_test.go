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

// TestQueriesPurity verifies that every queries.go file in the project
// contains ONLY string constants/variables and package declaration.
// No functions, no structs, no non-stdlib imports, no logic.
func TestQueriesPurity(t *testing.T) {
	rootDirs := []string{"../../internal/modules", "../../internal/core"}

	for _, root := range rootDirs {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() != "queries.go" {
				return nil
			}

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if err != nil {
				t.Errorf("cannot parse %s: %v", path, err)
				return nil
			}

			// Check imports: only stdlib allowed
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.Contains(path, ".") {
					t.Errorf("%s imports non-stdlib package %q — queries.go must only contain SQL strings", path, imp.Path.Value)
				}
			}

			// Check top-level declarations
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					// const/var declarations are allowed
					if d.Tok != token.CONST && d.Tok != token.VAR {
						t.Errorf("%s contains %q declaration — queries.go must only have const/var strings", fset.Position(d.Pos()).String(), d.Tok.String())
					}
				case *ast.FuncDecl:
					t.Errorf("%s contains function %q — queries.go must not contain functions", fset.Position(d.Pos()).String(), d.Name.Name)
				default:
					t.Errorf("%s contains unexpected declaration at %s — queries.go must only have const/var strings", fset.Position(d.Pos()).String(), fset.Position(d.Pos()).String())
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", root, err)
		}
	}
}
