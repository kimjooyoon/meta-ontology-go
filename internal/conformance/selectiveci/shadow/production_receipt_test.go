package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func (output productionOutput) selfDigest() string {
	copy := output
	copy.CanonicalDigest = ""
	copy.ChangedSemanticIDs = sortedUniqueProduction(copy.ChangedSemanticIDs)
	copy.SelectedCommands = sortedCommandsProduction(copy.SelectedCommands)
	copy.SelectedGuards = sortedCommandsProduction(copy.SelectedGuards)
	copy.SelectedWorkIDs = sortedUniqueProduction(copy.SelectedWorkIDs)
	copy.ResourceReceipts = sortedReceiptsProduction(copy.ResourceReceipts)
	data, err := json.Marshal(copy)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sortedUniqueProduction(values []string) []string {
	copy := append([]string(nil), values...)
	sort.Strings(copy)
	result := copy[:0]
	for _, value := range copy {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	if result == nil {
		return []string{}
	}
	return result
}

func sortedCommandsProduction(values []productionCommand) []productionCommand {
	copy := append([]productionCommand(nil), values...)
	sort.Slice(copy, func(left, right int) bool { return copy[left].ID < copy[right].ID })
	for index := range copy {
		copy[index].Argv = append([]string(nil), copy[index].Argv...)
	}
	if copy == nil {
		return []productionCommand{}
	}
	return copy
}

func sortedReceiptsProduction(values []productionResourceReceipt) []productionResourceReceipt {
	copy := append([]productionResourceReceipt(nil), values...)
	sort.Slice(copy, func(left, right int) bool { return copy[left].CommandID < copy[right].CommandID })
	if copy == nil {
		return []productionResourceReceipt{}
	}
	return copy
}
