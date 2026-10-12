package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestCriteriaQuestionsRetainInputsAcrossCandidatesAndChangedPredicate(t *testing.T) {
	ctx := context.Background()
	source := []byte(criteriaQuestionFixture())
	g, _ := NewTypedPathGenerator("")
	prior, err := g.GenerateSourceAssembly(ctx, "questions.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := RealizeRecordCandidate(ctx, "questions.gooo", source, "Select", 0); err == nil {
		t.Fatal("question history silently dropped")
	}
	for mask := range 2 {
		candidate, _, err := RealizeObservedRecordCandidate(ctx, "questions.gooo", source, "Select", prior.Report.RecordAssembly, uint16(mask))
		if err != nil || candidate.CriteriaQuestions == nil || candidate.Schema != "gooo/record-candidate/v3" ||
			candidate.Attempt.Criteria.Total != 2 || candidate.Attempt.Criteria.Passed != mask+1 || len(candidate.CriteriaQuestions.AcquiredInputs) != 1 {
			t.Fatal("caller candidate lost acquired criteria input", candidate, err)
		}
	}
	changed := []byte(strings.Replace(string(source), "return input1.value == 0-input0", "return input1.value == input0", 1))
	result, err := ResumeRecordAssembly(ctx, "questions.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 0 || r.Criteria.Passed != 2 || r.Continuation.RecheckedAttempts != 1 || r.Continuation.AddedAttempts != 1 ||
		!equalLocalCallEvidence(r.CriteriaQuestions, prior.Report.RecordAssembly.CriteriaQuestions) {
		t.Fatal("continuation rewrote questions or lost input obligations", r)
	}
	if _, err := RealizeSourceAssembly(ctx, "questions.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
}

func TestCriteriaQuestionsRetainedOracleHonorsCombinedBudget(t *testing.T) {
	// Existing value-oracle retention must count criteria inputs in the shared
	// budget too, even though it retains full expected records instead of questions.
	s := &assemblyspec.Spec{Choices: []assemblyspec.Choice{{ID: "value", Kind: "field_value", Alternative: "0-input", Intent: "Negate."}}, MaxAttempts: 2,
		ValueCases:  []assemblyspec.ValueCase{{Inputs: "[0]", Expected: "{}"}},
		Criteria:    &assemblyspec.Criteria{Predicates: []assemblyspec.Criterion{{ID: "negative", Activity: "Negative"}}},
		Observation: &assemblyspec.Observation{ValueInputs: []string{"[900]"}, OracleActivity: "Expected", MaxCandidates: 2, MaxRounds: 1}}
	for i := range 127 {
		s.Criteria.Inputs = append(s.Criteria.Inputs, fmt.Sprintf("[%d]", i+1))
	}
	contract, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	options, _ := json.Marshal(s.Observation)
	c := assemblyspec.ValueCase{Inputs: "[900]", Expected: "{}"}
	cases := append(append([]assemblyspec.ValueCase(nil), s.ValueCases...), c)
	r := &RecordObservationReceipt{Schema: "gooo/record-observation/v1", Status: "ROUND_BUDGET", SourceSHA256: "origin", ContractSHA256: digest([]byte(contract)),
		Options: *s.Observation, OptionsSHA256: digest(options), InitialCases: s.ValueCases, EffectiveCases: 2, EffectiveCasesSHA256: recordCasesSHA(cases),
		Rounds: []RecordObservationRound{{RecommendedIndex: new(0), Probes: []RecordObservationProbe{{Input: "[900]"}}, Observation: &c}}}
	prior := &RecordAssemblyReceipt{OriginalSourceSHA256: "origin", ContractSHA256: r.ContractSHA256, Observation: r}
	if _, err := retainedRecordObservationCases(&recordAssemblyPlan{spec: s}, prior, r); err == nil {
		t.Fatal("retained oracle case exceeded shared criteria budget")
	}
}
