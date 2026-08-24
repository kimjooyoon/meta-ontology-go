package main

import (
	"os"
	"path/filepath"

	activation "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/externalconformanceactivation"
)

func readInput(opts options) (activation.Input, error) {
	assurance, err := os.ReadFile(opts.assurance)
	if err != nil {
		return activation.Input{}, err
	}
	eligibility, err := os.ReadFile(opts.eligibility)
	if err != nil {
		return activation.Input{}, err
	}
	merge, err := os.ReadFile(opts.merge)
	if err != nil {
		return activation.Input{}, err
	}
	return activation.Input{SubjectSHA: opts.subjectSHA, Assurance: assurance, Eligibility: eligibility, Merge: merge}, nil
}

func writeOutput(fileName string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fileName, raw, 0o644)
}
