package linecaps

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JSON returns the stable machine-readable report with a trailing newline.
func (r Report) JSON() ([]byte, error) {
	report := Report{Findings: orderedFindings(r.Findings)}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		if left.EndLine != right.EndLine {
			return left.EndLine < right.EndLine
		}
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Actual != right.Actual {
			return left.Actual < right.Actual
		}
		if left.Limit != right.Limit {
			return left.Limit < right.Limit
		}
		return left.Detail < right.Detail
	})
}
func orderedFindings(findings []Finding) []Finding {
	ordered := append([]Finding(nil), findings...)
	sortFindings(ordered)
	return ordered
}
func formatFinding(output *strings.Builder, finding Finding) {
	switch finding.Rule {
	case RuleFileLines:
		fmt.Fprintf(output, "%s: %s: got %d, limit %d\n", finding.Path, finding.Rule, finding.Actual, finding.Limit)
	case RuleFunctionLines:
		fmt.Fprintf(output, "%s:%d-%d: %s %s: got %d, limit %d\n", finding.Path, finding.StartLine, finding.EndLine, finding.Rule, finding.Name, finding.Actual, finding.Limit)
	case RuleRefactorReturn, RuleRefactorAssign:
		fmt.Fprintf(output, "%s:%d-%d: %s %s: %s (actual=%d)\n", finding.Path, finding.StartLine, finding.EndLine, finding.Rule, finding.Name, finding.Detail, finding.Actual)
	default:
		fmt.Fprintf(output, "%s: %s: %s\n", finding.Path, finding.Rule, finding.Detail)
	}
}
