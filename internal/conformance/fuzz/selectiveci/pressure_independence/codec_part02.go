package pressureindependence

import (
	"encoding/json"
	"sort"
)

func EncodeOutputJSON(output Output) ([]byte, error) {
	output.SelectedIDs = sortedUnique(output.SelectedIDs)
	output.UnselectedIDs = sortedUnique(output.UnselectedIDs)
	output.UnknownIDs = sortedUnique(output.UnknownIDs)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

type outputDigestView struct {
	Schema             string      `json:"schema"`
	FixtureID          string      `json:"fixture_id"`
	InputDigest        string      `json:"input_digest"`
	SelectedIDs        []string    `json:"selected_ids"`
	UnselectedIDs      []string    `json:"unselected_ids"`
	UnknownIDs         []string    `json:"unknown_ids"`
	DistinctGroupCount uint64      `json:"distinct_group_count"`
	Decision           Decision    `json:"decision"`
	Reason             Reason      `json:"reason"`
	FullSuiteRequired  bool        `json:"full_suite_required"`
	ProofValid         bool        `json:"proof_valid"`
	CostReceipt        CostReceipt `json:"cost_receipt"`
}

func normalizeInput(input Input) Input {
	input.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	input.RequiredPressureIDs = append([]string(nil), input.RequiredPressureIDs...)
	input.GuardIDs = append([]string(nil), input.GuardIDs...)
	input.FinitePathIDs = append([]string(nil), input.FinitePathIDs...)
	sort.Slice(input.PressureRecords, func(i, j int) bool {
		left, right := input.PressureRecords[i], input.PressureRecords[j]
		return pressureKey(left) < pressureKey(right)
	})
	sort.Strings(input.RequiredPressureIDs)
	sort.Strings(input.GuardIDs)
	sort.Strings(input.FinitePathIDs)
	return input
}
func pressureKey(record PressureRecord) string {
	return record.PressureID + "\x00" + record.CategoryID + "\x00" +
		record.IndependenceGroupID + "\x00" + record.ApplicabilityRuleID
}
func sortedUnique(values []string) []string {
	if values == nil {
		return []string{}
	}
	values = append([]string(nil), values...)
	if len(values) == 0 {
		return []string{}
	}
	sort.Strings(values)
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
