package main

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Origin columns transport recorded claims. The Gooo consumer owns the metric
// rules, grouping and interpretation; this adapter performs no coverage math.
type localFillColumns struct {
	SelectedSourceSHA string   `json:"selected_source_sha"`
	RecordedWrites    int      `json:"recorded_writes"`
	PageComplete      bool     `json:"page_complete"`
	RecordingComplete bool     `json:"recording_complete"`
	CaseIndexes       []int    `json:"case_indexes"`
	Assignments       []int    `json:"assignments"`
	Kinds             []string `json:"kinds"`
	SourceSHAs        []string `json:"source_shas"`
	Starts            []int    `json:"starts"`
	Ends              []int    `json:"ends"`
}

func shapePackageOriginRows(rows []json.RawMessage, origin *bodyexecution.NativeOriginInput, shape string) ([]json.RawMessage, error) {
	if shape == "" || shape == "rows" {
		return rows, nil
	}
	if shape != "local-fill-columns" || origin.Query.Kind != "local-fill-fragments" || origin.LocalObservation == nil {
		return nil, fmt.Errorf("local-fill-columns requires a scalar fragment observation")
	}
	n := len(rows)
	columns := localFillColumns{SelectedSourceSHA: origin.SelectedSourceSHA256,
		RecordedWrites:    origin.LocalObservation.RecordedWrites,
		PageComplete:      origin.Query.Offset == 0 && !origin.Truncated,
		RecordingComplete: origin.LocalObservation.Status == "COMPLETE",
		CaseIndexes:       make([]int, n), Assignments: make([]int, n), Kinds: make([]string, n),
		SourceSHAs: make([]string, n), Starts: make([]int, n), Ends: make([]int, n)}
	for index, row := range rows {
		var fact bodyexecution.LocalFillFragmentFact
		if err := json.Unmarshal(row, &fact); err != nil {
			return nil, err
		}
		columns.CaseIndexes[index], columns.Assignments[index] = fact.CaseIndex, fact.AssignmentIndex
		columns.Kinds[index], columns.SourceSHAs[index] = fact.Category, fact.SourceSHA
		columns.Starts[index], columns.Ends[index] = fact.SourceStart, fact.SourceEnd
	}
	encoded, err := json.Marshal(columns)
	if err != nil {
		return nil, err
	}
	origin.InputShape, origin.ConsumerRows = "gooo/local-fill-columns/v1", 1
	return []json.RawMessage{encoded}, nil
}
