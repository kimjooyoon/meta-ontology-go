package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFailureManifestRejectsTamperedClassificationAndHandoff(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Class = "gate"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure classification was accepted")
	}
	manifest.Class = "test"
	manifest.HandoffOwner = "gate"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure handoff owner was accepted")
	}
}
func TestFailureCatalogMatchesCheckedInDocument(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate failure catalog test source")
	}
	document, err := os.ReadFile(filepath.Join(filepath.Dir(source), "docs", "failure-reasons.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := "sha256:" + digestBytes(document); got != failureCatalogDigest {
		t.Fatalf("catalog digest is not bound to document bytes: got %s want %s", failureCatalogDigest, got)
	}
	if err := validateFailureCatalog(); err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for line := range strings.SplitSeq(string(document), "\n") {
		if !strings.HasPrefix(line, "| `CI-") {
			continue
		}
		parts := strings.Split(line, "`")
		if len(parts) < 2 {
			t.Fatalf("malformed catalog row: %s", line)
		}
		counts[parts[1]]++
	}
	if len(counts) != len(failureCatalogRecords) {
		t.Fatalf("catalog code count mismatch: docs=%d machine=%d", len(counts), len(failureCatalogRecords))
	}
	for _, record := range failureCatalogRecords {
		if counts[record.Code] != 1 {
			t.Fatalf("catalog code %s is missing or duplicated: %d", record.Code, counts[record.Code])
		}
	}
	for code := range counts {
		if _, ok := failureCatalog[code]; !ok {
			t.Fatalf("catalog contains unknown code %s", code)
		}
	}
}
