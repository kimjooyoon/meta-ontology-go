package nonmonotonicrefutationoracle

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutation"
)

func sourceFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../examples/nonmonotonic-refutation/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestIndependentOracleReplaysNonMonotonicHistory(t *testing.T) {
	source := sourceFixture(t)
	value, err := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source, true)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Judge(data, source)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PASS" || report.Metrics.ObservationAttemptTotal != 8 || report.Metrics.AcceptedStateTransitionTotal != 6 || report.Metrics.RejectedObservationTotal != 0 || report.Metrics.TransitionTotal != 8 || report.Metrics.SupportsTotal != 5 || report.Metrics.ContradictsTotal != 3 ||
		report.Metrics.DischargedToRefutedTotal != 2 || report.Metrics.RefutedToDischargedTotal != 1 {
		t.Fatalf("report = %#v", report)
	}
	if report.Cases[0].InitialStatus != "OPEN" || report.Cases[1].StatusHistory[1] != "DISCHARGED" || report.Cases[1].CurrentStatus != "REFUTED" {
		t.Fatalf("refutation history = %#v", report.Cases[1])
	}
	if report.Transitions[2].Relation != "CONTRADICTS" || !strings.Contains(report.Transitions[2].EvidenceBasis, "observed=0") || report.Transitions[2].EvidenceDigest == "" || report.Transitions[2].ProofAdmission == "" || !report.Transitions[2].ProofAdmitted || report.SourceBindingDigest == "" {
		t.Fatalf("refutation basis = %#v", report.Transitions[2])
	}
	if report.Transitions[1].PreviousDigest != report.Transitions[0].TransitionDigest ||
		report.Transitions[2].PreviousDigest != report.Transitions[1].TransitionDigest ||
		report.Transitions[7].PreviousDigest != report.Transitions[6].TransitionDigest {
		t.Fatalf("transition chain = %#v", report.Transitions[:3])
	}
	if report.Cases[2].StatusHistory[2] != "REFUTED" || report.Cases[2].CurrentStatus != "DISCHARGED" || report.Cases[2].ActiveRefutationHistory[len(report.Cases[2].ActiveRefutationHistory)-1] != 0 || report.Transitions[7].RevisionRelation != "SUPERSEDES" || report.Transitions[7].SupersedesEvidenceDigest != report.Transitions[5].EvidenceDigest || report.Transitions[7].CorrectionTargetTransitionDigest != report.Transitions[5].TransitionDigest || report.Transitions[7].Coordinate.Reason != "TARGETED_CORRECTION_SUPERSEDES_EXACT_REFUTATION" {
		t.Fatalf("re-evaluation history = %#v", report.Cases[2])
	}
	if report.Conformance.Decision != "PASS" || report.SubjectResolution.Resolution != "PARTIAL" || report.Vocabulary.FixtureKnowledge != "HISTORICAL_FIXTURE_ONLY" || report.Cases[2].CurrentEvidenceID != "gooo://nonmonotonic-refutation/observation/gamma-5" {
		t.Fatalf("resolution split = %#v / %#v", report.Conformance, report.SubjectResolution)
	}
}

func TestIndependentOracleRejectsSourceDigestMismatch(t *testing.T) {
	source := sourceFixture(t)
	value, err := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source, true)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	changed := append([]byte(nil), source...)
	changed = append(changed, []byte("\n# changed source\n")...)
	report, err := Judge(data, changed)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "SOURCE_RAW_DIGEST_MISMATCH" {
		t.Fatalf("report = %#v", report)
	}
}

func judgeVariant(t *testing.T, replacements ...string) Report {
	t.Helper()
	source := string(sourceFixture(t))
	for index := 0; index < len(replacements); index += 2 {
		changed := strings.Replace(source, replacements[index], replacements[index+1], 1)
		if changed == source {
			t.Fatalf("replacement %q did not match source", replacements[index])
		}
		source = changed
	}
	value, err := producer.Produce("examples/nonmonotonic-refutation/main.gooo", []byte(source), true)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Judge(data, []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestRejectedEvidencePreservesPriorAcceptedState(t *testing.T) {
	unknown := judgeVariant(t,
		"observation_quality=SUFFICIENT;provider_class=HISTORICAL_FIXTURE;provenance=fixture-provider-beta-2", "observation_quality=UNRESOLVED;provider_class=HISTORICAL_FIXTURE;provenance=fixture-provider-beta-2")
	if unknown.Decision != "PASS" || unknown.Resolution != "LOWER_RESOLUTION" || unknown.Transitions[2].Before != "DISCHARGED" || unknown.Transitions[2].After != "DISCHARGED" || unknown.Transitions[2].Accepted || unknown.Cases[1].CurrentStatus != "DISCHARGED" {
		t.Fatalf("unknown after discharge = %#v", unknown)
	}

	insufficient := judgeVariant(t,
		";observed=1;observed_material=fixture:gamma:input-gamma:value-1;observation_quality=SUFFICIENT;provider_class=HISTORICAL_FIXTURE;provenance=fixture-provider-gamma-5",
		";observed=;observed_material=fixture:gamma:input-gamma:value-1;observation_quality=SUFFICIENT;provider_class=HISTORICAL_FIXTURE;provenance=fixture-provider-gamma-5")
	if insufficient.Decision != "PASS" || insufficient.Transitions[7].Relation != "INSUFFICIENT" || insufficient.Transitions[7].Before != "REFUTED" || insufficient.Transitions[7].After != "REFUTED" || insufficient.Transitions[7].Accepted || insufficient.Cases[2].ActiveRefutationTotal != 1 {
		t.Fatalf("insufficient after refutation = %#v", insufficient)
	}
}

func TestOrdinarySupportCannotEraseRefutation(t *testing.T) {
	report := judgeVariant(t,
		"revision_relation=SUPERSEDES;supersedes_claim_id=gooo://nonmonotonic-refutation/claim/gamma;supersedes_evidence_digest=sha256:99b22d86135d07def6e445ce2c08c239522ef0939e07e05e73a6c901327edffa", "revision_relation=NONE;supersedes_claim_id=none;supersedes_evidence_digest=none")
	if report.Decision != "PASS" || report.Transitions[7].Relation != "SUPPORTS" || report.Transitions[7].Before != "REFUTED" || report.Transitions[7].After != "REFUTED" || report.Transitions[7].Accepted || report.Cases[2].CurrentStatus != "REFUTED" || report.Cases[2].ActiveRefutationTotal != 1 {
		t.Fatalf("ordinary support after refutation = %#v", report)
	}
}

func TestProofChoiceAndCorrectionTargetAreCausal(t *testing.T) {
	proofRejected := judgeVariant(t, "proof_choice=REGRESSION;stage=REASSESS;step=05", "proof_choice=FOUNDATION;stage=REASSESS;step=05")
	if proofRejected.Decision != "PASS" || proofRejected.Transitions[7].ProofAdmitted || proofRejected.Transitions[7].After != "REFUTED" || proofRejected.Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("proof rejection = %#v", proofRejected)
	}

	wrongTarget := judgeVariant(t, "supersedes_evidence_digest=sha256:99b22d86135d07def6e445ce2c08c239522ef0939e07e05e73a6c901327edffa", "supersedes_evidence_digest=sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if wrongTarget.Decision != "PASS" || wrongTarget.Transitions[7].After != "REFUTED" || wrongTarget.Transitions[7].Accepted || wrongTarget.Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("wrong correction target = %#v", wrongTarget)
	}
}
