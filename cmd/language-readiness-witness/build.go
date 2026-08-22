package main

import (
	"fmt"
	"io"
	"os"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func build(cfg config) (readinessartifact.Receipt, error) {
	raw, err := os.ReadFile(cfg.input)
	if err != nil {
		return readinessartifact.Receipt{}, err
	}
	return readinessartifact.Build(raw, cfg.expectedSHA)
}

func printSummary(stdout io.Writer, receipt readinessartifact.Receipt) {
	summary := receipt.Snapshot.Summary
	fmt.Fprintf(stdout,
		"language-readiness-artifact: decision=%s completed=%d total=%d bps=%d fixed_point=%s writes=%d digest=%s\n",
		receipt.Snapshot.Decision, summary.Completed, summary.Total, summary.ReadinessBPS,
		receipt.FixedPoint.Decision, receipt.Snapshot.RepositoryWrites, receipt.ArtifactDigest)
}
