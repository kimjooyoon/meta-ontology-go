package main

import (
	"path/filepath"
	"testing"
)

func TestCheckedInReceiptIsIndependentlyAdjudicated(t *testing.T) {
	root := filepath.Join("..", "..")
	if err := validate(filepath.Join(root, "examples/semantic-resolution-lattice/main.gooo"), filepath.Join(root, "examples/semantic-resolution-lattice/receipt.json")); err != nil {
		t.Fatal(err)
	}
}
