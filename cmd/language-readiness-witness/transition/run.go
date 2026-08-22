package main

import (
	"fmt"
	"io"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func run(args []string, output io.Writer) error {
	value, err := parseConfig(args)
	if err != nil {
		return err
	}
	if err := validateExternalPaths(value); err != nil {
		return err
	}
	beforeRaw, before, err := readReadiness(value.before)
	if err != nil {
		return fmt.Errorf("read baseline: %w", err)
	}
	_, after, err := readReadiness(value.after)
	if err != nil {
		return fmt.Errorf("read current: %w", err)
	}
	receipt, err := readinessartifact.BuildImprovement(
		beforeRaw, before, after, value.expectedSHA,
	)
	if err != nil {
		return err
	}
	raw, err := encode(receipt)
	if err != nil {
		return err
	}
	if value.output != "" {
		err = writeArtifact(value.output, raw)
	} else {
		err = compareArtifact(value.check, raw)
	}
	if err != nil {
		return err
	}
	transition := receipt.Transition
	fmt.Fprintf(output, "language-readiness-improvement: decision=%s "+
		"before=%d after=%d total=%d delta=%d bps_delta=%d gains=%d "+
		"regressions=%d unresolved=%d writes=%d digest=%s\n",
		transition.Decision, transition.BeforeCompleted,
		transition.AfterCompleted, transition.Total,
		transition.CompletedDelta, transition.BasisPointsDelta,
		transition.Gains, transition.Regressions,
		transition.AfterUnresolved, receipt.RepositoryWrites,
		receipt.ArtifactDigest)
	return nil
}
