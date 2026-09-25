package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
)

func run(cfg config) (bool, error) {
	if cfg.root == "" || cfg.actionability == "" || cfg.observations == "" || cfg.report == "" {
		return false, fmt.Errorf("root, actionability, observations, and report are required")
	}
	outside, err := outsideRoot(cfg.root, cfg.report)
	if err != nil {
		return false, err
	}
	if !outside {
		return false, fmt.Errorf("coverage report must be outside the repository root")
	}
	action, observations, err := artifactcoverage.Load(cfg.actionability, cfg.observations)
	if err != nil {
		return false, err
	}
	report, err := artifactcoverage.Evaluate(cfg.root, action, observations)
	if err != nil {
		return false, err
	}
	data, err := artifactcoverage.Marshal(report)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(cfg.report, data, 0o644); err != nil {
		return false, err
	}
	fmt.Printf("operation-artifact: decision=%s operations=%d/%d selected=%s digest=%s\n",
		report.Decision, report.Summary.CanonicalOperations,
		report.Summary.RequiredOperations, report.SelectedOperation, report.ReportDigest)
	return cfg.check && report.Decision == "FAIL_CLOSED", nil
}

func outsideRoot(root, output string) (bool, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false, err
	}
	outputPath, err := filepath.Abs(filepath.Clean(output))
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return false, err
	}
	return relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)), nil
}
