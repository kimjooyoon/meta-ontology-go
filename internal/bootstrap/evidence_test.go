package bootstrap

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoHostedBaselineMatchesCheckedInFixture(t *testing.T) {
	source := readExample(t, "examples/bootstrap/main.gooo")
	expected := NewGoHostedBaseline(source)
	fixture := readExample(t, "examples/bootstrap/go-hosted-baseline.json")
	var observed Evidence
	if err := json.Unmarshal(fixture, &observed); err != nil {
		t.Fatal(err)
	}
	if err := observed.Validate(); err != nil {
		t.Fatal(err)
	}
	want, err := expected.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	got, err := observed.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		t.Fatalf("fixture does not describe the Go-hosted baseline:\nwant %s\ngot  %s", want, got)
	}
	if observed.SemanticDigest != nil || observed.ProvenanceDigest != nil || observed.PromotionEligible {
		t.Fatal("deferred baseline claimed unavailable or promotable evidence")
	}
}

func TestGoHostedEvidenceDigestIsReproducible(t *testing.T) {
	source := []byte("package bootstrap\nentity Source\n")
	first := NewGoHostedBaseline(source)
	second := NewGoHostedBaseline(source)
	firstJSON, err := first.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := second.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) || !strings.HasSuffix(string(firstJSON), "\n") {
		t.Fatal("identical Go-hosted evidence was not canonical")
	}
	firstDigest, err := first.Digest()
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := second.Digest()
	if err != nil || firstDigest != secondDigest {
		t.Fatalf("identical evidence has different digests: %v", err)
	}
	changed := NewGoHostedBaseline(append(source, []byte("activity Verify\n")...))
	changedDigest, err := changed.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if changedDigest == firstDigest {
		t.Fatal("source change did not change evidence digest")
	}
}

func TestDeferredEvidenceCannotBecomePromotable(t *testing.T) {
	evidence := NewGoHostedBaseline([]byte("source"))
	evidence.PromotionEligible = true
	if err := evidence.Validate(); err == nil {
		t.Fatal("deferred evidence was promoted without semantic or provenance proof")
	}
	evidence = NewGoHostedBaseline([]byte("source"))
	semantic := DigestBytes([]byte("semantic"))
	evidence.SemanticDigest = &semantic
	evidence.Checks.SemanticCLI = StatusPass
	evidence.Decision = StatusPass
	evidence.EvidenceStatus = StatusPass
	evidence.PromotionEligible = true
	if err := evidence.Validate(); err == nil {
		t.Fatal("promotion without provenance evidence was accepted")
	}
}

func TestEvidenceRejectsMalformedDigest(t *testing.T) {
	evidence := NewGoHostedBaseline([]byte("source"))
	bad := "not-a-digest"
	evidence.SourceDigest = &bad
	if err := evidence.Validate(); err == nil {
		t.Fatal("malformed source digest was accepted")
	}
}

func readExample(t *testing.T, relativePath string) []byte {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate test source")
	}
	root := filepath.Join(filepath.Dir(sourceFile), "..", "..")
	data, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
