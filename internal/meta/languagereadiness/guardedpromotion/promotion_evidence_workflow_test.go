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
	required := []string{
		"name: Proposal promotion evidence ledger",
		`workflows: ["CI [push full]"]`,
		"github.event.workflow_run.conclusion == 'success'",
		"GOTOOLCHAIN: go1.27.0",
		"persist-credentials: false",
		"go run ./cmd/language-readiness-witness/proposal-promotion",
		`--check "$RUNNER_TEMP/language-readiness-proposal-promotion-a.json"`,
		"name: language-readiness-proposal-promotion-${{ env.HEAD_SHA }}",
		"name: language-readiness-proposal-promotion-v2-${{ env.HEAD_SHA }}",
	}
	for _, fragment := range required {
		if !strings.Contains(workflow, fragment) {
			t.Fatalf("promotion evidence workflow is missing %q", fragment)
		}
	}
	for _, forbidden := range []string{"promotion-authorized-continuity", "guarded-promotion-receipt"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("promotion evidence producer depends on judge %q", forbidden)
		}
	}
}
