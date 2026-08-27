package proposalpromotion

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

const (
	testCurrent  = "2222222222222222222222222222222222222222"
	testEvidence = "1111111111111111111111111111111111111111"
)

func testObservationEvidence() proposalpredecessor.ObservationEvidence {
	return proposalpredecessor.ObservationEvidence{
		Schema: proposalpredecessor.ObservationSchema, CachePath: "/tmp/proposal-observation.json",
		CacheBytes: 1, CacheDigest: "sha256:" + strings.Repeat("0", 64),
		ResponseTotal: 1, ResponseConsumed: 1,
	}
}

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
	receipt := evaluate(testCurrent, testEvidence, source, testObservationEvidence())
	if err := Validate(receipt, testCurrent); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateProducesExactEightCoordinatePromotion(t *testing.T) {
	receipt := evaluate(testCurrent, testEvidence, validSource(), testObservationEvidence())
	if err := Validate(receipt, testCurrent); err != nil {
		t.Fatal(err)
	}
	if receipt.Summary.Satisfied != 8 || receipt.Summary.Total != 8 ||
		receipt.Summary.ReadinessBPS != 10_000 || len(receipt.Indicators) != 7 ||
		len(receipt.Proofs) != 3 {
		t.Fatalf("receipt = %+v", receipt)
	}
}
