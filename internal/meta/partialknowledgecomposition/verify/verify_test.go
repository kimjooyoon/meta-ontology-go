package verify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/provider"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
)

const testHeadSHA = "0123456789abcdef0123456789abcdef01234567"

func TestVerifierRejectsReceiptWhenSourceObservationIsTampered(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	logicalSourcePath := "examples/partial-knowledge-composition/main.gooo"
	source, err := os.ReadFile(filepath.Join(root, logicalSourcePath))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := provider.Observe(provider.Input{
		Repository: "kimjooyoon/meta-ontology-go", HeadSHA: testHeadSHA,
		SourcePath: logicalSourcePath, Source: source,
		BeforeTracked: []byte("examples/partial-knowledge-composition/main.gooo\n"),
		AfterTracked:  []byte("examples/partial-knowledge-composition/main.gooo\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	rawJSON, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	produced, err := meta.Evaluate(meta.Input{Repository: "kimjooyoon/meta-ontology-go", HeadSHA: testHeadSHA, SourcePath: logicalSourcePath, Source: source, RawEvidence: rawJSON})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := json.Marshal(produced)
	if err != nil {
		t.Fatal(err)
	}

	tampered := []byte(strings.Replace(string(source), "left.observation_recipe=missing", "left.observation_recipe=exact", 1))
	if _, err := independent.Verify(independent.Input{
		Repository: "kimjooyoon/meta-ontology-go", HeadSHA: testHeadSHA, SourcePath: logicalSourcePath,
		Source: tampered, RawEvidence: rawJSON, Receipt: receipt,
	}); err == nil {
		t.Fatal("source observation tampering was accepted")
	}
}
