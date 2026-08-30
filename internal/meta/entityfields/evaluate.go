package entityfields

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/entityfieldsv1"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Evaluate(observation entityfieldsv1.Observation) Report {
	report := baseReport(observation)
	report.Cells = make([]CellObservation, 0, len(cellSpecs))
	for _, spec := range cellSpecs {
		cell := CellObservation{CellSpec: spec, Observed: "CLOSED", Expected: "CLOSED", Decision: DecisionPass, Resolution: ResolutionExact, EvidenceDigest: observation.EvidenceDigests[spec.EvidenceKey]}
		if cell.EvidenceDigest == "" {
			cell = unknownCell(spec, "ENTITY_FIELDS_EVIDENCE_MISSING")
		}
		report.Cells = append(report.Cells, cell)
	}
	report.Summary = summarize(report.Cells)
	if report.Summary.UnknownCells > 0 {
		report.Decision = DecisionRefuted
		report.Resolution = ResolutionLower
		report.Reason = "ENTITY_FIELDS_EVIDENCE_INCOMPLETE"
	} else {
		report.Decision = DecisionPass
		report.Resolution = ResolutionExact
		report.Reason = "ENTITY_FIELDS_V1_CLOSED"
	}
	report.Counterexamples = FixedCounterexamples()
	report.HumanReport = humanReport(report)
	return report
}

func baseReport(observation entityfieldsv1.Observation) Report {
	profile := observation.Profile
	activities := make([]string, 0, len(cellSpecs))
	for _, spec := range cellSpecs {
		activities = append(activities, spec.Activity)
	}
	return Report{Schema: ReportSchema, ProfileID: profile.ID, ProfileVersion: profile.Version, ProfileDigest: profile.Digest, Activities: activities, ActivityCount: len(activities), BindingCount: len(activities), EvidenceDigests: cloneDigests(observation.EvidenceDigests), Authority: Authority{PromotionAuthorized: false}, Improvement: "UNKNOWN"}
}

func unknownCell(spec CellSpec, reason string) CellObservation {
	unknown := &Unknown{Stage: "observe-entity-fields", Step: spec.Activity, Reason: reason, UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_ENTITY_FIELDS_EVIDENCE", BlockedBy: []string{}}
	return CellObservation{CellSpec: spec, Observed: "UNKNOWN", Expected: "CLOSED", Decision: "UNKNOWN", Resolution: ResolutionLower, Reason: reason, Unknown: unknown}
}

func summarize(cells []CellObservation) Summary {
	summary := Summary{CellsTotal: len(cells)}
	for _, cell := range cells {
		if cell.Decision == DecisionPass { summary.ClosedCells++ }
		if cell.Decision == "UNKNOWN" { summary.UnknownCells++ }
		if cell.Decision == DecisionRefuted { summary.RefutedCells++ }
		switch cell.ProofChoice { case "FOUNDATION": summary.FoundationCells++; case "COHERENCE": summary.CoherenceCells++; case "REGRESSION": summary.RegressionCells++ }
		switch cell.IndicatorClass { case "DRIVER": summary.DriverCells++; case "OUTCOME": summary.OutcomeCells++; case "GUARDRAIL": summary.GuardrailCells++ }
	}
	return summary
}

func cloneDigests(values map[string]string) map[string]string {
	copy := make(map[string]string, len(values))
	for key, value := range values { copy[key] = value }
	return copy
}

func humanReport(report Report) string {
	return fmt.Sprintf("EntityFields V1: cells=%d closed=%d unknown=%d refuted=%d activities=%d bindings=%d proof=%d/%d/%d indicators=%d/%d/%d decision=%s resolution=%s", report.Summary.CellsTotal, report.Summary.ClosedCells, report.Summary.UnknownCells, report.Summary.RefutedCells, report.ActivityCount, report.BindingCount, report.Summary.FoundationCells, report.Summary.CoherenceCells, report.Summary.RegressionCells, report.Summary.DriverCells, report.Summary.OutcomeCells, report.Summary.GuardrailCells, report.Decision, report.Resolution)
}

func FailureReport(reason string) Report {
	return Report{Schema: ReportSchema, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: reason, Counterexamples: []Counterexample{{ID: "entity-fields-observation-failure", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: reason, PartialOutput: false}}, Authority: Authority{PromotionAuthorized: false}, Improvement: "UNKNOWN", HumanReport: "EntityFields V1 rejected input: " + strings.TrimSpace(reason)}
}

func FixedCounterexamples() []Counterexample {
	return []Counterexample{
		{ID: "unsupported-type", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ENTITY_FIELDS_UNSUPPORTED_TYPE", Expected: "string", Observed: "number", PartialOutput: false},
		{ID: "unsupported-presence", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ENTITY_FIELDS_UNSUPPORTED_PRESENCE", Expected: "required", Observed: "optional", PartialOutput: false},
		{ID: "unsupported-cardinality", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ENTITY_FIELDS_UNSUPPORTED_CARDINALITY", Expected: "one", Observed: "many", PartialOutput: false},
		{ID: "stable-id-collision", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ENTITY_FIELDS_ID_COLLISION", Expected: "global unique ID", Observed: "duplicate", PartialOutput: false},
		{ID: "formatted-replay-mismatch", Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ENTITY_FIELDS_REPLAY_MISMATCH", Expected: "same canonical bytes", Observed: "different canonical bytes", PartialOutput: false},
		{ID: "missing-source", Decision: "UNKNOWN", Resolution: ResolutionLower, Reason: "ENTITY_FIELDS_SOURCE_MISSING", PartialOutput: false, Unknown: &Unknown{Stage: "observe-entity-fields", Step: "read-source", Reason: "ENTITY_FIELDS_SOURCE_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_ENTITY_FIELDS_SOURCE", BlockedBy: []string{}}},
	}
}

func cellEvidenceContract(cell CellObservation, observation entityfieldsv1.Observation) error {
	if cell.EvidenceDigest == "" || observation.EvidenceDigests[cell.EvidenceKey] != cell.EvidenceDigest { return fmt.Errorf("cell %s evidence is not bound", cell.ID) }
	return nil
}

func validateProfile(report Report) error {
	profile := syntax.EntityFieldsV1Support().Profile
	if report.ProfileID != profile.ID || report.ProfileVersion != profile.Version || report.ProfileDigest != profile.Digest { return fmt.Errorf("EntityFields profile mismatch") }
	return nil
}
