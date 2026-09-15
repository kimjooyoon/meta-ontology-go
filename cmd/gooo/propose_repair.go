package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const proposeRepairUsage = "usage: gooo propose-repair <replay-comparison.json> --out <directory>"

func runProposeRepair(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] != "--out" || args[0] == "" || args[2] == "" {
		fmt.Fprintln(stderr, proposeRepairUsage)
		return exitUsage
	}
	data, err := reader.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var comparison valueexecution.ReplayComparison
	if err := json.Unmarshal(data, &comparison); err != nil {
		fmt.Fprintf(stderr, "gooo: repair candidate: decode comparison: %v\n", err)
		return exitFailure
	}
	repair, err := valueexecution.ProposeRepair(comparison)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: repair candidate: %v\n", err)
		return exitFailure
	}
	payload, err := json.MarshalIndent(repair, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: repair candidate: encode: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeRepairCandidate(args[2], payload); err != nil {
		fmt.Fprintf(stderr, "gooo: repair candidate: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "repair candidate: %s\n", filepath.Join(args[2], "repair-candidate.json"))
	return exitOK
}

func writeRepairCandidate(outputDir string, payload []byte) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	return os.WriteFile(filepath.Join(outputDir, "repair-candidate.json"), payload, 0o644)
}
