package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func writeFixture(t *testing.T, root string, duplicate bool) string {
	t.Helper()
	sha := strings.Repeat("a", 40)
	candidate := feedbackpredecessor.Candidate{
		ArtifactID: 11, RunID: 22, RunAttempt: 1,
		ArtifactName: "artifact-feedback-resolution-" + sha,
		HeadSHA: sha, HeadBranch: "dev", Workflow: "CI", Event: "push",
		Conclusion: "success", ReceiptDigest: "sha256:" + strings.Repeat("1", 64),
	}
	candidates := []feedbackpredecessor.Candidate{candidate}
	if duplicate {
		candidates = append(candidates, candidate)
	}
	input := feedbackpredecessor.Input{
		Repository: "kimjooyoon/meta-ontology-go", PredecessorSHA: sha,
		CanonicalBranch: "dev", CanonicalWorkflow: "CI", Candidates: candidates,
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "input.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
