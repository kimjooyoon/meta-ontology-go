package generation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestObservationOwnershipSeparatesReplay(t *testing.T) {
	root := t.TempDir()
	plan := filepath.Join(root, "plan.json")
	first := ObservationBundlePath(plan, filepath.Join(root, "self-improvement-execution.json"))
	replay := ObservationBundlePath(plan, filepath.Join(root, "self-improvement-execution-replay.json"))
	if first == replay || first != filepath.Join(root, "meta-operation-observations.json") {
		t.Fatalf("ambiguous or incompatible ownership: first=%s replay=%s", first, replay)
	}
	if err := os.WriteFile(first, []byte("first observation"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(replay, []byte("replay observation"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(first)
	if err != nil || string(got) != "first observation" {
		t.Fatalf("replay replaced first observation: %q %v", got, err)
	}
}

func TestManifestObservationOwnershipPreservesDirectoryIdentity(t *testing.T) {
	plan := filepath.Join("input", "plan.json")
	first := ObservationBundlePath(plan, filepath.Join("input", "self-improvement-execution.json"))
	other := ObservationBundlePath(plan, filepath.Join("other", "self-improvement-execution.json"))
	if first == other || other != filepath.Join("other", "self-improvement-execution.json.observations.json") {
		t.Fatalf("different manifest directories share evidence: %s %s", first, other)
	}
}
