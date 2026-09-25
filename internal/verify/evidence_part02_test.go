package verify

import (
	"bytes"
	"strings"
	"testing"
)

func TestEvidenceManifestJSONIsStable(t *testing.T) {
	evidence := sampleEvidence(EvidenceProducerGo, nil)
	first, err := evidence.ManifestJSON()
	if err != nil {
		t.Fatal(err)
	}
	second, err := evidence.ManifestJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || !bytes.HasSuffix(first, []byte{'\n'}) {
		t.Fatalf("manifest was not deterministic JSONL: %q", first)
	}
	if !strings.Contains(string(first), `"producer":"go"`) {
		t.Fatalf("manifest omitted producer identity: %s", first)
	}
}
func sampleEvidence(producer string, facts []EvidenceFact) EvidenceArtifact {
	return EvidenceArtifact{
		Producer: producer,
		Bundle: EvidenceBundle{
			Schema:   EvidenceSchemaVersion,
			Stage:    StageDualEvidence,
			Fixture:  "billing/main.gooo",
			Decision: "pass",
			Facts:    facts,
		},
	}
}
