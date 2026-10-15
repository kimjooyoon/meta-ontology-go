package workspaceexecution

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestContinuationCarriesSavedFillOrigins6841(t *testing.T) {
	path := os.Getenv("GOOO_TRANSFER_SAVED_INPUT")
	if path == "" {
		path = "../../../docs/observations/value-transfers/transfer6373.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Manifest packageruntime.Manifest `json:"manifest"`
		Result   Result                  `json:"result"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	program, err := Prepare(saved.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	// Construct the continuation record only. Do not run the saved program or cases.
	continued, err := resumedWorkspaceResult(program, saved.Result, nil, saved.Result.Composition)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(saved.Result.FillOrigins)
	after, _ := json.Marshal(continued.FillOrigins)
	if string(before) != string(after) || continued.FillOriginIssue != "" || len(continued.FillOrigins) != 1 {
		t.Fatal("lost original candidate locations", string(after), continued.FillOriginIssue)
	}
	oversized := saved.Result
	oversized.BodyFills = make([]BodyFillStep, 129)
	origins, issue := collectWorkspaceFillOrigins(program, oversized)
	if len(origins) != 0 || issue == "" {
		t.Fatal("bounded omission needs an explicit issue", origins, issue)
	}
	t.Log("6841: continuation retains saved original fill sources; over-limit metadata reports an issue; source/candidate/model executions0")
}
