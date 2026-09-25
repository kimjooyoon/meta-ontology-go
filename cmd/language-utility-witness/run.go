package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
)

func run(args []string) error {
	config, err := parseConfig(args)
	if err != nil {
		return err
	}
	contractRaw, err := os.ReadFile(config.contract)
	if err != nil {
		return err
	}
	observationRaw, err := os.ReadFile(config.observation)
	if err != nil {
		return err
	}
	contract, err := languageutility.DecodeContract(contractRaw)
	if err != nil {
		return err
	}
	observation, err := languageutility.DecodeObservation(observationRaw)
	if err != nil {
		return err
	}
	report, err := languageutility.Evaluate(contract, observation)
	if err != nil {
		return err
	}
	program, err := languageutility.GenerateProgram(contract)
	if err != nil {
		return err
	}
	reportRaw, err := languageutility.MarshalReport(report)
	if err != nil {
		return err
	}
	if err := output(config.report, reportRaw, config.check); err != nil {
		return err
	}
	if err := output(config.program, []byte(program), config.check); err != nil {
		return err
	}
	fmt.Printf("language utility: cells=%d/%d use-cases=%d/%d remaining=%d decision=%s\n",
		report.Summary.ClosedCells, report.Summary.CellsTotal, report.Summary.CompleteUseCases,
		report.Summary.UseCasesTotal, report.Summary.RemainingCells, report.Decision)
	if report.Decision == "FAIL_CLOSED" {
		return fmt.Errorf("%s", report.Reason)
	}
	return nil
}

func output(path string, data []byte, check bool) error {
	if check {
		existing, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(existing, data) {
			return fmt.Errorf("replay differs: %s", path)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
