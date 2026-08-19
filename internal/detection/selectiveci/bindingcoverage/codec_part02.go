package bindingcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

func (output Output) Canonical() string {
	data, err := output.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}
func (output Output) StableDigest() string {
	return digestBytes([]byte(output.Canonical()))
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func normalizeInput(input Input) Input {
	input.RequiredBindings = copyBindings(input.RequiredBindings)
	input.Partitions = copyPartitions(input.Partitions)
	input.PrecedenceRegistry = copyPrecedence(input.PrecedenceRegistry)
	sort.SliceStable(input.RequiredBindings, func(i, j int) bool {
		return requiredBindingKey(input.RequiredBindings[i]) < requiredBindingKey(input.RequiredBindings[j])
	})
	sort.SliceStable(input.Partitions, func(i, j int) bool {
		return partitionKey(input.Partitions[i]) < partitionKey(input.Partitions[j])
	})
	sort.SliceStable(input.PrecedenceRegistry, func(i, j int) bool {
		return precedenceKey(input.PrecedenceRegistry[i]) < precedenceKey(input.PrecedenceRegistry[j])
	})
	return input
}
func copyBindings(values []RequiredBinding) []RequiredBinding {
	if values == nil {
		return nil
	}
	return append(make([]RequiredBinding, 0, len(values)), values...)
}
func copyPartitions(values []Partition) []Partition {
	if values == nil {
		return nil
	}
	return append(make([]Partition, 0, len(values)), values...)
}
func copyPrecedence(values []PrecedenceEntry) []PrecedenceEntry {
	if values == nil {
		return nil
	}
	return append(make([]PrecedenceEntry, 0, len(values)), values...)
}
func requiredBindingKey(binding RequiredBinding) string {
	return binding.BindingID + "\x00" + binding.FromFieldID + "\x00" + binding.ToFieldID + "\x00" + string(binding.Kind) + "\x00" + binding.ExpectedStage + "\x00" + binding.ExpectedReason
}
func partitionKey(partition Partition) string {
	return partition.PartitionID + "\x00" + partition.BindingID + "\x00" + string(partition.Polarity) + "\x00" + partition.ExpectedStage + "\x00" + partition.ExpectedReason
}
func precedenceKey(entry PrecedenceEntry) string {
	return fmt.Sprintf("%020d\x00%s\x00%s", entry.Rank, entry.Stage, entry.Reason)
}
func normalizeOutput(output Output) Output {
	output.MissingMatchBindingIDs = sortedStrings(output.MissingMatchBindingIDs)
	output.MissingMismatchBindingIDs = sortedStrings(output.MissingMismatchBindingIDs)
	return output
}
