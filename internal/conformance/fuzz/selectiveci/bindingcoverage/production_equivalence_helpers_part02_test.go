package bindingcoverage

import (
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"strings"
)

func translateInput(input Input) (production.Input, error) {
	result := production.Input{
		SchemaVersion:          translateSchema(input.Schema),
		ContractID:             fixedProductionContractID,
		SnapshotDigest:         translateDigest(input.SnapshotDigest),
		ExpectedSnapshotDigest: translateDigest(input.ExpectedDigest),
	}
	if input.Precedence != nil {
		result.PrecedenceRegistry = make([]production.PrecedenceEntry, 0, len(input.Precedence))
	}
	if input.RequiredBindings != nil {
		result.RequiredBindings = make([]production.RequiredBinding, 0, len(input.RequiredBindings))
	}
	if input.Partitions != nil {
		result.Partitions = make([]production.Partition, 0, len(input.Partitions))
	}
	for _, entry := range input.Precedence {
		if entry.Rank < 0 {
			return production.Input{}, fmt.Errorf("negative precedence rank")
		}
		result.PrecedenceRegistry = append(result.PrecedenceRegistry, production.PrecedenceEntry{
			Rank: uint64(entry.Rank), Stage: entry.Stage, Reason: entry.Reason,
		})
	}
	for _, binding := range input.RequiredBindings {
		result.RequiredBindings = append(result.RequiredBindings, production.RequiredBinding{
			BindingID: translateID(binding.ID), FromFieldID: translateID(binding.FromFieldID),
			ToFieldID: translateID(binding.ToFieldID), Kind: production.BindingKind(binding.Kind),
			ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason,
		})
	}
	partitionOrdinals := make(map[string]int)
	for _, partition := range input.Partitions {
		key := partition.BindingID + "\x00" + partition.Polarity
		ordinal := partitionOrdinals[key]
		partitionOrdinals[key] = ordinal + 1
		result.Partitions = append(result.Partitions, production.Partition{
			PartitionID: generatedPartitionID(partition, ordinal),
			BindingID:   translateID(partition.BindingID), Polarity: production.Polarity(partition.Polarity),
			ExpectedStage: partition.Stage, ExpectedReason: partition.Reason,
		})
	}
	return result, nil
}
func translateSchema(schema string) string {
	if schema == SchemaV1 {
		return production.SchemaVersion
	}
	if strings.HasPrefix(schema, "binding-coverage/") {
		return "gooo/selective-ci-binding-coverage/" + strings.TrimPrefix(schema, "binding-coverage/")
	}
	return schema
}
func translateDigest(digest string) string {
	if strings.HasPrefix(digest, "sha256:") {
		return strings.TrimPrefix(digest, "sha256:")
	}
	return digest
}
func translateID(id string) string {
	if strings.HasPrefix(id, "sid:") {
		return "urn:bindingcoverage:" + strings.TrimPrefix(id, "sid:")
	}
	return id
}
