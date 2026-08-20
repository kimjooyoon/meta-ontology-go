package cycles

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Has reports whether at least one diagnostic has code.
func (d Diagnostics) Has(code Code) bool {
	for _, diagnostic := range d {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

// Codes returns the distinct diagnostic codes in lexical order.
func (d Diagnostics) Codes() []Code {
	seen := make(map[Code]struct{}, len(d))
	for _, diagnostic := range d {
		seen[diagnostic.Code] = struct{}{}
	}
	result := make([]Code, 0, len(seen))
	for code := range seen {
		result = append(result, code)
	}
	slices.Sort(result)
	return result
}
func sortDiagnostics(diagnostics Diagnostics) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		left, right := diagnostics[i], diagnostics[j]
		leftKey := diagnosticKey(left)
		rightKey := diagnosticKey(right)
		return leftKey < rightKey
	})
}
func diagnosticKey(diagnostic Diagnostic) string {
	return strings.Join([]string{
		string(diagnostic.Code), diagnostic.Namespace, diagnostic.Name,
		diagnostic.NodeID, diagnostic.Subject, string(diagnostic.Predicate),
		diagnostic.Object, strings.Join(diagnostic.Cycle, "\x00"),
		diagnostic.Message, diagnostic.Span.File,
		fmt.Sprintf("%09d:%09d", diagnostic.Span.Start.Line, diagnostic.Span.Start.Column),
	}, "\x00")
}
