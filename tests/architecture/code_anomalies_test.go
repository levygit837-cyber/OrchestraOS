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

// sqlPattern detects raw SQL strings outside queries.go files.
// It requires multiple SQL keywords in sequence to avoid false positives
// from error messages like "failed to update heartbeat".
//
// Updated per ADR-0019: also detects SELECT func(...) patterns (SQL without FROM).
var sqlPattern = regexp.MustCompile(`(?i)(SELECT\s+.*\s+FROM|SELECT\s+\w+\s*\(|INSERT\s+INTO|UPDATE\s+.*\s+SET|DELETE\s+FROM|CREATE\s+TABLE|ALTER\s+TABLE|DROP\s+TABLE|JOIN\s+.*\s+ON)`)

// safeToIgnoreCalls lists function names where ignoring the error is common
// and safe (e.g., SetDeadline, Close on read-only resources).
var safeToIgnoreCalls = []string{
	"Close",
	"SetDeadline",
	"SetReadDeadline",
	"SetWriteDeadline",
	"Write",
	"Sync",
}

// TestCodeAnomalies detects anti-patterns in the codebase:
//   - panic() calls (except in cmd/ and tests)
//   - fmt.Println / fmt.Printf (except in cmd/)
//   - SQL strings outside queries.go
//   - Error ignores without a documented reason comment
func TestCodeAnomalies(t *testing.T) {
	roots := []string{"../../internal/modules", "../../internal/core"}

	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
				return nil // tests are allowed more latitude
			}

			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if perr != nil {
				// Not a parseable Go file (e.g., build tags that exclude current platform)
				return nil
			}

			// Check for panic() calls
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				ident, ok := call.Fun.(*ast.Ident)
				if ok && ident.Name == "panic" {
					pos := fset.Position(call.Pos())
					t.Errorf("panic() call at %s — modules and core must NEVER call panic(). Return apperrors.Error instead (CODING_STANDARDS.md).", pos)
				}
				return true
			})

			// Check for fmt.Println / fmt.Printf
			for _, imp := range f.Imports {
				impPath := strings.Trim(imp.Path.Value, `"`)
				if impPath == "fmt" {
					scanFmtCalls(t, path)
				}
			}

			// Check for SQL strings outside queries.go
			if filepath.Base(path) != "queries.go" {
				scanForInlineSQL(t, path)
			}

			// Check for _ = someCall() without documented reason
			scanForIgnoredErrors(t, fset, f, path)

			// Check for _ = variable (not just calls) without documented reason
			scanForIgnoredVariables(t, fset, f, path)

			// Check for _, _ = someCall() without documented reason
			scanForFullyIgnoredTuple(t, fset, f, path)

			// Check for _ = call() inside defer func() without documented reason
			scanForIgnoredErrorsInDefer(t, fset, f, path)

			return nil
		})
		if err != nil {
			t.Fatalf("walk error in %s: %v", root, err)
		}
	}
}

func scanFmtCalls(t *testing.T, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck // read-only file, safe to ignore

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, "fmt.Println") || strings.Contains(line, "fmt.Printf") {
			t.Errorf("fmt.Println / fmt.Printf found at %s:%d — modules and core must use structured logging or return errors, not stdout (CODING_STANDARDS.md).", path, lineNum)
		}
	}
}

func scanForInlineSQL(t *testing.T, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck // read-only file, safe to ignore

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
		// Skip error message strings — they often contain SQL-like words
		if strings.Contains(trimmed, `fmt.Errorf(`) || strings.Contains(trimmed, `errors.New(`) {
			continue
		}
		// Detect actual SQL patterns inside string literals
		if sqlPattern.MatchString(line) {
			if strings.Contains(line, "`") || strings.Contains(line, `"`) {
				t.Errorf("possible inline SQL at %s:%d — SQL must live in queries.go only (ADR-0015, CODING_STANDARDS.md). Line: %s", path, lineNum, trimmed)
			}
		}
	}
}

func scanForIgnoredErrors(t *testing.T, fset *token.FileSet, f *ast.File, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if ALL LHS are "_" — only then is it a fully ignored error.
		// Patterns like "_, err := call()" or "val, _ := call()" are NOT fully ignored.
		allBlank := true
		for _, lhs := range assign.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if !ok || ident.Name != "_" {
				allBlank = false
				break
			}
		}
		if !allBlank {
			return true
		}

		// Check if ALL RHS are calls (skip _ = non-call like _ = ctx)
		allCalls := true
		for _, rhs := range assign.Rhs {
			_, isCall := rhs.(*ast.CallExpr)
			if !isCall {
				allCalls = false
				break
			}
		}
		if !allCalls {
			return true
		}

		// Check if any RHS call is a known safe-to-ignore function
		for _, rhs := range assign.Rhs {
			if isSafeToIgnoreCall(rhs) {
				return true
			}
		}

		pos := fset.Position(assign.Pos())
		if !hasIgnoreComment(fset, f, pos.Line) && !hasInlineIgnoreComment(path, pos.Line) {
			t.Errorf("ignored error (_ = ...) at %s:%d without documented reason — add a comment explaining why the error is safe to ignore (CODING_STANDARDS.md).", path, pos.Line)
		}
		return true
	})
}

// scanForIgnoredVariables detects `_ = variable` assignments where the error
// (or other return value) is silently discarded without documentation.
//
// Per ADR-0019 / CODING_STANDARDS.md: all ignored values must be documented.
func scanForIgnoredVariables(t *testing.T, fset *token.FileSet, f *ast.File, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if ALL LHS are "_"
		allBlank := true
		for _, lhs := range assign.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if !ok || ident.Name != "_" {
				allBlank = false
				break
			}
		}
		if !allBlank {
			return true
		}

		// Check if ANY RHS is NOT a call (i.e., it's a variable, selector, etc.)
		hasNonCall := false
		for _, rhs := range assign.Rhs {
			if _, isCall := rhs.(*ast.CallExpr); !isCall {
				hasNonCall = true
				break
			}
		}
		if !hasNonCall {
			return true // handled by scanForIgnoredErrors
		}

		pos := fset.Position(assign.Pos())
		if !hasIgnoreComment(fset, f, pos.Line) && !hasInlineIgnoreComment(path, pos.Line) {
			t.Errorf("ignored value (_ = ...) at %s:%d without documented reason — add a comment explaining why it is safe to ignore (CODING_STANDARDS.md).", path, pos.Line)
		}
		return true
	})
}

// scanForFullyIgnoredTuple detects `_, _ = someCall()` where both return values
// are silently discarded without documentation.
//
// Per ADR-0019 / CODING_STANDARDS.md: all ignored values must be documented.
func scanForFullyIgnoredTuple(t *testing.T, fset *token.FileSet, f *ast.File, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Must have exactly 2 LHS
		if len(assign.Lhs) != 2 {
			return true
		}

		// Both LHS must be "_"
		lhs1, ok1 := assign.Lhs[0].(*ast.Ident)
		lhs2, ok2 := assign.Lhs[1].(*ast.Ident)
		if !ok1 || !ok2 || lhs1.Name != "_" || lhs2.Name != "_" {
			return true
		}

		// Check if any RHS call is a known safe-to-ignore function
		for _, rhs := range assign.Rhs {
			if isSafeToIgnoreCall(rhs) {
				return true
			}
		}

		pos := fset.Position(assign.Pos())
		if !hasIgnoreComment(fset, f, pos.Line) && !hasInlineIgnoreComment(path, pos.Line) {
			t.Errorf("ignored tuple (_, _ = ...) at %s:%d without documented reason — add a comment explaining why both values are safe to ignore (CODING_STANDARDS.md).", path, pos.Line)
		}
		return true
	})
}

// scanForIgnoredErrorsInDefer detects `_ = call()` inside defer func() { ... }()
// where errors are silently discarded. This is a common anti-pattern because
// deferred errors are often important (e.g., rollback, close, cleanup).
//
// Per ADR-0019 / CODING_STANDARDS.md: all ignored errors must be documented,
// especially in defer blocks where failures may indicate resource leaks.
func scanForIgnoredErrorsInDefer(t *testing.T, fset *token.FileSet, f *ast.File, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		deferStmt, ok := n.(*ast.DeferStmt)
		if !ok {
			return true
		}

		// Only check defer func() { ... }()
		funcLit, ok := deferStmt.Call.Fun.(*ast.FuncLit)
		if !ok {
			return true
		}

		// Walk the deferred function body
		ast.Inspect(funcLit.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}

			// Check if ALL LHS are "_"
			allBlank := true
			for _, lhs := range assign.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name != "_" {
					allBlank = false
					break
				}
			}
			if !allBlank {
				return true
			}

			// Check if ALL RHS are calls
			allCalls := true
			for _, rhs := range assign.Rhs {
				if _, isCall := rhs.(*ast.CallExpr); !isCall {
					allCalls = false
					break
				}
			}
			if !allCalls {
				return true
			}

			// Skip safe-to-ignore calls
			for _, rhs := range assign.Rhs {
				if isSafeToIgnoreCall(rhs) {
					return true
				}
			}

			pos := fset.Position(assign.Pos())
			if !hasIgnoreComment(fset, f, pos.Line) && !hasInlineIgnoreComment(path, pos.Line) {
				t.Errorf("ignored error in defer block at %s:%d — errors in defer are often critical (rollback, close, cleanup). Document why safe to ignore or handle the error (CODING_STANDARDS.md).", path, pos.Line)
			}
			return true
		})

		return true
	})
}

// isSafeToIgnoreCall checks if the RHS is a call to a function known to be safe to ignore.
func isSafeToIgnoreCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	var name string
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		name = fn.Name
	case *ast.SelectorExpr:
		name = fn.Sel.Name
	}
	for _, safe := range safeToIgnoreCalls {
		if name == safe {
			return true
		}
	}
	return false
}

// hasInlineIgnoreComment reads the source file line and checks for inline ignore comments.
func hasInlineIgnoreComment(path string, lineNum int) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close() //nolint:errcheck // read-only file, safe to ignore

	scanner := bufio.NewScanner(file)
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		if currentLine == lineNum {
			text := strings.ToLower(scanner.Text())
			return strings.Contains(text, "// ignore:") || strings.Contains(text, "//nolint:")
		}
	}
	return false
}

// hasIgnoreComment checks if there is a comment on the given line or the line before,
// or an inline comment on the assignment line itself.
func hasIgnoreComment(fset *token.FileSet, f *ast.File, line int) bool {
	// Check AST comments (previous line or same line via AST association)
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			cLine := fset.Position(c.Pos()).Line
			if cLine == line || cLine == line-1 {
				text := strings.ToLower(c.Text)
				if strings.Contains(text, "ignore") || strings.Contains(text, "nolint") || strings.Contains(text, "safe to ignore") || strings.Contains(text, "cannot fail") || strings.Contains(text, "intentionally") {
					return true
				}
			}
		}
	}
	return false
}
