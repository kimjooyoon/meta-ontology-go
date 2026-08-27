package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

func main() {
	sourcePath := flag.String("source", "examples/meta-circular-boundary/main.gooo", "Gooo source to re-read")
	reportPath := flag.String("report", "", "producer report to judge")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	grantPath := flag.String("grant", "", "raw external grant artifact path")
	effectPath := flag.String("effect-evidence", "", "raw workspace effect artifact path")
	replayPath := flag.String("replay-evidence", "", "raw replay evidence artifact path")
	executionDir := flag.String("execution-dir", "", "directory containing execution artifacts")
	judgeOutput := flag.String("judge-output", "", "independent judge receipt output path")
	flag.Parse()
	if *reportPath == "" {
		fatal(fmt.Errorf("--report is required"))
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	reportBytes, err := os.ReadFile(*reportPath)
	if err != nil {
		fatal(err)
	}
	var report contract.Report
	decoder := json.NewDecoder(bytes.NewReader(reportBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		fatal(err)
	}
	grant, err := readOptional(*grantPath)
	if err != nil {
		fatal(err)
	}
	effect, err := readOptional(*effectPath)
	if err != nil {
		fatal(err)
	}
	replay, err := readOptional(*replayPath)
	if err != nil {
		fatal(err)
	}
	artifacts, err := loadExecutionArtifacts(report, *executionDir)
	if err != nil {
		fatal(err)
	}
	input := contract.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source, GrantEvidence: grant, EffectEvidence: effect, ReplayEvidence: replay, ExecutionArtifacts: artifacts}
	judge, err := metacircularboundaryconsumer.JudgeWithReceipt(report, input)
	if err != nil {
		fatal(err)
	}
	if *judgeOutput != "" {
		encoded, err := json.MarshalIndent(judge, "", "  ")
		if err != nil {
			fatal(err)
		}
		encoded = append(encoded, '\n')
		if err := os.WriteFile(*judgeOutput, encoded, 0o644); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("consumer judge: PASS cases=%d indicators=%d\n", report.Summary.CasesPassed, len(report.Indicators))
}

func readOptional(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

func loadExecutionArtifacts(report contract.Report, root string) ([]contract.ExecutionArtifact, error) {
	if len(report.ExecutionArtifacts) == 0 {
		return nil, nil
	}
	if root == "" {
		return nil, fmt.Errorf("execution artifact directory is required")
	}
	artifacts := make([]contract.ExecutionArtifact, 0, len(report.ExecutionArtifacts))
	for _, reference := range report.ExecutionArtifacts {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(reference.Path)))
		if err != nil {
			return nil, err
		}
		var artifact contract.ExecutionArtifact
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&artifact); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
