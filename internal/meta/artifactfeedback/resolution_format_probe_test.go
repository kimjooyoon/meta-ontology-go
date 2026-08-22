package artifactfeedback

import (
	"bytes"
	"go/format"
	"os"
	"testing"
)

func TestEmitGo127ResolutionFormatReceipts(t *testing.T) {
	paths := []string{"resolution_indicators.go", "resolution_malformed_test.go", "resolution_preservation_test.go", "resolution_types.go", "resolution_unknown_test.go"}
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
