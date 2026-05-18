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

// migratedEntityTypes lists structs that were defined in internal/domain
// but per ADR-0022 must live in their respective modules.
// Presence of these in internal/domain/ is an architectural anomaly.
var migratedEntityTypes = []string{
	"Task",
	"WorkUnit",
	"Run",
	"Agent",
	"AgentSession",
	"TaskGraph",
	"PromptFragment",
	"PromptSnapshot",
	"ToolsetSnapshot",
	"Trigger",
	"Review",
}

// TestDomainPurity verifies that internal/domain/ does not contain entity structs
// that should have been migrated to their respective vertical modules.
//
// Per ADR-0022 Section 4 (Pilar 1):
//
//	"internal/domain is a package of infrastructure, not of entities."
//	It should only contain genuinely shared types: EventEnvelope, EventPriority,
//	checkpoint types, and generic event payloads.
func TestDomainPurity(t *testing.T) {
	domainDir := "../../internal/domain"

	entries, err := os.ReadDir(domainDir)
	if err != nil {
		t.Fatalf("cannot read domain directory: %v", err)
	}

	var foundTypes []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(domainDir, entry.Name())
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			// Skip unparseable files (e.g., platform-specific build tags)
			continue
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
				_, isStruct := ts.Type.(*ast.StructType)
				if !isStruct {
					continue
				}
				for _, migrated := range migratedEntityTypes {
					if ts.Name.Name == migrated {
						pos := fset.Position(ts.Pos())
						t.Logf("[MIGRATION VIOLATION] entity struct %q found in %s at %s — this type must live in its respective module (ADR-0022).", migrated, path, pos)
						foundTypes = append(foundTypes, migrated)
					}
				}
			}
			return true
		})
	}

	if len(foundTypes) > 0 {
		t.Errorf("%d entity struct(s) must be migrated from internal/domain/ to their respective modules", len(foundTypes))
	}
}
