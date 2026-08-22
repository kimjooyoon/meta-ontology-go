package proposalpromotion

import "testing"

const (
	testCurrent  = "2222222222222222222222222222222222222222"
	testEvidence = "1111111111111111111111111111111111111111"
)

func validSource() Source {
	return Source{
		Selection: SelectionSource{
			Repository:        "kimjooyoon/meta-ontology-go",
			CurrentSubjectSHA: testCurrent, PredecessorSHA: testEvidence,
			Decision: "SELECTED", Reason: "PROPOSAL_PREDECESSOR_SELECTED",
			ReportDigest: "sha256:selection", RunID: 1, RunAttempt: 1,
			HeadSHA: testEvidence, Event: "push", Status: "completed",
			Conclusion: "success", WorkflowName: "Metric counterfactual conformance",
			SynthesisJobID: 3, SynthesisJobName: "strategy",
			SynthesisJobStatus: "completed", SynthesisJobConclusion: "success",
			ArtifactID: 2, ArtifactName: "metric-strategy-" + testEvidence,
			ProposalFileSHA256: "sha256:file", ProposalReportDigest: "sha256:contract",
			ObservedRuns: 1, ExactRuns: 1, ObservedJobs: 5, ExactJobs: 1,
			ObservedArtifacts: 5, ExactArtifacts: 1,
			ValidCandidates: 1, SelectionBPS: 10_000,
			ProofsPassed: 5, ProofsTotal: 5,
		},
		Contract: ContractSource{
			SubjectSHA: testEvidence, Decision: "PASS",
			Reason:     "CHANGE_PROPOSAL_CONTRACT_READY",
			FileSHA256: "sha256:file", ReportDigest: "sha256:contract",
			SelectedActions: 2, Satisfied: 8, Total: 8, ReadinessBPS: 10_000,
		},
	}
}

func TestFailedWorkflowAcceptsSuccessfulSynthesisJob(t *testing.T) {
	source := validSource()
	source.Selection.Conclusion = "failure"
	receipt := evaluate(testCurrent, testEvidence, source)
	if err := Validate(receipt, testCurrent); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateProducesExactEightCoordinatePromotion(t *testing.T) {
	receipt := evaluate(testCurrent, testEvidence, validSource())
	if err := Validate(receipt, testCurrent); err != nil {
		t.Fatal(err)
	}
	if receipt.Summary.Satisfied != 8 || receipt.Summary.Total != 8 ||
		receipt.Summary.ReadinessBPS != 10_000 || len(receipt.Indicators) != 7 ||
		len(receipt.Proofs) != 3 {
		t.Fatalf("receipt = %+v", receipt)
	}
}

func TestGuardrailsFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Source)
	}{
		{"ambiguity", func(source *Source) { source.Selection.AmbiguousCandidates = 1 }},
		{"mutation authority", func(source *Source) {
			source.Selection.SelectedPromotionAuthorized = true
		}},
	}
	for _, test := range tests {
		source := validSource()
		test.mutate(&source)
		receipt := evaluate(testCurrent, testEvidence, source)
		if receipt.Decision != DecisionFailClosed || Validate(receipt, testCurrent) == nil {
			t.Fatalf("%s receipt = %+v", test.name, receipt)
		}
	}
}

func TestDigestTamperFailsClosed(t *testing.T) {
	receipt := evaluate(testCurrent, testEvidence, validSource())
	receipt.Source.Selection.RunAttempt++
	if Validate(receipt, testCurrent) == nil {
		t.Fatal("tampered receipt accepted")
	}
}
