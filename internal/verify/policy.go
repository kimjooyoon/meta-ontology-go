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

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
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

// LinePolicy defines deterministic source constraints used by CI.
type LinePolicy struct {
	MaxFileLines         int
	MaxFunctionLines     int
	MaxDirectDirectoryIn int
}

// DefaultLinePolicy returns the currently active repository constraints.
func DefaultLinePolicy() LinePolicy {
	return LinePolicy{
		MaxFileLines:         75,
		MaxFunctionLines:     75,
		MaxDirectDirectoryIn: 10,
	}
}

// CheckGoCaps checks the DAMP file limit and DRY function limit. If files is
// empty, all source files below root are discovered in lexical path order.
func CheckGoCaps(root string, files []string, maxFileLines, maxFunctionLines int) error {
	return CheckSourcePolicy(root, files, LinePolicy{
		MaxFileLines:     maxFileLines,
		MaxFunctionLines: maxFunctionLines,
	})
}

// CheckSourcePolicy checks all repository-wide source constraints from policy.
func CheckSourcePolicy(root string, files []string, policy LinePolicy) error {
	if policy.MaxFileLines <= 0 || policy.MaxFunctionLines <= 0 {
		return fmt.Errorf("source policy caps must be positive")
	}
	if len(files) == 0 {
		var err error
		files, err = discoverSourceFiles(root)
		if err != nil {
			return err
		}
	}
	violations := make([]Violation, 0)
	for _, path := range sortedUnique(files) {
		violations = append(violations, checkSourceFile(root, path, policy)...)
	}
	if policy.MaxDirectDirectoryIn > 0 {
		violations = append(violations, checkDirectoryLayout(root, policy.MaxDirectDirectoryIn)...)
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
	return fmt.Errorf("source policy failed:\n%s", strings.Join(lines, "\n"))
}

func checkSourceFile(root, path string, policy LinePolicy) []Violation {
	if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".gooo") {
		return nil
	}
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return []Violation{{Path: path, Rule: "read-gooo-file", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	if lines := lineCount(source); lines > policy.MaxFileLines {
		rule := "DAMP file lines"
		if strings.HasSuffix(path, ".gooo") {
			rule = "GOOO file lines"
		}
		violations = append(violations, Violation{Path: path, Rule: rule, Actual: lines, Limit: policy.MaxFileLines})
	}
	if strings.HasSuffix(path, ".gooo") {
		return violations
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
		if length := end - start + 1; length > policy.MaxFunctionLines {
			violations = append(violations, Violation{Path: path, Rule: "DRY function lines", Actual: length, Limit: policy.MaxFunctionLines, Detail: name})
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

func discoverSourceFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".go" && extension != ".gooo" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func checkDirectoryLayout(root string, maxDirectEntries int) []Violation {
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return []Violation{{Path: ".", Rule: "directory layout", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	for _, directory := range report.Directories {
		directEntries := directory.DirectFiles + directory.DirectFolders
		if directEntries > maxDirectEntries {
			violations = append(violations, Violation{
				Path:   directory.Path,
				Rule:   "directory direct entries",
				Actual: directEntries,
				Limit:  maxDirectEntries,
				Detail: "too many direct children",
			})
		}
		if directory.DirectFolders > 0 && directory.DirectFiles > 0 {
			violations = append(violations, Violation{
				Path:   directory.Path,
				Rule:   "directory mixed entries",
				Detail: "must contain either files or folders, not both",
			})
		}
	}
	return violations
}

func discoverGoFiles(root string) ([]string, error) {
	return discoverSourceFiles(root)
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
