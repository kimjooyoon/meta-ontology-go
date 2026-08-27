package verify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
)

func TestVerifierRejectsReceiptWhenSourceObservationIsTampered(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	sourcePath := filepath.Join(root, meta.SourcePath)
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	input := meta.Input{Repository: "kimjooyoon/meta-ontology-go", HeadSHA: "0123456789abcdef0123456789abcdef01234567", SourcePath: meta.SourcePath, Source: source, Intervention: meta.InterventionNone}
	receipt, err := meta.Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	receiptRaw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}

	tampered := append([]byte(nil), source...)
	needle := []byte("left.observed_available=false")
	for index := 0; index+len(needle) <= len(tampered); index++ {
		match := true
		for offset := range needle {
			if tampered[index+offset] != needle[offset] {
				match = false
				break
			}
		}
		if match {
			tampered[index] = 't'
			break
		}
	}
	if _, err := independent.Verify(independent.Input{Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath, Source: tampered, InterventionMode: string(meta.InterventionNone), Receipt: receiptRaw}); err == nil {
		t.Fatal("source observation tampering was accepted")
	}
}
