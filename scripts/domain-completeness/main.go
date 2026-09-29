package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	profilePath := flag.String("profile", "scripts/domain-completeness/profile.gooo", "Gooo domain profile")
	contractPath := flag.String("contract", "examples/language-utility/contract.json", "language utility contract")
	evidenceDir := flag.String("evidence", "", "exact language utility evidence directory")
	subject := flag.String("subject", "", "exact source commit SHA")
	runID := flag.Int64("run-id", 0, "exact workflow run ID")
	attempt := flag.Int("run-attempt", 0, "exact workflow run attempt")
	outputPath := flag.String("output", "", "JSON receipt output")
	programPath := flag.String("program", "", "generated Gooo receipt program output")
	check := flag.Bool("check", false, "require outputs to replay exactly")
	flag.Parse()

	if *evidenceDir == "" || *outputPath == "" || *programPath == "" {
		exitError(fmt.Errorf("-evidence, -output, and -program are required"))
	}
	report, program, err := evaluate(*profilePath, *contractPath, *evidenceDir, *subject, *runID, *attempt)
	if err != nil {
		exitError(err)
	}
	reportRaw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		exitError(err)
	}
	reportRaw = append(reportRaw, '\n')
	if err := output(*outputPath, reportRaw, *check); err != nil {
		exitError(err)
	}
	if err := output(*programPath, program, *check); err != nil {
		exitError(err)
	}
	fmt.Printf("domain completeness: dimensions=%d pass=%d progress=%d unknown=%d fail_closed=%d decision=%s\n",
		report.Summary.DimensionsTotal, report.Summary.Pass, report.Summary.Progress,
		report.Summary.Unknown, report.Summary.FailClosed, report.Decision)
}

func output(path string, expected []byte, check bool) error {
	if check {
		actual, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, expected) {
			return fmt.Errorf("domain completeness replay differs: %s", path)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, expected, 0o644)
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, "domain-completeness:", err)
	os.Exit(2)
}

func canonicalSubject(value string) bool {
	if len(value) != 40 || strings.ToLower(value) != value {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
