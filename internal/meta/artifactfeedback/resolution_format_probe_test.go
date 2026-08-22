package artifactfeedback

import (
	"bytes"
	"go/format"
	"os"
	"testing"
)

func TestEmitGo127ResolutionFormatReceipts(t *testing.T) {
	paths := []string{"resolution_contract.go", "resolution_evaluate.go"}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := format.Source(source)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(source, canonical) {
			t.Errorf("GOFORMAT_CANONICAL %s %q", path, canonical)
		}
	}
}
