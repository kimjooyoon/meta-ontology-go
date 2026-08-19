package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
	"strings"
)

func canonicalSemanticDiagnostics(err error) []semanticDiagnostic {
	var validation *semantic.ValidationErrors
	if errors.As(err, &validation) && len(validation.Issues) > 0 {
		issues := append([]semantic.ValidationIssue(nil), validation.Issues...)
		sort.Slice(issues, func(i, j int) bool {
			left, right := issues[i], issues[j]
			if left.Code != right.Code {
				return left.Code < right.Code
			}
			if left.Message != right.Message {
				return left.Message < right.Message
			}
			if left.Subject != right.Subject {
				return left.Subject < right.Subject
			}
			return left.Object < right.Object
		})
		result := make([]semanticDiagnostic, 0, len(issues))
		for _, issue := range issues {
			result = append(result, semanticDiagnostic{
				Code: semanticValidationCode(issue.Code), Message: issue.Message,
			})
		}
		return result
	}
	return []semanticDiagnostic{{Code: semanticDiagnosticCode(err), Message: err.Error()}}
}
func formatSemanticDiagnostics(span syntax.Span, err error) ([]byte, error) {
	diagnostics := canonicalSemanticDiagnostics(err)
	if len(diagnostics) > maxDiagnosticCount {
		return nil, errDiagnosticLimit
	}
	var output strings.Builder
	for _, diagnostic := range diagnostics {
		line := fmt.Sprintf("%s: error %s: %s\n", span.String(), diagnostic.Code, diagnostic.Message)
		if output.Len()+len(line) > maxDiagnosticBytes {
			return nil, errDiagnosticLimit
		}
		output.WriteString(line)
	}
	return []byte(output.String()), nil
}
func semanticValidationCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return "semantic.invalid"
	}
	if strings.HasPrefix(code, "semantic.") {
		return code
	}
	switch code {
	case "relation-kind":
		return "semantic.invalid-kind"
	case "missing-subject", "missing-object":
		return "semantic.invalid-endpoint"
	case "unknown-relation":
		return "semantic.invalid-relation"
	}
	return "semantic." + code
}
