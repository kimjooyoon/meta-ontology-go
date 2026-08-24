package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/rollbackintegrityshadow"
)

func run(cfg config) error {
	if cfg.output == "" {
		return fmt.Errorf("output is required")
	}
	var raw []byte
	var err error
	if !cfg.unavailable {
		if cfg.assurance == "" {
			return fmt.Errorf("assurance is required unless unavailable")
		}
		raw, err = os.ReadFile(cfg.assurance)
		if err != nil {
			return err
		}
	}
	report := rollbackintegrityshadow.Evaluate(raw)
	if err = rollbackintegrityshadow.Validate(report); err != nil {
		return err
	}
	if err = os.WriteFile(cfg.output, rollbackintegrityshadow.Encode(report), 0o644); err != nil {
		return err
	}
	fmt.Printf("rollback-integrity-shadow: decision=%s resolution=%s cases=%d/%d projected=%d/%d writes=%d digest=%s\n",
		report.Decision, report.Resolution, report.Summary.CasesPassed,
		report.Summary.CasesTotal, report.Summary.ProjectedOperating,
		report.Summary.DenominatorTotal, report.RepositoryWrites, report.ReportDigest)
	return nil
}
