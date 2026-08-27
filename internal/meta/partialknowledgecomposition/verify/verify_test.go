package verify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifierRejectsReceiptWhenSourceObservationIsTampered(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	logicalSourcePath := "examples/partial-knowledge-composition/main.gooo"
	sourcePath := filepath.Join(root, logicalSourcePath)
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	input := Input{Repository: "kimjooyoon/meta-ontology-go", HeadSHA: "0123456789abcdef0123456789abcdef01234567", SourcePath: logicalSourcePath, Source: source, InterventionMode: "none"}
	model, err := parseSource(logicalSourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	intervention, err := applyIntervention(&model, input.InterventionMode)
	if err != nil {
		t.Fatal(err)
	}
	receipt := reconstruct(input, model, intervention)
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
	if _, err := Verify(Input{Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath, Source: tampered, InterventionMode: input.InterventionMode, Receipt: receiptRaw}); err == nil {
		t.Fatal("source observation tampering was accepted")
	}
}
