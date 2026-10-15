package verify

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCommandEvolutionBindsBothChangedWorkflowJobs(t *testing.T) {
	baseline, err := os.ReadFile("../../docs/observations/ci-command-domain/predecessor-command-population.json")
	if err != nil {
		t.Fatal(err)
	}
	ledger := readCommandEvolution(t, baseline)
	var original struct {
		Jobs []struct {
			Name    string   `json:"job"`
			Digests []string `json:"original_command_body_sha256"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(baseline, &original); err != nil {
		t.Fatal(err)
	}
	workflow, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	for job, rows := range ledger.Jobs {
		var digests []string
		for _, prior := range original.Jobs {
			if prior.Name == job {
				digests = prior.Digests
			}
		}
		_, body, found := strings.Cut(string(workflow), "\n  "+job+":\n")
		if !found || len(digests) == 0 {
			t.Fatal("changed job has no original source population", job)
		}
		body, _, _ = strings.Cut(body, "\n  failure-report:\n")
		body, _, _ = strings.Cut(body, "\n  proof:\n")
		if err := bindCommandEvolution(workCommands124073(body), digests, rows); err != nil {
			t.Fatal("changed source lost command ancestry", job, err)
		}
	}
	if len(ledger.Jobs["evidence"]) == 0 || len(ledger.Jobs["proof"]) == 0 {
		t.Fatal("changed tool producer and consumer need explicit command origins")
	}
}
