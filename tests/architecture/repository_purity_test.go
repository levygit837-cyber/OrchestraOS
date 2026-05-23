package architecture

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repositoryPurityRoots defines where to look for repository.go files.
var repositoryPurityRoots = []string{"../../internal/modules", "../../internal/core"}

// repositoryPurityPatterns detects business logic in repository.go files.
//
// Heuristics (per ADR-0019 Pilar 3: "repository.go é CRUD puro"):
//  1. Status-based branching: if statements that compare variables with Status* constants.
//  2. Deduplication logic: patterns like "if existing != nil" followed by comparison.
//  3. Reference/upsert detection: returning booleans indicating if record was created or referenced.
//  4. Hardcoded status strings: string literals like "active", "inactive", "running", etc.
//  5. Field validation: checking ID == "", Sequence == 0, CreatedAt.IsZero() (beyond simple nil check).
//  6. ON CONFLICT clauses in SQL strings.
type repositoryPurityPattern struct {
	name        string
	regex       *regexp.Regexp
	description string
}

var repositoryPurityPatterns = []repositoryPurityPattern{
	{
		name:        "status-branching",
		regex:       regexp.MustCompile(`(?i)if\s+.*Status[A-Z]`),
		description: "status-based conditional logic — repositories must not branch on business state",
	},
	{
		name:        "deduplication",
		regex:       regexp.MustCompile(`(?i)if\s+existing\s*!=?\s*nil`),
		description: "deduplication logic — repositories must not query existing records to decide insert/update",
	},
	{
		name:        "reference-detection",
		regex:       regexp.MustCompile(`(?i)return\s+.*!=\s*.*\|\|\s*.*>\s*1`),
		description: "reference/upsert detection logic — returning boolean indicating if record was created or referenced",
	},
	{
		name:        "hardcoded-status",
		regex:       regexp.MustCompile(`"(active|inactive|running|completed|failed|cancelled|pending|validated|ready)"`),
		description: "hardcoded status string — status values must be set by the service/state machine layer, not the repository",
	},
	{
		name:        "field-validation",
		regex:       regexp.MustCompile(`(?i)if\s+.*(Sequence\s*==?\s*0|CreatedAt\.IsZero\(\))`),
		description: "field validation logic — validating Sequence or CreatedAt is business logic, not CRUD",
	},
	{
		name:        "on-conflict",
		regex:       regexp.MustCompile(`(?i)ON\s+CONFLICT`),
		description: "ON CONFLICT clause — upsert/dedup logic belongs in the service layer",
	},
}

// TestRepositoryPurity verifies that repository.go files contain only CRUD
// operations and no business logic.
//
// Per ADR-0019 Pilar 3:
//
//	"repository.go é CRUD puro. Não computar timestamps baseados em status,
//	não fazer deduplicação, não fazer upsert logic."
func TestRepositoryPurity(t *testing.T) {
	for _, root := range repositoryPurityRoots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, "repository.go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			scanRepositoryForBusinessLogic(t, path)

			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", root, err)
		}
	}
}

func scanRepositoryForBusinessLogic(t *testing.T, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck // read-only file

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		for _, pattern := range repositoryPurityPatterns {
			if pattern.regex.MatchString(line) {
				t.Errorf(
					"repository business logic detected at %s:%d — %s (ADR-0019). "+
						"Line: %s",
					path, lineNum, pattern.description, trimmed,
				)
			}
		}
	}
}

// allowedRepositoryMethodPrefixes lists prefixes that indicate CRUD or
// database-helper methods. Methods not matching these prefixes are flagged.
var allowedRepositoryMethodPrefixes = []string{
	"Create", "Get", "List", "Update", "Delete", "Count", "Exists",
	"Save", "Insert", "Find", "Fetch", "Remove", "Query",
	"scan", // row scanning helpers
}

// TestRepositoryMethodNames verifies that repository.go files only contain
// methods with CRUD-appropriate names.
//
// Disallowed patterns (business logic methods):
//
//	Validate*, Check*, Process*, Transition*, Start*, Stop*, Execute*, etc.
func TestRepositoryMethodNames(t *testing.T) {
	for _, root := range repositoryPurityRoots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, "repository.go") {
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

			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if fn.Recv == nil {
					continue // not a method
				}
				name := fn.Name.Name

				allowed := false
				for _, prefix := range allowedRepositoryMethodPrefixes {
					if strings.HasPrefix(name, prefix) {
						allowed = true
						break
					}
				}
				if !allowed {
					pos := fset.Position(fn.Pos())
					t.Errorf(
						"repository %s contains non-CRUD method %q at %s — "+
							"repositories must only have CRUD methods (Create, Get, List, Update, Delete, scan*, etc.) (ADR-0019).",
						path, name, pos,
					)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", root, err)
		}
	}
}
