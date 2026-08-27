package main

import (
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicealgebra"
)

func produce(source, root, beforePath, afterPath string, baseline []byte) any {
	before := snapshot(root)
	data, err := os.ReadFile(source)
	if err != nil {
		return proofchoicealgebra.Evaluate(source, nil, before, nil, baseline)
	}
	receipt := proofchoicealgebra.Evaluate(source, data, before, nil, baseline)
	after := snapshot(root)
	writeSnapshot(beforePath, before)
	writeSnapshot(afterPath, after)
	receipt.Effects = proofchoicealgebra.ObserveEffects(before, after)
	sealed, err := proofchoicealgebra.Seal(receipt)
	if err != nil {
		return receipt
	}
	return sealed
}

func readOptional(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}
