package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func semanticFixDiagnostics(err error, span fixPlanSpan) []fixPlanDiagnostic {
	var validation *semantic.ValidationErrors
	if !errors.As(err, &validation) || len(validation.Issues) == 0 {
		return []fixPlanDiagnostic{newFixPlanDiagnostic(
			"semantic", "error", "semantic.lowering", err.Error(), span, "not-evaluated",
		)}
	}
	semanticDiagnostics := canonicalSemanticDiagnostics(err)
	result := make([]fixPlanDiagnostic, 0, len(semanticDiagnostics))
	for _, diagnostic := range semanticDiagnostics {
		result = append(result, newFixPlanDiagnostic(
			"semantic", "error", diagnostic.Code, diagnostic.Message, span, "not-evaluated",
		))
	}
	return result
}
func newFixPlanDiagnostic(phase, severity, code, message string, span fixPlanSpan, applicability string) fixPlanDiagnostic {
	return fixPlanDiagnostic{
		RepairID: stableRepairID(phase, severity, code, message, span), Phase: phase,
		Severity: severity, Code: code, Message: message, Span: span,
		Applicability: applicability, Status: "deferred",
	}
}
func stableRepairID(phase, severity, code, message string, span fixPlanSpan) string {
	canonical := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", phase, severity, code, message, span.canonical())
	return "repair-" + semantic.StableHash([]byte(canonical))
}
func canonicalFixPlanDiagnostics(diagnostics []fixPlanDiagnostic) []fixPlanDiagnostic {
	result := make([]fixPlanDiagnostic, 0, len(diagnostics))
	result = append(result, diagnostics...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Span.File != right.Span.File {
			return left.Span.File < right.Span.File
		}
		if left.Span.Start.Offset != right.Span.Start.Offset {
			return left.Span.Start.Offset < right.Span.Start.Offset
		}
		if left.Span.End.Offset != right.Span.End.Offset {
			return left.Span.End.Offset < right.Span.End.Offset
		}
		if left.Phase != right.Phase {
			return left.Phase < right.Phase
		}
		if left.Severity != right.Severity {
			return left.Severity < right.Severity
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Message != right.Message {
			return left.Message < right.Message
		}
		if left.Applicability != right.Applicability {
			return left.Applicability < right.Applicability
		}
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		return left.RepairID < right.RepairID
	})
	return result
}
