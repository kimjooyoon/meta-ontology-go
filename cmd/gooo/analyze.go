package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	fixPlanSchemaVersion   = "gooo-fix-plan/v1"
	fixPlanReady           = "ready"
	fixPlanSyntaxInvalid   = "syntax-invalid"
	fixPlanSemanticInvalid = "semantic-invalid"
)

func runAnalyze(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	return runAnalyzeWithLowerer(args, reader, parser, stdout, stderr, bidir.Lower)
}

func runAnalyzeWithLowerer(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer, lower func(*syntax.File) (semantic.IR, error)) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo analyze <file.gooo>")
		return exitUsage
	}
	filename := args[0]
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, syntaxDiagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", filename, err)
		return exitFailure
	}
	plan := newFixPlan(source, syntaxDiagnostics, file)
	if syntaxDiagnostics.HasErrors() {
		plan.Status = fixPlanSyntaxInvalid
	} else {
		ir, lowerErr := lowerInspectIRWith(file, remainingDeadline(deadline), lower)
		if lowerErr != nil {
			plan.Status = fixPlanSemanticInvalid
			plan.Diagnostics = append(plan.Diagnostics, semanticFixDiagnostics(lowerErr, fileSpan(file))...)
		} else {
			applyFixPlanIR(&plan, ir)
		}
	}
	plan.Diagnostics = canonicalFixPlanDiagnostics(plan.Diagnostics)
	payload, err := marshalFixPlan(plan)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: fix plan failed: %v\n", filename, err)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, payload, deadline); err != nil {
		fmt.Fprintf(stderr, "gooo: fix plan output: %v\n", err)
		return exitFailure
	}
	if plan.Status != fixPlanReady || hasErrorFixDiagnostic(plan.Diagnostics) {
		return exitFailure
	}
	return exitOK
}

func newFixPlan(source []byte, diagnostics syntax.Diagnostics, file *syntax.File) fixPlan {
	plan := fixPlan{
		SchemaVersion: fixPlanSchemaVersion,
		Status:        fixPlanReady,
		SourceDigest:  semantic.StableHash(source),
		IR: graphIRStatus{
			Status: "unavailable", Reason: "semantic IR is unavailable until diagnostics are resolved",
		},
		Evidence: graphReferenceState{
			Status: "missing", Reason: "semantic IR is unavailable; no evidence records can be reported",
		},
		Provenance: graphReferenceState{
			Status: "missing", Reason: "no provenance records are attached",
		},
		Projection: graphStatus{
			Status: "deferred", Reason: "read-only fix plan does not run projection",
		},
		Lowering: graphStatus{
			Status: "deferred", Reason: "bidir lowering has no cooperative cancellation contract",
		},
		Output: graphStatus{
			Status: "deferred", Reason: "generic writers have no cooperative cancellation contract",
		},
		Repairs: graphStatus{
			Status: "deferred", Reason: "automatic repair edits are not generated",
		},
		GraphPatch: graphStatus{
			Status: "deferred", Reason: "read-only fix plan does not produce graph patches",
		},
		Workspace: graphStatus{
			Status: "deferred", Reason: "read-only fix plan does not write workspace files",
		},
		SemanticLoop: graphStatus{
			Status: "deferred", Reason: "full semantic repair loop is not implemented",
		},
		Authorities: graphAuthorities{
			GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
			Provenance: "authoritative", Graph: "derived",
		},
		Diagnostics: syntaxFixDiagnostics(diagnostics),
	}
	if file == nil {
		plan.Status = fixPlanSyntaxInvalid
	}
	return plan
}

func applyFixPlanIR(plan *fixPlan, ir semantic.IR) {
	plan.GraphHash = authoritativeGraphHash(ir.Graph)
	plan.IR = graphIRStatus{Status: "available", SemanticDigest: authoritativeIRHash(ir)}
	refs := make([]string, 0, len(ir.Evidence()))
	for _, evidence := range ir.Evidence() {
		refs = append(refs, evidence.ID.String())
	}
	sort.Strings(refs)
	plan.Evidence = graphReferences(refs, "no semantic evidence records are attached")
}

func syntaxFixDiagnostics(diagnostics syntax.Diagnostics) []fixPlanDiagnostic {
	result := make([]fixPlanDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		result = append(result, newFixPlanDiagnostic(
			"syntax", diagnostic.Severity.String(), string(diagnostic.Code), diagnostic.Message,
			fixPlanSpanFromSyntax(diagnostic.Span), "potential",
		))
	}
	return result
}

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
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Message != right.Message {
			return left.Message < right.Message
		}
		return left.RepairID < right.RepairID
	})
	return result
}

func hasErrorFixDiagnostic(diagnostics []fixPlanDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return true
		}
	}
	return false
}

func marshalFixPlan(plan fixPlan) ([]byte, error) {
	payload, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}
	if len(payload) > maxDiagnosticBytes {
		return nil, errDiagnosticLimit
	}
	return append(payload, '\n'), nil
}
