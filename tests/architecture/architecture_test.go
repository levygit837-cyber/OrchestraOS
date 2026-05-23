package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const projectRoot = "../../"

// allowedImports defines the permitted internal import graph.
// Key = package path suffix, Value = allowed internal import suffixes.
var allowedImports = map[string][]string{
	"internal/domain":            {},
	"internal/apperrors":         {},
	"internal/store":             {"internal/domain", "internal/apperrors"},
	"internal/event":             {"internal/domain", "internal/store"},
	"internal/planner":           {"internal/domain", "internal/apperrors"},
	"internal/executor":          {"internal/domain", "internal/apperrors", "internal/store"},
	"internal/sse":               {},
	"internal/retry":             {"internal/apperrors"},
	"internal/runtime":           {"internal/domain", "internal/apperrors"},
	"internal/provider/gemini":   {"internal/domain", "internal/apperrors", "internal/runtime", "internal/sse"},
	"internal/provider/deepseek": {"internal/domain", "internal/apperrors", "internal/runtime", "internal/sse"},
	"internal/daggen":            {"internal/domain", "internal/apperrors"},
	"internal/decomposer":        {"internal/domain", "internal/apperrors", "internal/daggen", "internal/retry"},
	"internal/assignment":         {"internal/domain", "internal/apperrors"},
	"internal":                   {"internal/domain", "internal/executor", "internal/planner", "internal/store"},
	"cmd/orchestraos":            {"internal", "internal/domain", "internal/executor", "internal/planner", "internal/runtime", "internal/store", "internal/provider/gemini", "internal/provider/deepseek"},
}

const modulePath = "github.com/levygit837-cyber/OrchestraOS/"

func TestDependencyDirection(t *testing.T) {
	t.Parallel()

	for pkgSuffix, allowed := range allowedImports {
		pkgDir := filepath.Join(projectRoot, pkgSuffix)
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			continue
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, pkgDir, goFileFilter, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", pkgSuffix, err)
		}

		allowedSet := map[string]bool{}
		for _, a := range allowed {
			allowedSet[modulePath+a] = true
		}

		for _, pkg := range pkgs {
			for filePath, f := range pkg.Files {
				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					if !strings.HasPrefix(importPath, modulePath) {
						continue
					}
					if !allowedSet[importPath] {
						t.Errorf("%s imports %s — not allowed for %s",
							filepath.Base(filePath), importPath, pkgSuffix)
					}
				}
			}
		}
	}
}

func TestDomainPurity(t *testing.T) {
	t.Parallel()

	domainDir := filepath.Join(projectRoot, "internal/domain")
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, domainDir, goFileFilter, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse domain: %v", err)
	}

	for _, pkg := range pkgs {
		for filePath, f := range pkg.Files {
			for _, imp := range f.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(importPath, modulePath) {
					t.Errorf("domain/%s imports internal package %s — domain must be pure",
						filepath.Base(filePath), importPath)
				}
			}
		}
	}
}

func TestPackageSizeLimit(t *testing.T) {
	t.Parallel()

	const maxLines = 800

	packages := []string{
		"internal/domain",
		"internal/apperrors",
		"internal/store",
		"internal/event",
		"internal/planner",
		"internal/executor",
		"internal/sse",
		"internal/retry",
		"internal/runtime",
		"internal/provider/gemini",
		"internal/provider/deepseek",
		"internal/daggen",
		"internal/decomposer",
		"internal/assignment",
	}

	for _, pkg := range packages {
		pkgDir := filepath.Join(projectRoot, pkg)
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			continue
		}

		totalLines := 0
		err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			totalLines += strings.Count(string(data), "\n") + 1
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", pkg, err)
		}

		if totalLines > maxLines {
			t.Errorf("package %s has %d lines (max %d)", pkg, totalLines, maxLines)
		}
	}
}

func TestMaxFunctionComplexity(t *testing.T) {
	t.Parallel()

	const maxFuncLines = 40

	err := filepath.Walk(filepath.Join(projectRoot, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			start := fset.Position(fn.Body.Lbrace)
			end := fset.Position(fn.Body.Rbrace)
			lines := end.Line - start.Line
			if lines > maxFuncLines {
				t.Errorf("%s:%d func %s has %d lines (max %d)",
					filepath.Base(path), start.Line, fn.Name.Name, lines, maxFuncLines)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func TestSQLConfinement(t *testing.T) {
	t.Parallel()

	sqlPatterns := []string{
		"SELECT ", "INSERT ", "UPDATE ", "DELETE ",
		"CREATE TABLE", "ALTER TABLE", "DROP TABLE",
		"sql.Open", "sql.DB", "pgx.", "database/sql",
	}

	err := filepath.Walk(filepath.Join(projectRoot, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}

		rel, _ := filepath.Rel(filepath.Join(projectRoot, "internal"), path)
		if strings.HasPrefix(rel, "store") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, pattern := range sqlPatterns {
			if strings.Contains(content, pattern) {
				t.Errorf("SQL pattern %q found in %s — SQL must be confined to store/",
					pattern, rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func TestNoGlobalState(t *testing.T) {
	t.Parallel()

	mutableTypes := map[string]bool{
		"map":       true,
		"chan":      true,
		"sync.":     true,
		"[]":        true,
		"*":         true,
		"atomic.":   true,
		"sync.Once": true,
	}

	err := filepath.Walk(filepath.Join(projectRoot, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				typeStr := typeString(vs.Type)
				for typePat := range mutableTypes {
					if strings.Contains(typeStr, typePat) {
						pos := fset.Position(vs.Pos())
						t.Errorf("%s:%d global mutable var %s (type contains %q) — avoid global state",
							filepath.Base(path), pos.Line, vs.Names[0].Name, typePat)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func typeString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.ChanType:
		return "chan " + typeString(t.Value)
	default:
		return ""
	}
}

func goFileFilter(fi os.FileInfo) bool {
	return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
}
