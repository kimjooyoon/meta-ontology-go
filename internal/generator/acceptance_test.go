package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

const acceptanceFingerprint = "66cc8b603ecded2a22a2ba4e0ad07905c719c4ceb6e803be64b6687e093a50da"

func TestAcceptanceFixtureReproducibility(t *testing.T) {
	// GEN-DET-001: repeated projection bytes and evidence hashes are stable.
	ir := acceptanceFixture()
	first := mustAcceptanceResult(t, ir, nil)
	second := mustAcceptanceResult(t, ir, nil)
	firstFingerprint := acceptanceResultFingerprint(t, first)
	secondFingerprint := acceptanceResultFingerprint(t, second)
	if firstFingerprint != secondFingerprint {
		t.Fatalf("reproducibility fingerprint changed: %s != %s", firstFingerprint, secondFingerprint)
	}
	if acceptanceFingerprint != "" && firstFingerprint != acceptanceFingerprint {
		t.Fatalf("fixture fingerprint changed: got %s want %s", firstFingerprint, acceptanceFingerprint)
	}
	firstEvidence := acceptanceEvidence(first, "pass")
	secondEvidence := acceptanceEvidence(second, "pass")
	if err := verify.CompareEvidence(firstEvidence, secondEvidence); err != nil {
		t.Fatalf("repeated Go evidence changed: %v", err)
	}
}

func TestAcceptanceFixtureLocalityUsesStableIDs(t *testing.T) {
	ir := acceptanceFixture()
	first := mustAcceptanceResult(t, ir, nil)
	previous := bytes.Replace(first.Source, []byte("package bootstrapgen\n"), []byte("package bootstrapgen\n\nvar Keep = 7\n"), 1)
	changed := acceptanceFixture()
	changed.Activities[0].Name = "CompileBootstrap"
	changed.Activities[0].GoName = "CompileBootstrap"
	second := mustAcceptanceResult(t, changed, previous)
	if !strings.Contains(string(second.Source), "var Keep = 7") {
		t.Fatal("marker-outside text was not preserved")
	}
	if !bytes.Equal(testGeneratedBlock(t, first.Source, "gooo://entity/source"), testGeneratedBlock(t, second.Source, "gooo://entity/source")) {
		t.Fatal("unrelated entity region changed")
	}
	if !bytes.Equal(testGeneratedBlock(t, first.Source, "gooo://activity/inspect"), testGeneratedBlock(t, second.Source, "gooo://activity/inspect")) {
		t.Fatal("unrelated activity region changed")
	}
	if !strings.Contains(string(second.Source), `//gooo:generated:start id="gooo://activity/compile"`) {
		t.Fatal("stable activity ID was not retained in its marker")
	}
	if len(second.SourceMap.Lookup("gooo://activity/compile")) != 1 {
		t.Fatal("changed activity lost its source-map identity")
	}
}

func TestAcceptanceFixtureDefersGoooEvidence(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	goEvidence := acceptanceEvidence(result, "pass")
	goooEvidence := acceptanceEvidence(result, "DEFERRED")
	goooEvidence.Producer = verify.EvidenceProducerGooo
	if goooEvidence.Bundle.Decision != "DEFERRED" {
		t.Fatal("unimplemented gooo stage was not marked DEFERRED")
	}
	if err := verify.CompareEvidence(goEvidence, goooEvidence); err == nil {
		t.Fatal("deferred gooo evidence was accepted as parity")
	}
}

func TestAcceptanceFixtureRejectsBrokenRollbackMarkers(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	broken := strings.Replace(string(result.Source), `//gooo:generated:end id="gooo://activity/compile"`, `//gooo:generated:end id="wrong"`, 1)
	if _, err := Generate(acceptanceFixture(), []byte(broken)); err == nil {
		t.Fatal("broken rollback marker topology was accepted")
	}
}

func mustAcceptanceResult(t *testing.T, ir SemanticIR, previous []byte) Result {
	t.Helper()
	result, err := Generate(ir, previous)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func acceptanceResultFingerprint(t *testing.T, result Result) string {
	t.Helper()
	payload, err := json.Marshal(struct {
		Source    []byte
		SourceMap SourceMap
	}{Source: result.Source, SourceMap: result.SourceMap})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func acceptanceEvidence(result Result, decision string) verify.EvidenceArtifact {
	digest := sha256.Sum256(result.Source)
	return verify.EvidenceArtifact{
		Producer: verify.EvidenceProducerGo,
		Bundle: verify.EvidenceBundle{
			Schema: verify.EvidenceSchemaVersion, Stage: verify.StageGoBaseline,
			Fixture: "generator/acceptance", Decision: decision,
			Facts: []verify.EvidenceFact{{ID: "artifact", Kind: "sha256", Value: hex.EncodeToString(digest[:])}},
		},
	}
}
