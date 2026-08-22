package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/predecessorbinding"
)

func TestScanDistinguishesStaticAndParameterBindings(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(predecessorbinding.SourcePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	source := `package artifact
type BaselineReference struct { RunID int64; ArtifactName, HeadSHA, FileSHA256, ArtifactDigest, SnapshotDigest string; Completed, BasisPoints int }
type Selection struct { RunID int64; ArtifactName, HeadSHA, FileSHA256, ArtifactDigest, SnapshotDigest string; Completed, BasisPoints int }
func FoundationBaseline(selected Selection) BaselineReference {
	return BaselineReference{RunID: 1, ArtifactName: "fixed", HeadSHA: selected.HeadSHA,
		FileSHA256: selected.FileSHA256, ArtifactDigest: selected.ArtifactDigest,
		SnapshotDigest: selected.SnapshotDigest, Completed: selected.Completed,
		BasisPoints: selected.BasisPoints}
}`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	observed, err := scan(root)
	if err != nil {
		t.Fatal(err)
	}
	report := predecessorbinding.Evaluate(testSHA, observed, 0)
	if report.Summary.StaticLiteral != 2 || report.Summary.DynamicInput != 6 ||
		report.Summary.Unknown != 0 {
		t.Fatalf("unexpected scan: %+v", report.Summary)
	}
}

const testSHA = "2222222222222222222222222222222222222222"
