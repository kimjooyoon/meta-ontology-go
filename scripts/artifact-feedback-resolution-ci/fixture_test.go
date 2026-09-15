package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
)

func writeSourceFixtures(t *testing.T, root string) (string, string) {
	t.Helper()
	head := strings.Repeat("a", 40)
	coverage := artifactcoverage.Report{
		Schema: artifactcoverage.ReportSchema, CommitSHA: head,
		Repository: "kimjooyoon/meta-ontology-go", Decision: "FIXED_POINT",
		Summary: artifactcoverage.Summary{
			RequiredOperations: 6, CanonicalOperations: 6,
		},
	}
	coverage.ReportDigest = fixtureDigest(t, coverage)
	provenance := provenanceEnvelope{
		SchemaVersion: "v1", HeadSHA: head, Decision: "BOUND",
		Reason:         "ARTIFACT_PROVENANCE_BOUND",
		EnvelopeDigest: strings.Repeat("1", 64),
		ReplayDigest:   strings.Repeat("2", 64),
	}
	coveragePath := filepath.Join(root, "coverage.json")
	provenancePath := filepath.Join(root, "provenance.json")
	writeFixtureJSON(t, coveragePath, coverage)
	writeFixtureJSON(t, provenancePath, provenance)
	return coveragePath, provenancePath
}

func writeFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixtureDigest(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
