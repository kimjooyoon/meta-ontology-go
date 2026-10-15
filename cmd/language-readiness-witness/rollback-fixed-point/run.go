package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.guard == "" || cfg.transformation == "" ||
		cfg.expectedSHA == "" || cfg.output == "" {
		return fmt.Errorf("root, guard, transformation, expected-sha, and output are required")
	}
	if err := requireExternal(cfg.root, cfg.output); err != nil {
		return err
	}
	source := rollbackfixedpoint.Collect(cfg.guard, cfg.transformation, cfg.expectedSHA)
	return writeReport(cfg, rollbackfixedpoint.Build(source))
}

func writeReport(cfg config, report rollbackfixedpoint.Report) error {
	if err := rollbackfixedpoint.Validate(report); err != nil {
		return err
	}
	payload := rollbackfixedpoint.Encode(report)
	var err error
	if cfg.check {
		err = compareFile(cfg.output, payload)
	} else {
		err = writeExclusive(cfg.output, payload)
	}
	if err != nil {
		return err
	}
	if report.Decision != rollbackfixedpoint.DecisionPass &&
		!rollbackfixedpoint.IsKnownMixedTerminal(report) && !rollbackfixedpoint.IsKnownAppliedTerminal(report) {
		return fmt.Errorf("rollback fixed-point decision = %s reason = %s",
			report.Decision, report.Reason)
	}
	if rollbackfixedpoint.IsKnownAppliedTerminal(report) {
		fmt.Println("rollback-fixed-point: observed=SANDBOX_APPLIED_NON_PROMOTING recovery=false promotion=false")
	}
	fmt.Printf("rollback-fixed-point: mode=%s coordinates=%d/%d bps=%d writes=%d digest=%s\n",
		report.Mode, report.Summary.Satisfied, report.Summary.Total,
		report.Summary.ReadinessBPS, report.RepositoryWrites, report.ReportDigest)
	return nil
}
