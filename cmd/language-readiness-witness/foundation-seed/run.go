package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/foundationseed"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.input == "" || cfg.expectedHead == "" || cfg.output == "" {
		return fmt.Errorf("root, input, expected-head, and output are required")
	}
	if err := requireOutside(cfg.root, cfg.output); err != nil {
		return err
	}
	raw, err := os.ReadFile(cfg.input)
	if err != nil {
		return err
	}
	var resolution predecessorresolution.Report
	if err := decodeStrict(raw, &resolution); err != nil {
		return err
	}
	report := foundationseed.Evaluate(resolution, cfg.expectedHead)
	if err := foundationseed.Validate(report); err != nil {
		return err
	}
	encoded, err := encodeJSON(report)
	if err != nil {
		return err
	}
	if err := writeExclusive(cfg.output, encoded); err != nil {
		return err
	}
	if cfg.check && report.Decision != foundationseed.DecisionAuthorized {
		return fmt.Errorf("%s: %s", report.Decision, report.Reason)
	}
	fmt.Printf("readiness-foundation: decision=%s coordinates=%d/%d "+
		"attempts=%d/%d writes=%d delta_claims=%d digest=%s\n",
		report.Decision, report.Summary.Satisfied, report.Summary.Total,
		report.Source.MissingAttempts, report.Source.SearchLimit,
		report.Source.RepositoryWrites, report.Source.ReadinessDeltaClaims,
		report.Digest)
	return nil
}
