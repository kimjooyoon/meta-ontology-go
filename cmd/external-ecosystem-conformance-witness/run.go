package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	conformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalecosystemconformance"
)

func run(arguments []string) error {
	opts, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	capsule, err := conformance.Reference()
	if err != nil {
		return err
	}
	readme, err := os.ReadFile(opts.ReadmePath)
	if err != nil {
		return fmt.Errorf("read upstream README: %w", err)
	}
	goMod, err := os.ReadFile(opts.GoModPath)
	if err != nil {
		return fmt.Errorf("read upstream go.mod: %w", err)
	}
	evidence := conformance.Evidence{Readme: readme, GoMod: goMod}
	var receipt any
	if opts.CaseID == "suite" {
		receipt = conformance.RunSuite(opts.SubjectSHA, capsule, evidence)
	} else {
		receipt = conformance.RunCase(opts.SubjectSHA, opts.CaseID, capsule, evidence)
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(opts.OutputPath, append(raw, '\n'), 0o644)
}
