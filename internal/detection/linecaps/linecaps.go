// Package linecaps independently checks the repository's Go size policy.
//
// The checker measures physical source lines and AST source ranges. A file or
// function exactly at its configured limit passes; only values above a limit
// produce findings. Function literals are checked as well as named functions.
// It also emits lightweight refactorability findings for trivially small
// function bodies.
package linecaps

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

const (
	// DefaultMaxFileLines is the DAMP file cap used by this repository.
	DefaultMaxFileLines = 300
	// DefaultMaxFunctionLines is the DRY function cap used by this repository.
	DefaultMaxFunctionLines = 75
)

// Limits contains the inclusive maximum sizes accepted by Analyze.
type Limits struct {
	MaxFileLines     int
	MaxFunctionLines int
}

// DefaultLimits returns the repository's 300/75 line policy.
func DefaultLimits() Limits {
	return Limits{MaxFileLines: DefaultMaxFileLines, MaxFunctionLines: DefaultMaxFunctionLines}
}

func (l Limits) validate() error {
	if l.MaxFileLines <= 0 || l.MaxFunctionLines <= 0 {
		return fmt.Errorf("linecaps limits must be positive")
	}
	return nil
}

// Discover returns repository-relative Go paths in lexical order. .git and
// vendor directories are excluded so generated or vendored copies cannot
// silently expand the verification scope.
func Discover(root string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("linecaps root must not be empty")
	}
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// Analyze checks the supplied repository-relative Go paths. An empty paths
// slice discovers all Go files below root. I/O and parse failures are findings,
// allowing one invocation to report every independently unverifiable file.
func Analyze(root string, paths []string, limits Limits) (Report, error) {
	if err := limits.validate(); err != nil {
		return Report{}, err
	}
	if root == "" {
		return Report{}, fmt.Errorf("linecaps root must not be empty")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	if len(paths) == 0 {
		paths, err = Discover(absoluteRoot)
		if err != nil {
			return Report{}, err
		}
	} else {
		paths, err = normalizePaths(absoluteRoot, paths)
		if err != nil {
			return Report{}, err
		}
	}
	findings := make([]Finding, 0)
	for _, path := range paths {
		displayPath, fullPath, err := resolvePath(absoluteRoot, path)
		if err != nil {
			return Report{}, err
		}
		source, err := os.ReadFile(fullPath)
		if err != nil {
			findings = append(findings, Finding{Path: displayPath, Rule: RuleReadFile, Detail: err.Error()})
			continue
		}
		fileFindings, err := AnalyzeSource(displayPath, source, limits)
		if err != nil {
			return Report{}, err
		}
		findings = append(findings, fileFindings...)
	}
	sortFindings(findings)
	return Report{Findings: findings}, nil
}

// AnalyzeSource checks one Go source buffer. It is useful for callers that
// already own file contents and makes the line/range rules independently
// testable without a repository walk.
func AnalyzeSource(path string, source []byte, limits Limits) ([]Finding, error) {
	if err := limits.validate(); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("linecaps path must not be empty")
	}
	findings := make([]Finding, 0)
	if lines := lineCount(source); lines > limits.MaxFileLines {
		findings = append(findings, Finding{Path: path, Rule: RuleFileLines, Actual: lines, Limit: limits.MaxFileLines})
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		findings = append(findings, Finding{Path: path, Rule: RuleParseFile, Detail: err.Error()})
		sortFindings(findings)
		return findings, nil
	}
	ast.Inspect(file, func(node ast.Node) bool {
		finding, ok := functionFinding(fset, node, path, limits.MaxFunctionLines)
		if ok {
			findings = append(findings, finding)
		}
		refactorFinding, refactorOK := refactorCandidateFinding(fset, node, path)
		if refactorOK {
			findings = append(findings, refactorFinding)
		}
		return true
	})
	sortFindings(findings)
	return findings, nil
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

func functionFinding(fset *token.FileSet, node ast.Node, path string, limit int) (Finding, bool) {
	name, start, end, ok := functionSpan(fset, node)
	if !ok {
		return Finding{}, false
	}
	actual := end - start + 1
	if actual <= limit {
		return Finding{}, false
	}
	return Finding{Path: path, Rule: RuleFunctionLines, Name: name, StartLine: start, EndLine: end, Actual: actual, Limit: limit}, true
}

func functionSpan(fset *token.FileSet, node ast.Node) (name string, start int, end int, ok bool) {
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
		return name, fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, true
	case *ast.FuncLit:
		return "function literal", fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, true
	default:
		return "", 0, 0, false
	}
}

func normalizePaths(root string, paths []string) ([]string, error) {
	result := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		normalized, err := normalizePath(root, path)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result, nil
}

func normalizePath(root, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("linecaps path must not be empty")
	}
	normalized := strings.ReplaceAll(path, "\\", "/")
	cleaned := filepath.Clean(filepath.FromSlash(normalized))
	if filepath.IsAbs(cleaned) {
		relative, err := filepath.Rel(root, cleaned)
		if err != nil {
			return "", err
		}
		if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("linecaps path escapes root: %q", path)
		}
		return filepath.ToSlash(relative), nil
	}
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(cleaned), nil
}

func resolvePath(root, path string) (string, string, error) {
	if path == "" || path == "." {
		return "", "", fmt.Errorf("linecaps path must not be empty")
	}
	fullPath := filepath.FromSlash(path)
	if !filepath.IsAbs(fullPath) {
		return filepath.ToSlash(path), filepath.Join(root, fullPath), nil
	}
	relative, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(relative), fullPath, nil
}
