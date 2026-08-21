package linecaps

import (
	"errors"
	"fmt"
	"strings"
)

// Rule identifies the kind of finding in a Report.
type Rule string

const (
	RuleFileLines      Rule = "file-lines"
	RuleFunctionLines  Rule = "function-lines"
	RuleRefactorReturn Rule = "refactor-return"
	RuleRefactorAssign Rule = "refactor-assign-return"
	RuleReadFile       Rule = "read-file"
	RuleParseFile      Rule = "parse-file"
)

// Finding is a deterministic policy violation or an error that prevented a
// file from being independently verified. The latter uses RuleReadFile or
// RuleParseFile and records its explanation in Detail.
type Finding struct {
	Path      string `json:"path"`
	Rule      Rule   `json:"rule"`
	Name      string `json:"name,omitempty"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Actual    int    `json:"actual"`
	Limit     int    `json:"limit"`
	Detail    string `json:"detail,omitempty"`
}

// Report is the complete result of one analysis. Findings are sorted by path,
// source position, rule, and detail before they are returned or formatted.
type Report struct {
	Findings []Finding `json:"findings"`
}

// OK reports whether every requested Go file passed analysis.
func (r Report) OK() bool {
	return len(r.Findings) == 0
}

// Err converts a non-empty report into an error for command-line callers.
func (r Report) Err() error {
	if r.OK() {
		return nil
	}
	return errors.New(strings.TrimSuffix(r.Text(), "\n"))
}

// Text returns stable, line-oriented human output. JSON is the machine-readable
// format; see JSON for the corresponding structured representation.
func (r Report) Text() string {
	findings := orderedFindings(r.Findings)
	if len(findings) == 0 {
		return "linecaps: ok\n"
	}
	var output strings.Builder
	fmt.Fprintf(&output, "linecaps: violations=%d\n", len(findings))
	for _, finding := range findings {
		formatFinding(&output, finding)
	}
	return output.String()
}
