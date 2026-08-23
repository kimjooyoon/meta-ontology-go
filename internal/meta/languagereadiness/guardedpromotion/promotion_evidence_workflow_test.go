package guardedpromotion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromotionEvidenceWorkflowSeparatesProducerFromJudge(t *testing.T) {
	workflowPath := filepath.Join("..", "..", "..", "..", TransformationPath)
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	for _, fragment := range []string{
		"name: Metric counterfactual conformance",
		"actions: read",
	} {
		if !strings.Contains(workflow, fragment) {
			t.Fatalf("promotion evidence authority is missing %q", fragment)
		}
	}
	start := strings.Index(workflow, "\n  proposal-promotion-evidence:\n")
	if start < 0 {
		t.Fatal("promotion evidence producer job is missing")
	}
	producerTail := workflow[start+1:]
	end := strings.Index(producerTail, "\n  program:\n")
	if end < 0 {
		t.Fatal("promotion evidence producer boundary is missing")
	}
	producer := producerTail[:end]
	required := []string{
		"needs: strategy",
		"GOTOOLCHAIN: go1.27.0",
		"persist-credentials: false",
		"go run ./cmd/language-readiness-witness/proposal-promotion",
		`--check "$RUNNER_TEMP/language-readiness-proposal-promotion-a.json"`,
		"name: language-readiness-proposal-promotion-${{ env.HEAD_SHA }}",
		"name: language-readiness-proposal-promotion-v2-${{ env.HEAD_SHA }}",
	}
	for _, fragment := range required {
		if !strings.Contains(producer, fragment) {
			t.Fatalf("promotion evidence workflow is missing %q", fragment)
		}
	}
	for _, forbidden := range []string{
		"needs: proposal-continuity",
		"promotion-authorized-continuity",
		"guarded-promotion-receipt",
	} {
		if strings.Contains(producer, forbidden) {
			t.Fatalf("promotion evidence producer depends on judge %q", forbidden)
		}
	}
}
