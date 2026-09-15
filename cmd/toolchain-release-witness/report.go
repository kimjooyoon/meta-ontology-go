package main

import (
	"fmt"
	"io"

	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func produceReport(cfg config, report release.Report,
	evidence []release.PlatformEvidence, stdout io.Writer) error {
	if err := release.WriteReport(cfg.output, report); err != nil {
		return err
	}
	if err := release.WriteBundle(cfg.bundle, evidence); err != nil {
		return err
	}
	if err := release.Validate(report, cfg.expectedHead); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "toolchain-release: %d/%d PASS/EXACT\n",
		report.Summary.CasesSatisfied, report.Summary.CasesTotal)
	return err
}

func checkReport(cfg config, expected release.Report,
	evidence []release.PlatformEvidence, stdout io.Writer) error {
	report, err := release.ReadReport(cfg.check)
	if err != nil {
		return err
	}
	if err := release.Validate(report, cfg.expectedHead); err != nil {
		return err
	}
	if report.ReportDigest != expected.ReportDigest {
		return fmt.Errorf("TOOLCHAIN_RELEASE_REPLAY_DRIFT")
	}
	if err := release.ValidateBundle(cfg.bundle, evidence); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, "toolchain-release: replay PASS/EXACT")
	return err
}
