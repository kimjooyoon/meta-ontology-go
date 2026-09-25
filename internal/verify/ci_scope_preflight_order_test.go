package verify

import (
	"os"
	"strings"
	"testing"
)

func TestScopePreflightRunsBeforeExpensiveEvidenceAndKeepsFinalCheck(t *testing.T) {
	raw, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	names := []string{
		"      - name: Preflight changed-file scope before meta execution",
		"      - name: Emit source metrics (go/gooo + folders/files)",
		"      - name: Execute declared meta-operation plan",
		"      - name: Replay declared meta-operation plan",
		"      - name: Emit source metric receipts and feedback",
		"      - name: Check changed-file scope and PR target",
	}
	positions := make([]int, len(names))
	previous := -1
	for index, name := range names {
		positions[index] = strings.Index(workflow, name)
		if strings.Count(workflow, name) != 1 || positions[index] <= previous {
			t.Fatalf("scope/evidence phase missing, duplicated or reordered: %s", name)
		}
		previous = positions[index]
	}
	preflight := workflow[positions[0]:positions[1]]
	for _, required := range []string{
		"if: ${{ github.event_name == 'pull_request' && github.event.pull_request.base.ref == 'dev' }}",
		"timeout-minutes: 5",
		"GOOO_SCOPE_FROM: ${{ github.event.pull_request.base.sha }}",
		"GOOO_SCOPE_TO: ${{ github.event.pull_request.head.sha }}",
		"GOOO_EXPECTED_HEAD: ${{ github.event.pull_request.head.sha }}",
		"GOOO_SCOPE_BRANCH: ${{ github.event.pull_request.head.ref }}",
		"run: go run ./scripts/verify --skip-caps --head= --base=",
	} {
		if !strings.Contains(preflight, required) {
			t.Errorf("preflight lost exact input or failure behavior: %s", required)
		}
	}
	for _, forbidden := range []string{"continue-on-error", "|| true", "METRICS_DIR", "GOOO_HUMAN_DECISION"} {
		if strings.Contains(preflight, forbidden) {
			t.Errorf("preflight depends on late evidence or grants/ignores authority: %s", forbidden)
		}
	}
	final := workflow[positions[len(positions)-1]:]
	if end := strings.Index(final[len(names[len(names)-1]):], "\n      - "); end >= 0 {
		final = final[:len(names[len(names)-1])+end]
	}
	if !strings.Contains(final, "go run ./scripts/verify --skip-caps") ||
		strings.Contains(final, "--head=") || strings.Contains(final, "--base=") ||
		strings.Contains(final, "continue-on-error") || strings.Contains(final, "|| true") {
		t.Fatal("the original final scope and route gate was weakened")
	}
}
