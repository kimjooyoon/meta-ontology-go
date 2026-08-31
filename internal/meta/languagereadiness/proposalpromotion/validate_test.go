package proposalpromotion

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

const (
	testRepository = "kimjooyoon/meta-ontology-go"
	testCurrent    = "2222222222222222222222222222222222222222"
	testEvidence   = "1111111111111111111111111111111111111111"
)

func testObservationEvidence() proposalpredecessor.ObservationEvidence {
	raw := []byte(`{"schema":"gooo/language-readiness-api-observation/v1","responses":[{"kind":"GET","url":"https://api.example.test/observed","status_code":200,"body":"eyJvayI6dHJ1ZX0="}]}`)
	sum := sha256.Sum256(raw)
	return proposalpredecessor.ObservationEvidence{
		Schema: proposalpredecessor.ObservationSchema, CachePath: proposalpredecessor.ObservationMemberPath,
		CacheRole:  proposalpredecessor.ObservationRole,
		CacheBytes: len(raw), CacheDigest: "sha256:" + hex.EncodeToString(sum[:]),
		ResponseTotal: 1, ResponseConsumed: 1,
	}
}

func validSource() Source {
	return Source{
		Selection: SelectionSource{
			Repository:        testRepository,
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
	if err := Validate(receipt, testRepository, testCurrent, testEvidence); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateProducesExactEightCoordinatePromotion(t *testing.T) {
	receipt := evaluate(testCurrent, testEvidence, validSource(), testObservationEvidence())
	if err := Validate(receipt, testRepository, testCurrent, testEvidence); err != nil {
		t.Fatal(err)
	}
	if receipt.Summary.Satisfied != 8 || receipt.Summary.Total != 8 ||
		receipt.Summary.ReadinessBPS != 10_000 || len(receipt.Indicators) != 7 ||
		len(receipt.Proofs) != 3 {
		t.Fatalf("receipt = %+v", receipt)
	}
}

func TestResealedPromotionContextMismatchFailsClosed(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"repository", func(receipt *Receipt) { receipt.Repository = "other/repository" }},
		{"current head", func(receipt *Receipt) { receipt.CurrentHeadSHA = strings.Repeat("3", 40) }},
		{"predecessor", func(receipt *Receipt) { receipt.EvidenceHeadSHA = strings.Repeat("4", 40) }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			receipt := evaluate(testCurrent, testEvidence, validSource(), testObservationEvidence())
			test.mutate(&receipt)
			receipt = seal(receipt)
			if err := Validate(receipt, testRepository, testCurrent, testEvidence); err == nil ||
				err.Error() != "FAIL_CLOSED: proposal promotion context mismatch" {
				t.Fatalf("context mutation accepted or misclassified: %v", err)
			}
		})
	}
}
