package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/predecessorbinding"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.expectedSHA == "" || cfg.output == "" {
		return fmt.Errorf("root, expected-sha, and output are required")
	}
	outside, err := outsideRoot(cfg.root, cfg.output)
	if err != nil {
		return err
	}
	if !outside {
		return fmt.Errorf("predecessor binding receipt must be outside the repository root")
	}
	observations, err := scan(cfg.root)
	if err != nil {
		return err
	}
	report := predecessorbinding.Evaluate(cfg.expectedSHA, observations, 0)
	replay := predecessorbinding.Evaluate(cfg.expectedSHA, observations, 0)
	if report.ReportDigest != replay.ReportDigest {
		return fmt.Errorf("predecessor binding replay mismatch")
	}
	if err := predecessorbinding.Validate(report, cfg.expectedSHA); err != nil {
		return err
	}
	if cfg.check && report.Decision != predecessorbinding.DecisionPass {
		return fmt.Errorf("%s: %s", report.Decision, report.Reason)
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if err := writeExclusive(cfg.output, payload); err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	checksum := []byte(hex.EncodeToString(sum[:]) + "  " + filepath.Base(cfg.output) + "\n")
	if err := writeExclusive(cfg.output+".sha256", checksum); err != nil {
		return err
	}
	fmt.Printf("predecessor-binding: static=%d dynamic=%d unknown=%d total=%d bps=%d writes=%d\n",
		report.Summary.StaticLiteral, report.Summary.DynamicInput, report.Summary.Unknown,
		report.Summary.Total, report.Summary.DynamicBPS, report.RepositoryWrites)
	return nil
}

func outsideRoot(root, output string) (bool, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return false, err
	}
	return relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func writeExclusive(path string, payload []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
