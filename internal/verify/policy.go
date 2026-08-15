// Package verify contains deterministic repository and semantic conformance
// checks used by CI. It intentionally depends only on the standard library;
// semantic-specific tests in this package exercise the compiler boundaries.
package verify

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Violation is one deterministic policy failure.
type Violation struct {
	Path   string
	Rule   string
	Actual int
	Limit  int
	Detail string
}

func (v Violation) Error() string {
	if v.Detail != "" {
		return fmt.Sprintf("%s: %s: %s", v.Path, v.Rule, v.Detail)
	}
	return fmt.Sprintf("%s: %s: got %d, limit %d", v.Path, v.Rule, v.Actual, v.Limit)
}

// CheckGoCaps checks the DAMP file limit and DRY function limit. If files is
// empty, all Go files below root are discovered in lexical path order.
func CheckGoCaps(root string, files []string, maxFileLines, maxFunctionLines int) error {
	if maxFileLines <= 0 || maxFunctionLines <= 0 {
		return fmt.Errorf("Go caps must be positive")
	}
	if len(files) == 0 {
		var err error
		files, err = discoverGoFiles(root)
		if err != nil {
			return err
		}
	}
	violations := make([]Violation, 0)
	for _, path := range sortedUnique(files) {
		violations = append(violations, checkGoFile(root, path, maxFileLines, maxFunctionLines)...)
	}
	if len(violations) == 0 {
		return nil
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Path != violations[j].Path {
			return violations[i].Path < violations[j].Path
		}
		return violations[i].Rule < violations[j].Rule
	})
	lines := make([]string, len(violations))
	for i, violation := range violations {
		lines[i] = violation.Error()
	}
	return fmt.Errorf("Go size policy failed:\n%s", strings.Join(lines, "\n"))
}

func checkGoFile(root, path string, maxFileLines, maxFunctionLines int) []Violation {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return []Violation{{Path: path, Rule: "read-go-file", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	if lines := lineCount(source); lines > maxFileLines {
		violations = append(violations, Violation{Path: path, Rule: "DAMP file lines", Actual: lines, Limit: maxFileLines})
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return append(violations, Violation{Path: path, Rule: "parse-go-file", Detail: err.Error()})
	}
	ast.Inspect(file, func(node ast.Node) bool {
		start, end, name, ok := functionRange(fset, node)
		if !ok {
			return true
		}
		if length := end - start + 1; length > maxFunctionLines {
			violations = append(violations, Violation{Path: path, Rule: "DRY function lines", Actual: length, Limit: maxFunctionLines, Detail: name})
		}
		return true
	})
	return violations
}

func functionRange(fset *token.FileSet, node ast.Node) (int, int, string, bool) {
	var name string
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
	case *ast.FuncLit:
		name = "function literal"
	default:
		return 0, 0, "", false
	}
	return fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, name, true
}

func discoverGoFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(relative))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func lineCount(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}

// CheckPathScope rejects changed paths outside the verifier's ownership
// boundary. Paths are repository-relative and use slash separators.
func CheckPathScope(paths, allowedPrefixes []string) error {
	allowed := normalizePrefixes(allowedPrefixes)
	violations := make([]string, 0)
	for _, path := range sortedUnique(paths) {
		canonical := filepath.ToSlash(filepath.Clean(path))
		if path == "" || path != canonical || strings.Contains(path, "\\") || strings.HasPrefix(path, "/") {
			violations = append(violations, path)
			continue
		}
		if path == "." || isAllowed(path, allowed) {
			continue
		}
		violations = append(violations, path)
	}
	if len(violations) > 0 {
		return fmt.Errorf("changed paths outside CI ownership: %s", strings.Join(violations, ", "))
	}
	return nil
}

func normalizePrefixes(prefixes []string) []string {
	result := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		prefix = strings.Trim(strings.ReplaceAll(prefix, "\\", "/"), "/")
		if prefix != "" {
			result = append(result, prefix)
		}
	}
	return sortedUnique(result)
}

func isAllowed(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

// CheckPullRequestPolicy enforces the steady-state branch policy used by CI.
func CheckPullRequestPolicy(head, base string) error {
	if base == "main" {
		if head != "dev" {
			return fmt.Errorf("main promotion head must be dev, got %q", head)
		}
		return nil
	}
	if base != "dev" {
		return fmt.Errorf("feature pull request base must be dev, got %q", base)
	}
	if !strings.HasPrefix(head, "agent/") || len(strings.TrimPrefix(head, "agent/")) == 0 {
		return fmt.Errorf("feature pull request head must use agent/*, got %q", head)
	}
	return nil
}

func sortedUnique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	if len(result) < 2 {
		return result
	}
	write := 1
	for _, value := range result[1:] {
		if value == result[write-1] {
			continue
		}
		result[write] = value
		write++
	}
	return result[:write]
}
