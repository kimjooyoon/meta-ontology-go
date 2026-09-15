package linecaps

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

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
