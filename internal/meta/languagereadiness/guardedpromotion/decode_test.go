package guardedpromotion

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromotionJSONRequiresOneJSONFile(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("promotion.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte(`{"decision":"PASS"}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := promotionJSON(buffer.Bytes())
	if err != nil || string(data) != `{"decision":"PASS"}` {
		t.Fatalf("data=%s err=%v", data, err)
	}
}

func TestPromotionJSONRejectsEmptyArchive(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := promotionJSON(buffer.Bytes()); err == nil {
		 t.Fatal("empty archive was accepted")
	}
}

func TestPromotionEvidenceWorkflowSeparatesProducerFromJudge(t *testing.T) {
	workflowPath := filepath.Join("..", "..", "..", "..", TransformationPath)
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	_, tail, found := strings.Cut(workflow, "\n  proposal-promotion-evidence:\n")
	producer, _, bounded := strings.Cut(tail, "\n  program:\n")
	if !found || !bounded || !strings.Contains(workflow, "actions: read") {
		t.Fatal("promotion evidence authority boundary is missing")
	}
	for _, fragment := range []string{
		"needs: strategy", "GOTOOLCHAIN: go1.27.0", "persist-credentials: false",
		"go run ./cmd/language-readiness-witness/proposal-promotion",
		`--check "$RUNNER_TEMP/language-readiness-proposal-promotion-a.json"`,
		"name: language-readiness-proposal-promotion-${{ env.HEAD_SHA }}",
		"name: language-readiness-proposal-promotion-v2-${{ env.HEAD_SHA }}",
	} {
		if !strings.Contains(producer, fragment) {
			t.Fatalf("promotion evidence producer is missing %q", fragment)
		}
	}
	if strings.Contains(producer, "guarded-promotion") ||
		strings.Contains(producer, "needs: proposal-continuity") {
		t.Fatal("promotion evidence producer depends on its judge")
	}
}
