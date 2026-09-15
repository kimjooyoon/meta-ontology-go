package proposalpredecessor

import (
	"strings"
	"testing"
)

func TestSelectsOneExactMergedProposal(t *testing.T) {
	predecessor := strings.Repeat("a", 40)
	candidate := readyCandidate(predecessor, 7)
	collection := Collection{RequestedRoute: RouteDev, ObservedRuns: 1, ExactRuns: 1, ObservedJobs: 5,
		ExactJobs: 1, ObservedArtifacts: 4, ExactArtifacts: 1,
		Candidates: []Candidate{candidate}}
	report, payload, err := Select("owner/repository", strings.Repeat("b", 40), predecessor, collection)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Ready() || report.Summary.ValidCandidates != 1 || report.Summary.ProofsPassed != 5 || report.Summary.ProofsTotal != 5 || string(payload) != "proposal" {
		t.Fatalf("unexpected selection: %+v %q", report, payload)
	}
}

func TestAmbiguityFailsClosed(t *testing.T) {
	predecessor := strings.Repeat("c", 40)
	collection := Collection{RequestedRoute: RouteDev, ObservedRuns: 2, ExactRuns: 2, ObservedJobs: 10,
		ExactJobs: 2, ObservedArtifacts: 2, ExactArtifacts: 2,
		Candidates: []Candidate{readyCandidate(predecessor, 1), readyCandidate(predecessor, 2)}}
	report, _, err := Select("owner/repository", strings.Repeat("d", 40), predecessor, collection)
	if err == nil || report.Decision != "FAIL_CLOSED" || report.Reason != "PROPOSAL_PREDECESSOR_AMBIGUOUS" || report.Summary.AmbiguousCandidates != 1 {
		t.Fatalf("ambiguity did not fail closed: %+v %v", report, err)
	}
}

func readyCandidate(predecessor string, runID int64) Candidate {
	return Candidate{
		RunID: runID, RunAttempt: 1, HeadSHA: predecessor, HeadBranch: RouteDev, Event: "push", Status: "completed", Conclusion: "failure", WorkflowName: workflowName,
		SynthesisJobID: runID + 50, SynthesisJobName: synthesisJobName,
		SynthesisJobStatus: "completed", SynthesisJobConclusion: "success",
		ArtifactID: runID + 100, ArtifactName: "metric-strategy-" + predecessor,
		ProposalFileSHA256: "sha256:file", ProposalReportDigest: "sha256:report",
		ContractSatisfied: 8, ContractTotal: 8, ContractBPS: 10000, ProposalPayload: []byte("proposal")}
}
