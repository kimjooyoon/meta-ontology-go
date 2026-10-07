package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
)

type CompositionCaseSeries struct {
	Schema string             `json:"schema"`
	Suites []CompositionCases `json:"suites"`
}

func DecodeCompositionCaseSeries(raw []byte) (CompositionCaseSeries, error) {
	var series CompositionCaseSeries
	if err := decode(raw, &series, 512<<10); err != nil {
		return series, err
	}
	if series.Schema != "gooo/body-composition-case-series/v1" || len(series.Suites) < 1 || len(series.Suites) > 16 {
		return series, fmt.Errorf("composition case series requires schema and 1..16 suites")
	}
	for i, suite := range series.Suites {
		encoded, err := json.Marshal(suite)
		if err != nil {
			return series, err
		}
		if _, err := DecodeCompositionCases(encoded); err != nil {
			return series, fmt.Errorf("suite %d: %w", i, err)
		}
	}
	return series, nil
}

// ValidateCompositionSuites checks the ordered series against one source graph
// before generation. Every suite's root values and named expectations are typed.
func ValidateCompositionSuites(ctx context.Context, filename string, source []byte, suites []CompositionCases) error {
	return ValidateCompositionSuitesForEntry(ctx, filename, source, suites, "")
}

func ValidateCompositionSuitesForEntry(ctx context.Context, filename string, source []byte, suites []CompositionCases, entry string) error {
	if len(suites) < 1 || len(suites) > 16 {
		return fmt.Errorf("composition requires 1..16 suites")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, entry)
	if err != nil {
		return err
	}
	for i, suite := range suites {
		if _, err := graph.inputRows(suite); err != nil {
			return fmt.Errorf("suite %d: %w", i, err)
		}
	}
	return nil
}
