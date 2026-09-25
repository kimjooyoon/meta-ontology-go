package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
	"testing"
)

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
