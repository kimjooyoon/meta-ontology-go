package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
)

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
