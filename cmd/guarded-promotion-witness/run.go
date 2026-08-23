package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

func run(ctx context.Context, config config) error {
	if config.repository == "" || config.currentHeadSHA == "" || config.sourceRunID == 0 {
		return fmt.Errorf("repository, current-head-sha, and source-run-id are required")
	}
	if config.token == "" || config.expectDecision == "" {
		return fmt.Errorf("token and expect-decision are required")
	}
	collector := guardedpromotion.NewCollector(config.apiURL, config.token)
	source := collector.Collect(ctx, config.repository, config.currentHeadSHA, config.sourceRunID)
	report := guardedpromotion.Build(source)
	if err := guardedpromotion.Validate(report); err != nil {
		return err
	}
	if err := writeReport(config.out, report); err != nil {
		return err
	}
	if report.Decision != config.expectDecision {
		return fmt.Errorf("guarded promotion decision = %s, want %s", report.Decision, config.expectDecision)
	}
	fmt.Printf("decision=%s satisfied=%d total=%d bps=%d writes=%d feeds-next-promotion=%t\n",
		report.Decision, report.Summary.Satisfied, report.Summary.Total,
		report.Summary.ReadinessBPS, report.Summary.RepositoryWrites, feedsNextPromotion(report))
	return nil
}

func feedsNextPromotion(report guardedpromotion.Report) bool {
	const fixedCoordinateDenominator = 8
	return report.Summary.Satisfied == fixedCoordinateDenominator &&
		report.Summary.Total == fixedCoordinateDenominator &&
		report.Summary.ReadinessBPS == 10_000 &&
		report.Summary.RepositoryWrites == 0
}

func writeReport(path string, report guardedpromotion.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:]) + "\n"
	return os.WriteFile(path+".sha256", []byte(digest), 0o644)
}
