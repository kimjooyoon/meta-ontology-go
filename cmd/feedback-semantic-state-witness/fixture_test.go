package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func writeSemanticFixture(t *testing.T, root, decision string) (string, string) {
	t.Helper()
	sha, repository := strings.Repeat("c", 40), "example/meta"
	payload, receiptDigest := resolutionPayload(t, sha, repository, decision)
	input := feedbackpredecessor.Input{
		Repository: repository, PredecessorSHA: sha, CanonicalBranch: "dev", CanonicalWorkflow: "CI",
		Candidates: []feedbackpredecessor.Candidate{{
			ArtifactID: 11, RunID: 22, RunAttempt: 1,
			ArtifactName: "artifact-feedback-resolution-" + sha,
			HeadSHA:      sha, HeadBranch: "dev", Workflow: "CI", Event: "push", Conclusion: "success",
			ReceiptDigest: receiptDigest, PayloadDigest: rawDigest(payload),
			ReceiptPayload: base64.StdEncoding.EncodeToString(payload),
		}},
	}
	selection, err := feedbackpredecessor.Select(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt := predecessorReceipt{Schema: predecessorReceiptSchema, Report: selection,
		ReplayReportDigest: selection.ReportDigest, ReplayVerified: true,
		ReceiptDigest: "sha256:" + strings.Repeat("d", 64)}
	inputPath, receiptPath := filepath.Join(root, "input.json"), filepath.Join(root, "predecessor.json")
	for path, value := range map[string]any{inputPath: input, receiptPath: receipt} {
		data, marshalErr := json.MarshalIndent(value, "", "  ")
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		if writeErr := os.WriteFile(path, append(data, '\n'), 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	return inputPath, receiptPath
}
