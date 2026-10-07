package main

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/domaincompleteness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
)

func TestGenerationCoverageDistinguishesObservedZero(t *testing.T) {
	tests := []struct {
		name, state, evidence, want string
		unknown                     int
	}{
		{"observed gap", languageutility.StateOpen, "PASS", "PROGRESS", 0},
		{"unobserved cell", languageutility.StateUnknown, "PASS", "UNKNOWN", 1},
		{"missing stage", "missing", "PASS", "UNKNOWN", 1},
		{"missing reference", languageutility.StateClosed, "missing-reference", "UNKNOWN", 1},
		{"unvalidated reference", languageutility.StateClosed, "UNKNOWN", "UNKNOWN", 1},
		{"refuted cell", languageutility.StateRefuted, "PASS", "FAIL_CLOSED", 0},
		{"fulfilled cell", languageutility.StateClosed, "PASS", "PASS", 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := generationObservationFixture(t, test.state, test.evidence)
			wantNumerator := 0
			if test.want == "PASS" {
				wantNumerator = 1
			}
			if got.Status != test.want || got.Numerator != wantNumerator || got.Denominator != 1 || got.UnknownUnits != test.unknown {
				t.Fatalf("generation observation = %#v, want %s %d/1 unknown=%d", got, test.want, wantNumerator, test.unknown)
			}
			if test.want != "PASS" && got.FirstUnresolved == nil {
				t.Fatal("incomplete observation lost its next step")
			}
		})
	}
}

func generationObservationFixture(t *testing.T, state, evidence string) Dimension {
	t.Helper()
	contract := languageutility.Contract{UseCases: []languageutility.UseCaseSpec{{ID: "fixture"}}}
	cell := languageutility.CellResult{UseCaseID: "fixture", StageID: "USER_ARTIFACT_VERIFIED", State: state,
		Producer: "fixture", Step: "VERIFY_ARTIFACT", Reason: "OBSERVED", EvidencePath: "artifact.json", EvidenceDigest: digestBytes([]byte("artifact"))}
	if state == "missing" {
		cell.StageID = "SYNTAX_ACCEPTED"
	}
	if evidence == "missing-reference" {
		cell.EvidenceDigest = ""
		evidence = "PASS"
	}
	inputs := loadedInputs{evidenceState: evidence, discovery: capabilityDiscoveryEvidence{State: "PASS"}}
	report := languageutility.Report{Cells: []languageutility.CellResult{cell}}
	return measureGenerationCoverage(testProfileDimensions(t)[1], contract, report, inputs)
}

func TestObservedZeroCanBeComparedWithoutCountingUnknownAsProgress(t *testing.T) {
	baseline := observedZeroBaseline(t)
	current := baseline
	current.SubjectSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	current.Snapshot.SubjectSHA = current.SubjectSHA
	current.Dimensions = []Dimension{{ID: "generation_coverage", MetricID: "metric", Unit: "use_cases",
		Status: classify(2, 4, 0, false), Numerator: 2, Denominator: 4}}
	comparison := compareReports(current, baseline, true)
	if comparison.Status != "COMPARABLE" || comparison.Dimensions[0].BaselineStatus != "PROGRESS" || comparison.Dimensions[0].NumeratorDelta != 2 {
		t.Fatalf("measured zero comparison = %#v", comparison)
	}
	baseline.Dimensions[0].Status = classify(0, 4, 4, false)
	baseline.Dimensions[0].UnknownUnits = 4
	baseline.Digest, _ = reportDigest(baseline)
	if got := compareReports(current, baseline, true); got.Status != "PARTIAL" || got.Dimensions[0].Status != "UNKNOWN_UNRESOLVED_EVIDENCE" {
		t.Fatalf("unobserved baseline became a measured improvement: %#v", got)
	}
	baseline.Contract.SemanticHash = "previous-classifier"
	baseline.Generated.SemanticHash = "previous-classifier"
	baseline.Digest, _ = reportDigest(baseline)
	if got := compareReports(current, baseline, true); got.Status != "UNKNOWN_INCOMPATIBLE_BASELINE" {
		t.Fatalf("different classifier semantics became comparable: %#v", got)
	}
}

func observedZeroBaseline(t *testing.T) Report {
	t.Helper()
	report := Report{Schema: ReceiptSchema, ProfileID: ProfileID,
		SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Contract:   ContractRef{SemanticHash: "profile"},
		Generated:  GeneratedRef{SemanticHash: "profile", SemanticsEqual: true},
		Snapshot:   Snapshot{Repository: "owner/repo", SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		Dimensions: []Dimension{{ID: "generation_coverage", MetricID: "metric", Unit: "use_cases",
			Status: classify(0, 4, 0, false), Denominator: 4}},
	}
	var err error
	report.Digest, err = reportDigest(report)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestCompletenessProjectionMatchesGoooProfile(t *testing.T) {
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if err := domaincompleteness.VerifyGeneratedProjection("profile.gooo", source); err != nil {
		t.Fatal(err)
	}
}
