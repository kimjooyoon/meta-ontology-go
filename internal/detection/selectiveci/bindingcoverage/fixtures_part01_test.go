package bindingcoverage

import (
	"strings"
)

func fixtureInput() Input {
	names := []string{"base-head", "lane-registry", "base-manifest", "head-manifest", "plan-proof-digest", "changed-roots", "selected-command-union", "base-snapshot", "head-snapshot"}
	kinds := []BindingKind{KindExactValue, KindExactDigest, KindExactDigest, KindExactDigest, KindDerivedDigest, KindSetEqual, KindSetEqual, KindExactValue, KindExactValue}
	bindings := make([]RequiredBinding, 0, len(names))
	partitions := make([]Partition, 0, len(names)*2)
	for index, name := range names {
		binding := RequiredBinding{BindingID: bindingID(name), FromFieldID: id("field/" + name + "/from"), ToFieldID: id("field/" + name + "/to"), Kind: kinds[index],
			ExpectedStage:  "stage:" + name,
			ExpectedReason: "reason:binding-check"}
		bindings = append(bindings, binding)
		partitions = append(partitions, Partition{PartitionID: id("partition/" + name + "/match"), BindingID: binding.BindingID, Polarity: PolarityMatch, ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason})
		partitions = append(partitions, Partition{PartitionID: id("partition/" + name + "/mismatch"), BindingID: binding.BindingID, Polarity: PolarityMismatch, ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason})
	}
	precedence := make([]PrecedenceEntry, 0, len(names))
	for index, name := range names {
		precedence = append(precedence, PrecedenceEntry{Rank: uint64(index + 1), Stage: "stage:" + name, Reason: "reason:binding-check"})
	}
	snapshot := strings.Repeat("a", 64)
	return Input{SchemaVersion: SchemaVersion, ContractID: id("contract/selective-ci"), SnapshotDigest: snapshot, ExpectedSnapshotDigest: snapshot, RequiredBindings: bindings, Partitions: partitions, PrecedenceRegistry: precedence}
}
func sharedEndpointInput() Input {
	middle := id("field/shared/middle")
	bindings := []RequiredBinding{
		{BindingID: bindingID("shared-a"), FromFieldID: id("field/shared/from"), ToFieldID: middle, Kind: KindExactValue, ExpectedStage: "stage:shared-a", ExpectedReason: "reason:shared-check"},
		{BindingID: bindingID("shared-b"), FromFieldID: middle, ToFieldID: id("field/shared/to"), Kind: KindExactDigest, ExpectedStage: "stage:shared-b", ExpectedReason: "reason:shared-check"},
	}
	partitions := []Partition{
		{PartitionID: id("partition/shared-a/match"), BindingID: bindingID("shared-a"), Polarity: PolarityMatch, ExpectedStage: "stage:shared-a", ExpectedReason: "reason:shared-check"},
		{PartitionID: id("partition/shared-a/mismatch"), BindingID: bindingID("shared-a"), Polarity: PolarityMismatch, ExpectedStage: "stage:shared-a", ExpectedReason: "reason:shared-check"},
		{PartitionID: id("partition/shared-b/match"), BindingID: bindingID("shared-b"), Polarity: PolarityMatch, ExpectedStage: "stage:shared-b", ExpectedReason: "reason:shared-check"},
		{PartitionID: id("partition/shared-b/mismatch"), BindingID: bindingID("shared-b"), Polarity: PolarityMismatch, ExpectedStage: "stage:shared-b", ExpectedReason: "reason:shared-check"},
	}
	precedence := []PrecedenceEntry{{Rank: 1, Stage: "stage:shared-a", Reason: "reason:shared-check"}, {Rank: 2, Stage: "stage:shared-b", Reason: "reason:shared-check"}}
	snapshot := strings.Repeat("b", 64)
	return Input{SchemaVersion: SchemaVersion, ContractID: id("contract/shared-endpoint"), SnapshotDigest: snapshot, ExpectedSnapshotDigest: snapshot, RequiredBindings: bindings, Partitions: partitions, PrecedenceRegistry: precedence}
}
func bindingID(name string) string { return id("binding/" + name) }
func id(value string) string       { return "urn:bindingcoverage:" + value }
func withoutPartition(partitions []Partition, bindingID string, polarity Polarity) []Partition {
	result := make([]Partition, 0, len(partitions)-1)
	for _, partition := range partitions {
		if partition.BindingID == bindingID && partition.Polarity == polarity {
			continue
		}
		result = append(result, partition)
	}
	return result
}
func reverseBindings(values []RequiredBinding) []RequiredBinding {
	result := append([]RequiredBinding{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
func reversePartitions(values []Partition) []Partition {
	result := append([]Partition{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
