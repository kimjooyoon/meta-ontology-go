package feedbackpredecessor

import (
	"strings"
	"testing"
)

func predecessorFixture() Input {
	sha := strings.Repeat("a", 40)
	return Input{Repository: "kimjooyoon/meta-ontology-go", PredecessorSHA: sha,
		CanonicalBranch: "dev", CanonicalWorkflow: "CI", Candidates: []Candidate{{
			ArtifactID: 11, RunID: 22, RunAttempt: 1,
			ArtifactName: "artifact-feedback-resolution-" + sha,
			HeadSHA: sha, HeadBranch: "dev", Workflow: "CI", Event: "push",
			Conclusion: "success", ReceiptDigest: "sha256:" + strings.Repeat("1", 64),
		}}}
}

func TestSelectBindsUniqueCanonicalPredecessor(t *testing.T) {
	report, err := Select(predecessorFixture())
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionSelected || report.Selected == nil ||
		report.Selected.ArtifactID != 11 || report.ReportDigest == "" {
		t.Fatalf("selection = %#v", report)
	}
	proofs := map[string]bool{}
	for _, metric := range report.Indicators {
		if !metric.Satisfied || metric.Producer == "" || metric.MetaOperation == "" {
			t.Fatalf("indicator = %#v", metric)
		}
		proofs[metric.ProofChoice] = true
	}
	if len(report.Indicators) != 7 || !proofs[ProofFoundation] ||
		!proofs[ProofCoherence] || !proofs[ProofRegression] {
		t.Fatalf("proof choices = %#v", proofs)
	}
}

func TestSelectFailsClosedByCause(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		reason string
	}{
		{"stale", func(in *Input) { in.Candidates[0].HeadSHA = strings.Repeat("b", 40) }, ReasonNotFound},
		{"unsuccessful", func(in *Input) { in.Candidates[0].Conclusion = "failure" }, ReasonUnsuccessful},
		{"unbound receipt", func(in *Input) { in.Candidates[0].ReceiptDigest = "" }, ReasonReceiptUnbound},
		{"write effect", func(in *Input) { in.Candidates[0].RepositoryWrites = 1 }, ReasonWriteEffect},
		{"ambiguous", func(in *Input) {
			in.Candidates = append(in.Candidates, in.Candidates[0])
		}, ReasonAmbiguous},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := predecessorFixture()
			test.mutate(&input)
			report, err := Select(input)
			if err != nil {
				t.Fatal(err)
			}
			if report.Decision != DecisionFailClosed || report.Reason != test.reason ||
				report.Selected != nil {
				t.Fatalf("selection = %#v", report)
			}
		})
	}
}
