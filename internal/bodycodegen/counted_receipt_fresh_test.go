package bodycodegen

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCountedCompletenessRecord18811(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal/bodycodegen/testdata/counterfill18717.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	read := func() Report {
		var saved struct {
			BodyFills []struct {
				Generation Result `json:"generation"`
			} `json:"body_fills"`
		}
		if err := json.Unmarshal(raw, &saved); err != nil || len(saved.BodyFills) != 1 {
			t.Fatal("saved counted source-fill", err)
		}
		return saved.BodyFills[0].Generation.Report
	}
	if err := compareSourceFillCompleteness(read()); err != nil {
		t.Fatal("retained complete receipt", err)
	}
	changes := []func(*Report){func(r *Report) { r.CompletenessReceipt = nil }, func(r *Report) { r.SourceDigest = "different-source" }, func(r *Report) { r.GeneratedDigest = "different-projection" }}
	for i, change := range changes {
		report := read()
		change(&report)
		if err := compareSourceFillCompleteness(report); err == nil {
			t.Fatal("contradictory bound receipt accepted", i)
		}
	}
	t.Log("saved new counted assembly receipt accepted;three bound receipt contradictions rejected;JSON comparison separated without dropping checks;producer/model0")
}
