package formatfix

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/formatter"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Build(filename, source string) Plan {
	plan := Plan{
		Schema: PlanSchema, File: filename, SourceDigest: digestBytes([]byte(source)),
		SourceBytes: len(source), Edits: []Edit{}, Diagnostics: []string{},
	}
	file, diagnostics := syntax.ParseFile(filename, source)
	if diagnostics.HasErrors() {
		for _, diagnostic := range diagnostics {
			plan.Diagnostics = append(plan.Diagnostics, diagnostic.String())
		}
		return lower(plan, "FORMAT_FIX_SOURCE_UNKNOWN")
	}
	formatted := formatter.FormatSyntax(file)
	if formatted.HasErrors() {
		for _, diagnostic := range formatted.Diagnostics {
			plan.Diagnostics = append(plan.Diagnostics, diagnostic.String())
		}
		return lower(plan, "FORMAT_FIX_SURFACE_UNKNOWN")
	}
	before, beforeDiagnostics := formatter.SyntaxAdapter{}.Adapt(file)
	afterFile, afterDiagnostics := syntax.ParseFile(filename, formatted.Source)
	after, adapterDiagnostics := formatter.SyntaxAdapter{}.Adapt(afterFile)
	if before == nil || after == nil || beforeDiagnostics.HasErrors() ||
		afterDiagnostics.HasErrors() || adapterDiagnostics.HasErrors() {
		return lower(plan, "FORMAT_FIX_SEMANTIC_FINGERPRINT_UNKNOWN")
	}
	plan.SemanticBefore = before.SemanticFingerprint()
	plan.SemanticAfter = after.SemanticFingerprint()
	plan.SemanticEqual = plan.SemanticBefore != "" && plan.SemanticBefore == plan.SemanticAfter
	if !plan.SemanticEqual {
		return lower(plan, "FORMAT_FIX_SEMANTIC_DRIFT")
	}
	plan.ResultDigest = digestBytes([]byte(formatted.Source))
	plan.ResultBytes = len(formatted.Source)
	if formatted.Source == source {
		plan.Decision, plan.Resolution = DecisionFixedPoint, ResolutionExact
		plan.ReasonCode = "FORMAT_FIX_ALREADY_CANONICAL"
		return seal(plan)
	}
	plan.Decision, plan.Resolution = DecisionChangePlanned, ResolutionExact
	plan.ReasonCode, plan.Changed = "FORMAT_FIX_CHANGE_PLANNED", true
	plan.Edits = []Edit{{Start: 0, End: len(source), Replacement: formatted.Source}}
	return seal(plan)
}

func lower(plan Plan, reason string) Plan {
	plan.Decision, plan.Resolution, plan.ReasonCode = DecisionFailClosed, ResolutionLower, reason
	plan.Changed, plan.Edits = false, []Edit{}
	return seal(plan)
}
