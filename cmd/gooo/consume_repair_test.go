package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

type consumeRepairReader struct {
	data []byte
}

func (r consumeRepairReader) ReadFile(string) ([]byte, error) {
	return r.data, nil
}

func TestRunConsumeRepairWritesDeferredHandoff(t *testing.T) {
	baseline := valueexecution.Execution{Scope: "scope", PlanDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", InputDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", ExecutionDigest: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}
	candidate := baseline
	candidate.ExecutionDigest = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	repair, err := valueexecution.ProposeRepair(valueexecution.CompareReplay(baseline, candidate))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(repair)
	if err != nil {
		t.Fatal(err)
	}
	outputDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := runConsumeRepair([]string{"repair-candidate.json", "--out", outputDir}, consumeRepairReader{data: data}, &stdout, &stderr); code != exitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
	handoffData, err := os.ReadFile(filepath.Join(outputDir, "repair-handoff.json"))
	if err != nil {
		t.Fatal(err)
	}
	var handoff valueexecution.RepairHandoff
	if err := json.Unmarshal(handoffData, &handoff); err != nil {
		t.Fatal(err)
	}
	if handoff.Status != valueexecution.RepairHandoffDeferred || handoff.ExecutionAllowed || handoff.RepositoryWrites != 0 {
		t.Fatalf("handoff = %#v", handoff)
	}
}

func TestRunConsumeRepairRejectsInvalidCandidate(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runConsumeRepair([]string{"repair-candidate.json", "--out", t.TempDir()}, consumeRepairReader{data: []byte(`{"schema":"invalid"}`)}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
}
