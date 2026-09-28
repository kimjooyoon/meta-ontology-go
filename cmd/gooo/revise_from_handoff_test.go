package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestRunReviseFromHandoffBindsHandoffProvenance(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "revision.gooo")
	source := []byte(`package revision
namespace revision
entity Integer id "gooo://revision/entity/integer"
activity Observe(Integer) -> Integer computes "int.add:1"
`)
	if err := os.WriteFile(sourcePath, source, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(source)
	sourceDigest := "sha256:" + hex.EncodeToString(sum[:])
	handoffPath := filepath.Join(root, "repair-handoff.json")
	handoff := valueexecution.RepairHandoff{
		Schema: valueexecution.RepairHandoffSchema, CandidateID: "gooo://repair-candidate/candidate1", CandidateDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ComparisonDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", TriggerState: valueexecution.ReplayRefuted,
		TriggerReason: "VALUE_INTEGER_OVERFLOW", NextOperation: "PRESERVE_COUNTEREXAMPLE", Status: valueexecution.RepairHandoffDeferred,
	}
	handoffData, err := json.Marshal(handoff)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(handoffPath, handoffData, 0o644); err != nil {
		t.Fatal(err)
	}
	outputDir := filepath.Join(root, "revision")
	var stdout, stderr bytes.Buffer
	args := []string{sourcePath, "--handoff", handoffPath, "--source-digest", sourceDigest, "--activity", "Observe", "--expected", "int.add:1", "--replace", "int.add:0", "--out", outputDir}
	if code := runReviseFromHandoff(args, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	revisionData, err := os.ReadFile(filepath.Join(outputDir, "revision.json"))
	if err != nil {
		t.Fatal(err)
	}
	var revision valueexecution.SourceRevision
	if err := json.Unmarshal(revisionData, &revision); err != nil {
		t.Fatal(err)
	}
	if revision.TriggerReason != handoff.TriggerReason || revision.RepairCandidateID != handoff.CandidateID || revision.RepairHandoffDigest == "" || revision.ExecutionAllowed || revision.RepositoryWrites != 0 {
		t.Fatalf("revision=%#v", revision)
	}
}
