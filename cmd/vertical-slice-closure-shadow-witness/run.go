package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/verticalsliceclosureshadow"
)

func run(cfg config) error {
	if cfg.output == "" || cfg.head == "" {
		return fmt.Errorf("output and head are required")
	}
	input := verticalsliceclosureshadow.Input{HeadSHA: cfg.head, Artifacts: map[string][]byte{}}
	if !cfg.unavailableAssurance {
		raw, err := readRequired(cfg.assurance, "assurance")
		if err != nil {
			return err
		}
		input.Assurance = raw
	}
	paths := map[string]string{"syntax": cfg.syntax, "semantics": cfg.semantics,
		"binding": cfg.binding, "use-cases": cfg.useCases,
		"toolchain": cfg.toolchain, "release": cfg.release}
	if cfg.unavailableBoundary != "" {
		if _, ok := paths[cfg.unavailableBoundary]; !ok {
			return fmt.Errorf("unknown unavailable boundary %q", cfg.unavailableBoundary)
		}
	}
	for id, path := range paths {
		if id == cfg.unavailableBoundary {
			continue
		}
		raw, err := readRequired(path, id)
		if err != nil {
			return err
		}
		input.Artifacts[id] = raw
	}
	report := verticalsliceclosureshadow.Evaluate(input)
	if err := verticalsliceclosureshadow.Validate(report); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.output, verticalsliceclosureshadow.Encode(report), 0o644); err != nil {
		return err
	}
	fmt.Printf("vertical-slice-shadow: decision=%s resolution=%s boundaries=%d/%d links=%d/%d projected=%d/%d writes=%d digest=%s\n",
		report.Decision, report.Resolution, report.Summary.BoundariesSatisfied,
		report.Summary.BoundariesTotal, report.Summary.LinksSatisfied,
		report.Summary.LinksTotal, report.Summary.ProjectedOperating,
		report.Summary.DenominatorTotal, report.RepositoryWrites, report.ReportDigest)
	return nil
}

func readRequired(path, label string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("%s path is required", label)
	}
	return os.ReadFile(path)
}
