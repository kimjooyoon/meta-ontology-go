// This example creates a synthetic option-bias artifact to inspect the adapter.
// It performs no fitting and produces no language-understanding quality claim.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriadecision"
)

func main() {
	out := flag.String("out", "out/scalar-source-model-synthetic", "new directory for a synthetic model")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "use --out with a new directory")
		os.Exit(2)
	}
	if err := writeSyntheticModel(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeSyntheticModel(dir string) error {
	var parameters [criteriadecision.ParameterCount]float32
	parameters[criteriadecision.ParameterCount-criteriadecision.HiddenDim-1] = 3
	metadata, weights, err := criteriadecision.EncodeArtifact(&parameters)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf("synthetic artifact requires a new directory: %w", err)
	}
	for name, raw := range map[string][]byte{"metadata.json": metadata, "weights.bin": weights} {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			return err
		}
	}
	return nil
}
