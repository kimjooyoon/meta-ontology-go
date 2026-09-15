package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const capturedJournalHead = "d7635e11a2e1c70cf9cd1ac1fe38436cc51802ad"
const capturedReplayInvocation = "meta-execution-234709-1789064231986966238"

func capturedNativeJournal(t *testing.T, name string) []byte {
	t.Helper()
	dir := filepath.Join("testdata", "native-interruption-34510267647")
	metadata, err := os.ReadFile(filepath.Join(dir, "capture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var capture map[string]any
	if err := json.Unmarshal(metadata, &capture); err != nil {
		t.Fatal(err)
	}
	if capture["schema"] != "gooo/native-journal-capture/v1" || capture["head_sha"] != capturedJournalHead ||
		capture["run_id"] != float64(34510267647) || capture["artifact_id"] != float64(10167275876) {
		t.Fatalf("unexpected capture identity: %v", capture)
	}
	files, ok := capture["files"].(map[string]any)
	if !ok {
		t.Fatal("capture has no file bindings")
	}
	binding, ok := files[name].(map[string]any)
	if !ok {
		t.Fatalf("capture has no binding for %s", name)
	}
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); binding["sha256"] != got {
		t.Fatalf("captured bytes changed: %s", name)
	}
	return data
}
