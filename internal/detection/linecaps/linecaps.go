// Package linecaps independently checks the repository's Go size policy.
//
// The checker measures physical source lines and AST source ranges. A file or
// function exactly at its configured limit passes; only values above a limit
// produce findings. Function literals are checked as well as named functions.
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
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	if len(paths) == 0 {
		paths, err = Discover(absoluteRoot)
		if err != nil {
			return Report{}, err
		}
	}
	paths = uniquePaths(paths)
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
	name := ""
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
	case *ast.FuncLit:
		name = "function literal"
	default:
		return Finding{}, false
	}
	start := fset.Position(node.Pos()).Line
	end := fset.Position(node.End()).Line
	actual := end - start + 1
	if actual <= limit {
		return Finding{}, false
	}
	return Finding{Path: path, Rule: RuleFunctionLines, Name: name, StartLine: start, EndLine: end, Actual: actual, Limit: limit}, true
}

func resolvePath(root, path string) (string, string, error) {
	if path == "" {
		return "", "", fmt.Errorf("linecaps path must not be empty")
	}
	normalized := strings.ReplaceAll(path, "\\", "/")
	if filepath.IsAbs(filepath.FromSlash(normalized)) {
		fullPath := filepath.Clean(filepath.FromSlash(normalized))
		relative, err := filepath.Rel(root, fullPath)
		if err != nil {
			return "", "", err
		}
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return filepath.ToSlash(relative), fullPath, nil
		}
		return filepath.ToSlash(fullPath), fullPath, nil
	}
	relative := filepath.Clean(filepath.FromSlash(normalized))
	if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(relative), filepath.Join(root, relative), nil
}

func uniquePaths(paths []string) []string {
	result := append([]string(nil), paths...)
	sort.Slice(result, func(i, j int) bool {
		return strings.ReplaceAll(result[i], "\\", "/") < strings.ReplaceAll(result[j], "\\", "/")
	})
	if len(result) < 2 {
		return result
	}
	write := 1
	for _, path := range result[1:] {
		if path == result[write-1] {
			continue
		}
		result[write] = path
		write++
	}
	return result[:write]
}
