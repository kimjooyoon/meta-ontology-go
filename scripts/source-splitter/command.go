package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	if err := verifyRepositorySHA(cfg.root, cfg.sha); err != nil {
		return err
	}
	report, err := metricevidence.Load(cfg.metrics, cfg.sha)
	if err != nil {
		return err
	}
	indicators := report.GoSplitIndicators()
	if cfg.plan != "" {
		return runBatch(cfg, report, indicators)
	}
	if cfg.check {
		return checkSplitPlans(cfg, report, indicators)
	}
	if !metricevidence.Contains(indicators, cfg.subject) {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	plan, err := planSource(cfg.root, cfg.subject, report.Meta.Policy.MaxFileLines)
	if err != nil {
		return err
	}
	if err := validateTopology(report, plan); err != nil {
		return err
	}
	if cfg.evidenceJSON {
		return applySplitWithEvidence(cfg, plan, os.Stdout)
	}
	if err := applySplit(plan); err != nil {
		return err
	}
	fmt.Printf("source-splitter: subject=%s outputs=%d write=true\n", cfg.subject, len(plan.Parts))
	return nil
}
