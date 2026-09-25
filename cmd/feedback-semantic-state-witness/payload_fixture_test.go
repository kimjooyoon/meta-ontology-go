package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackstate"
)

func resolutionPayload(t *testing.T, sha, repository, decision string) ([]byte, string) {
	t.Helper()
	receiptDigest := "sha256:" + strings.Repeat("a", 64)
	reportDigest := "sha256:" + strings.Repeat("b", 64)
	nextOperation := ""
	if decision == "IMPROVE" {
		nextOperation = "split-go-declarations"
	}
	value := map[string]any{
		"schema": feedbackstate.ReceiptSchema,
		"report": map[string]any{
			"schema": feedbackstate.ResolutionSchema,
			"feedback": map[string]any{"commit_sha": sha, "repository": repository,
				"decision": decision, "next_operation": nextOperation},
			"source_decision": decision, "decision": decision, "reason": "fixture",
			"from_resolution": "exact_operation", "to_resolution": "exact_operation",
			"previous_descents": 0, "descents": 0, "repository_writes": 0,
			"report_digest": reportDigest,
		},
		"replay_report_digest": reportDigest, "replay_verified": true,
		"repository_writes": 0, "receipt_digest": receiptDigest,
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data, receiptDigest
}
