package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const consumeRepairUsage = "usage: gooo consume-repair <repair-candidate.json> --out <directory>"

func runConsumeRepair(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] != "--out" || args[0] == "" || args[2] == "" {
		fmt.Fprintln(stderr, consumeRepairUsage)
		return exitUsage
	}
	data, err := reader.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var candidate valueexecution.RepairCandidate
	if err := json.Unmarshal(data, &candidate); err != nil {
		fmt.Fprintf(stderr, "gooo: consume-repair: decode candidate: %v\n", err)
		return exitFailure
	}
	handoff, err := valueexecution.ConsumeRepairCandidate(candidate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: consume-repair: %v\n", err)
		return exitFailure
	}
	payload, err := json.MarshalIndent(handoff, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: consume-repair: encode handoff: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeRepairHandoff(args[2], payload); err != nil {
		fmt.Fprintf(stderr, "gooo: consume-repair: output: %v\n", err)
		return exitFailure
	}
	fname := filepath.Join(args[2], "repair-handoff.json")
	fmt.Fprintf(stdout, "repair handoff: %s\n", fname)
	return exitOK
}

func writeRepairHandoff(outputDir string, payload []byte) error {
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
	return os.WriteFile(filepath.Join(outputDir, "repair-handoff.json"), payload, 0o644)
}
