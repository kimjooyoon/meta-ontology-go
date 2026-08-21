package bindingcoverage

import (
	"encoding/json"
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"sort"
)

func marshalProductionInput(input production.Input) ([]byte, error) {
	wire := productionInputWire{
		SchemaVersion: input.SchemaVersion, ContractID: input.ContractID,
		SnapshotDigest: input.SnapshotDigest, ExpectedSnapshotDigest: input.ExpectedSnapshotDigest,
	}
	if input.RequiredBindings != nil {
		wire.RequiredBindings = make([]productionBindingWire, 0, len(input.RequiredBindings))
	}
	if input.Partitions != nil {
		wire.Partitions = make([]productionPartitionWire, 0, len(input.Partitions))
	}
	if input.PrecedenceRegistry != nil {
		wire.PrecedenceRegistry = make([]productionPrecedenceWire, 0, len(input.PrecedenceRegistry))
	}
	for _, binding := range input.RequiredBindings {
		wire.RequiredBindings = append(wire.RequiredBindings, productionBindingWire{
			BindingID: binding.BindingID, FromFieldID: binding.FromFieldID, ToFieldID: binding.ToFieldID,
			Kind: string(binding.Kind), ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason,
		})
	}
	for _, partition := range input.Partitions {
		wire.Partitions = append(wire.Partitions, productionPartitionWire{
			PartitionID: partition.PartitionID, BindingID: partition.BindingID, Polarity: string(partition.Polarity),
			ExpectedStage: partition.ExpectedStage, ExpectedReason: partition.ExpectedReason,
		})
	}
	for _, entry := range input.PrecedenceRegistry {
		wire.PrecedenceRegistry = append(wire.PrecedenceRegistry, productionPrecedenceWire{
			Rank: entry.Rank, Stage: entry.Stage, Reason: entry.Reason,
		})
	}
	sort.SliceStable(wire.RequiredBindings, func(i, j int) bool {
		return productionBindingKey(wire.RequiredBindings[i]) < productionBindingKey(wire.RequiredBindings[j])
	})
	sort.SliceStable(wire.Partitions, func(i, j int) bool {
		return productionPartitionKey(wire.Partitions[i]) < productionPartitionKey(wire.Partitions[j])
	})
	sort.SliceStable(wire.PrecedenceRegistry, func(i, j int) bool {
		return productionPrecedenceKey(wire.PrecedenceRegistry[i]) < productionPrecedenceKey(wire.PrecedenceRegistry[j])
	})
	return json.Marshal(wire)
}
func productionBindingKey(binding productionBindingWire) string {
	return binding.BindingID + "\x00" + binding.FromFieldID + "\x00" + binding.ToFieldID + "\x00" + binding.Kind + "\x00" + binding.ExpectedStage + "\x00" + binding.ExpectedReason
}
func productionPartitionKey(partition productionPartitionWire) string {
	return partition.PartitionID + "\x00" + partition.BindingID + "\x00" + partition.Polarity + "\x00" + partition.ExpectedStage + "\x00" + partition.ExpectedReason
}
func productionPrecedenceKey(entry productionPrecedenceWire) string {
	return fmt.Sprintf("%020d\x00%s\x00%s", entry.Rank, entry.Stage, entry.Reason)
}
