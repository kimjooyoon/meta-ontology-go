package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.head == "" || cfg.registry == "" || cfg.concept == "" {
		return fmt.Errorf("root, head, registry, and concept-artifact are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	for _, external := range []string{cfg.concept, target} {
		outside, err := outsideRoot(cfg.root, external)
		if err != nil || !outside {
			return fmt.Errorf("toolchain use case artifacts must be outside the repository root")
		}
	}
	if cfg.check != "" {
		return consume(cfg, stdout)
	}
	return produce(cfg, stdout)
}

func build(cfg config) (toolchainusecases.Report, error) {
	registryRaw, err := os.ReadFile(cfg.registry)
	if err != nil {
		return toolchainusecases.Report{}, err
	}
	conceptRaw, err := os.ReadFile(cfg.concept)
	if err != nil {
		return toolchainusecases.Report{}, err
	}
	artifact := languageconcept.Artifact{}
	if err := json.Unmarshal(conceptRaw, &artifact); err != nil {
		return toolchainusecases.Report{}, err
	}
	report := toolchainusecases.Evaluate(os.DirFS(cfg.root), cfg.head, registryRaw, artifact)
	if err := toolchainusecases.Validate(report, cfg.head); err != nil {
		return report, err
	}
	if report.Decision != toolchainusecases.DecisionPass {
		return report, fmt.Errorf("%s: %s", report.Decision, report.Reason)
	}
	return report, nil
}

func printSummary(output io.Writer, report toolchainusecases.Report) {
	fmt.Fprintf(output, "toolchain-use-cases: decision=%s resolution=%s satisfied=%d/%d writes=%d\n",
		report.Decision, report.Resolution, report.Summary.Satisfied,
		report.Summary.Total, report.RepositoryWrites)
}
