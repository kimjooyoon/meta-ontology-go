package main

import (
	"errors"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
)

// Regenerate the evaluator from its Gooo source; generated regions are never
// patched by hand. Run from the repository root with canonical relative paths.
func runGeneratePolicy(contractPath, outputPath string) error {
	if contractPath == "" || outputPath == "" || contractPath == outputPath {
		return errors.New("generate requires distinct contract and output paths")
	}
	source, err := os.ReadFile(contractPath)
	if err != nil {
		return err
	}
	_, generated, err := compatibilitypolicy.GenerateNamed(contractPath, source)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, generated, 0o644)
}
