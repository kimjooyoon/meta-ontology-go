package main

import (
	"bytes"
	"go/format"
	"os"
	"testing"
)

func TestEmitGo127CIFormatReceipts(t *testing.T) {
	paths := []string{"fixture_test.go", "run.go", "run_test.go"}
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
