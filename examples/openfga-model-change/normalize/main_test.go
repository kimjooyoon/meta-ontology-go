package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestRetainedObservationNormalizationFeedsTheActualCases(t *testing.T) {
	output := filepath.Join(t.TempDir(), "normalized.json")
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"normalize", "../raw-observations.json.gz", output}
	main()
	read := func(path string) []byte {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	var actual, retained map[string]any
	if json.Unmarshal(read(output), &actual) != nil ||
		json.Unmarshal(read("../normalized-observations.json"), &retained) != nil {
		t.Fatal("normalization must produce readable JSON")
	}
	retained["normalizer_go_version"] = runtime.Version()
	if !reflect.DeepEqual(actual, retained) {
		t.Fatal("retained normalization differs from the original observation bytes")
	}
	var suite struct {
		Cases []struct {
			Inputs map[string]any `json:"inputs"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(read("../observed-cases.json"), &suite); err != nil {
		t.Fatal(err)
	}
	pairs := actual["pairs"].([]any)
	if len(pairs) != 5 || len(suite.Cases) != len(pairs) {
		t.Fatal("exactly five retained observation pairs are required")
	}
	for index, pair := range pairs {
		if !reflect.DeepEqual(pair, suite.Cases[index].Inputs["Compare"]) {
			t.Fatalf("case %d no longer consumes the normalized observation", index)
		}
	}
}

func TestChangedObservationBytesAreRejectedBeforeOutput(t *testing.T) {
	raw, err := readSnapshot("../raw-observations.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(raw, []byte(`"allowed": true`), []byte(`"allowed":false`), 1)
	if bytes.Equal(raw, changed) {
		t.Fatal("counterexample did not change an observation")
	}
	root := t.TempDir()
	input, output := filepath.Join(root, "changed.json"), filepath.Join(root, "normalized.json")
	if err := os.WriteFile(input, changed, 0600); err != nil {
		t.Fatal(err)
	}
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"normalize", input, output}
	defer func() {
		if recover() == nil {
			t.Fatal("changed observation origin was accepted")
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatal("rejected observation wrote a normalized output", err)
		}
	}()
	main()
}
