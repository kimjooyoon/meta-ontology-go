package main

import (
	"os"
	"path/filepath"

	activation "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityactivation"
)

func readInput(opts options) (activation.Input, error) {
	assurance, err := os.ReadFile(opts.assurance)
	if err != nil {
		return activation.Input{}, err
	}
	upstream, err := os.ReadFile(opts.upstream)
	if err != nil {
		return activation.Input{}, err
	}
	eligibility, err := os.ReadFile(opts.eligibility)
	if err != nil {
		return activation.Input{}, err
	}
	return activation.Input{SubjectSHA: opts.subjectSHA, Assurance: assurance, Upstream: upstream, Eligibility: eligibility}, nil
}

func writeOutput(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
