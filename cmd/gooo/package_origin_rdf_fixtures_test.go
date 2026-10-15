package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedRDFFirstObservationBindings(t *testing.T) {
	const root = "../.."
	const directory = "docs/observations/qualified-usage"
	raw, err := os.ReadFile(filepath.Join(root, directory, "observations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var observation struct {
		Source  string            `json:"source_sha256"`
		Files   map[string]string `json:"public_files"`
		Archive struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"archive"`
	}
	if err := json.Unmarshal(raw, &observation); err != nil || len(observation.Files) < 10 {
		t.Fatal("source/input identities missing", err)
	}
	for name, want := range observation.Files {
		checkRecordedRDFFixtureDigest(t, filepath.Join(root, name), want)
	}
	checkRecordedRDFFixtureDigest(t, filepath.Join(root, directory, observation.Archive.Path), observation.Archive.SHA256)
	if observation.Source != observation.Files["examples/recorded-use-origins/inspect.gooo"] {
		t.Fatal("published source lost its first-observation binding")
	}
}

func checkRecordedRDFFixtureDigest(t *testing.T, name, want string) {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
		t.Fatal("first source/input/archive bytes changed", name, err)
	}
}
