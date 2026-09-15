package workfrontier

import (
	"encoding/json"
	"fmt"
	"sort"
)

// EncodeR4JSON emits one normalized, versioned input envelope.
func EncodeR4JSON(input R4Input) ([]byte, error) {
	input = normalizeR4Input(input)
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode r4 work frontier: %w", err)
	}
	return append(data, '\n'), nil
}

// EncodeR4ResultJSON emits the deterministic result without adding any
// authorization or proof fields.
func EncodeR4ResultJSON(result R4Result) ([]byte, error) {
	result = normalizeR4Result(result)
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode r4 work frontier result: %w", err)
	}
	return append(data, '\n'), nil
}
func normalizeR4Input(input R4Input) R4Input {
	input.Pressures = append([]Pressure(nil), input.Pressures...)
	input.States = append([]ObligationState(nil), input.States...)
	input.Paths = append([]RepairPath(nil), input.Paths...)
	input.RootObligationIDs = sortedCopy(input.RootObligationIDs)
	input.Rules = normalizeR4Rules(input.Rules)
	for index := range input.Paths {
		input.Paths[index].PrerequisiteObligationIDs = sortedCopy(input.Paths[index].PrerequisiteObligationIDs)
		input.Paths[index].ReadSet = sortedCopy(input.Paths[index].ReadSet)
		input.Paths[index].WriteSet = sortedCopy(input.Paths[index].WriteSet)
		input.Paths[index].RequiredPressureIDs = sortedCopy(input.Paths[index].RequiredPressureIDs)
	}
	sort.Slice(input.Pressures, func(i, j int) bool { return input.Pressures[i].StableID < input.Pressures[j].StableID })
	sort.Slice(input.States, func(i, j int) bool {
		if input.States[i].ObligationID != input.States[j].ObligationID {
			return input.States[i].ObligationID < input.States[j].ObligationID
		}
		return input.States[i].Status < input.States[j].Status
	})
	sort.Slice(input.Paths, func(i, j int) bool {
		return r4PathKey(input.Paths[i]) < r4PathKey(input.Paths[j])
	})
	return input
}
func normalizeR4Result(result R4Result) R4Result {
	result.Selected = uniqueInOrder(result.Selected)
	result.SelectedIDs = uniqueInOrder(result.SelectedIDs)
	result.WorkIDs = uniqueInOrder(result.WorkIDs)
	result.Unknown = sortedUnique(result.Unknown)
	result.Blocked = sortedUnique(result.Blocked)
	result.Shortfall = sortedUnique(result.Shortfall)
	return result
}
func r4PathKey(path RepairPath) string {
	return path.StableID + "\x00" + path.ObligationID + "\x00" +
		joinR4(path.PrerequisiteObligationIDs) + "\x00" + joinR4(path.ReadSet) + "\x00" +
		joinR4(path.WriteSet) + "\x00" + joinR4(path.RequiredPressureIDs) + "\x00" +
		path.WorkID + "\x00" + fmt.Sprint(path.PolicyPriority) + "\x00" + fmt.Sprint(path.CPUCoreNSUpperBound)
}
