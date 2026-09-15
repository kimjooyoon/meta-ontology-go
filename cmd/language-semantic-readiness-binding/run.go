package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemanticbinding"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	report, err := languagesemanticbinding.Evaluate(languagesemanticbinding.Input{
		ExpectedHeadSHA: options.head,
		ReadinessPath:   options.readiness,
		ConceptPath:     options.concept,
		SemanticPath:    options.semantic,
	})
	if err != nil {
		return err
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if options.check != "" {
		expected, err := os.ReadFile(options.check)
		if err != nil {
			return err
		}
		if !bytes.Equal(payload, expected) {
			return fmt.Errorf("semantic readiness binding replay mismatch")
		}
		return nil
	}
	if err := requireOutsideRepository(options.root, options.output); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(options.output, payload, 0o644)
}
