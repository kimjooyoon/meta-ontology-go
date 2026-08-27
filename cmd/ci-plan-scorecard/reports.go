package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/ciplanusecase"
	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

func readReports(directory string, contract ciplanusecase.Contract) (map[string]metainvocation.Report, error) {
	reports := make(map[string]metainvocation.Report, len(contract.Cases))
	for _, spec := range contract.Cases {
		raw, err := os.ReadFile(filepath.Join(directory, spec.ID+".json"))
		if err != nil {
			return nil, err
		}
		report := metainvocation.Report{}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&report); err != nil {
			return nil, fmt.Errorf("decode %s report: %w", spec.ID, err)
		}
		if err := metainvocation.Validate(report); err != nil {
			return nil, fmt.Errorf("validate %s report: %w", spec.ID, err)
		}
		reports[spec.ID] = report
	}
	return reports, nil
}

func readGoldens(directory string, contract ciplanusecase.Contract) (map[string]ciplanusecase.GoldenPlan, error) {
	goldens := map[string]ciplanusecase.GoldenPlan{}
	for _, spec := range contract.Cases {
		if spec.ExpectedDecision != metainvocation.DecisionPass {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(directory, spec.ID+".json"))
		if err != nil {
			return nil, err
		}
		golden, err := ciplanusecase.DecodeGolden(raw)
		if err != nil {
			return nil, err
		}
		goldens[spec.ID] = golden
	}
	return goldens, nil
}
