package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBodyCodegenCallBudgetExplainsSourceEdge(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	source := "package budget\nnamespace budget\nentity Integer id \"urn:budget:integer\"\n" +
		"activity Leaf(Integer) -> Integer id \"urn:budget:leaf\" computes `return input`\n" +
		"activity Step(Integer) -> Integer id \"urn:budget:step\" computes `return Leaf(input) + Leaf(input) + Leaf(input) + Leaf(input)`\n" +
		"activity Run(Integer) -> Integer id \"urn:budget:run\" computes `let total = 0; for let i = 0; i < 1024; i++ { total = Step(total) }; return total`\n"
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--activity", "Run", "budget.gooo"},
		fixtureReader{source: source}, &stdout, &stderr)
	var payload struct {
		Decision         string         `json:"decision"`
		Source           string         `json:"source"`
		RepositoryWrites int            `json:"repository_writes"`
		Error            string         `json:"error"`
		CallBudget       map[string]any `json:"call_budget"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err, stdout.String())
	}
	if code != exitFailure || stderr.Len() != 0 || payload.Decision != "FAIL_CLOSED" ||
		payload.Source != "" || payload.RepositoryWrites != 0 || !strings.Contains(payload.Error, "4096 calls") {
		t.Fatal("budget failure emitted code or lost its failure status", code, payload, stderr.String())
	}
	for key, want := range map[string]any{
		"caller": "Run", "caller_id": "urn:budget:run", "callee": "Step", "callee_id": "urn:budget:step",
		"call_ordinal": float64(0), "limit": float64(4096), "prefix_before": float64(0),
		"multiplicity": float64(1024), "multiplicity_capped": false, "callee_bound": float64(4),
		"prefix_after": float64(5120),
	} {
		if payload.CallBudget[key] != want {
			t.Errorf("call_budget.%s = %v, want %v", key, payload.CallBudget[key], want)
		}
	}
}
