package selectiveci

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func wireInputFrom(value Input) wireInput {
	paths := make([]wirePath, 0, len(value.Paths))
	for _, path := range value.Paths {
		paths = append(paths, wirePathFrom(path))
	}
	receipts := make([]wireCommandReceipt, 0, len(value.CommandReceipts))
	for _, receipt := range value.CommandReceipts {
		receipts = append(receipts, wireCommandReceiptFrom(receipt))
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].PathID < paths[j].PathID })
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	changedRoots := append([]semantic.ID(nil), value.ChangedRootIDs...)
	selectedCommands := append([]semantic.ID(nil), value.SelectedCommandIDs...)
	obligations := append([]semantic.ID(nil), value.ObligationIDs...)
	evidenceIDs := append([]semantic.ID(nil), value.EvidenceIDs...)
	sort.Slice(changedRoots, func(i, j int) bool { return changedRoots[i] < changedRoots[j] })
	sort.Slice(selectedCommands, func(i, j int) bool { return selectedCommands[i] < selectedCommands[j] })
	sort.Slice(obligations, func(i, j int) bool { return obligations[i] < obligations[j] })
	sort.Slice(evidenceIDs, func(i, j int) bool { return evidenceIDs[i] < evidenceIDs[j] })
	return wireInput{Schema: value.Schema, Snapshots: wireBindingFrom(value.Snapshots), RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, ChangedRootIDs: idsToStrings(changedRoots), SelectedCommandIDs: idsToStrings(selectedCommands), ObligationIDs: idsToStrings(obligations), Paths: paths, CommandReceipts: receipts, EvidenceIDs: idsToStrings(evidenceIDs), InferencePath: wirePathSetFrom(value.InferencePath)}
}
func inputFromWire(value wireInput) Input {
	paths := make([]Path, 0, len(value.Paths))
	for _, path := range value.Paths {
		paths = append(paths, pathFromWire(path))
	}
	receipts := make([]CommandReceipt, 0, len(value.CommandReceipts))
	for _, receipt := range value.CommandReceipts {
		receipts = append(receipts, commandReceiptFromWire(receipt))
	}
	return Input{Schema: value.Schema, Snapshots: bindingFromWire(value.Snapshots), RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, ChangedRootIDs: stringsToIDs(value.ChangedRootIDs), SelectedCommandIDs: stringsToIDs(value.SelectedCommandIDs), ObligationIDs: stringsToIDs(value.ObligationIDs), Paths: paths, CommandReceipts: receipts, EvidenceIDs: stringsToIDs(value.EvidenceIDs), InferencePath: pathSetFromWire(value.InferencePath)}
}
func wireReceiptFrom(value Receipt) wireReceipt {
	return wireReceipt{Schema: value.Schema, Status: string(value.Status), Fallback: string(value.Fallback), Code: value.Code, Snapshots: wireBindingFrom(value.Snapshots), RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, SelectedCommandIDs: idsToStrings(value.SelectedCommandIDs), ObligationIDs: idsToStrings(value.ObligationIDs), PathIDs: idsToStrings(value.PathIDs), RequiredCommandCount: value.RequiredCommandCount, RequiredObligationCount: value.RequiredObligationCount, VerifiedCommandCount: value.VerifiedCommandCount, VerifiedObligationCount: value.VerifiedObligationCount, VerifiedPathCount: value.VerifiedPathCount, VerifiedCommandIDs: idsToStrings(value.VerifiedCommandIDs), VerifiedObligationIDs: idsToStrings(value.VerifiedObligationIDs), VerifiedPathIDs: idsToStrings(value.VerifiedPathIDs), Digest: value.Digest}
}
func receiptFromWire(value wireReceipt) Receipt {
	return Receipt{Schema: value.Schema, Status: DecisionStatus(value.Status), Fallback: FallbackMode(value.Fallback), Code: value.Code, Snapshots: bindingFromWire(value.Snapshots), RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, SelectedCommandIDs: stringsToIDs(value.SelectedCommandIDs), ObligationIDs: stringsToIDs(value.ObligationIDs), PathIDs: stringsToIDs(value.PathIDs), RequiredCommandCount: value.RequiredCommandCount, RequiredObligationCount: value.RequiredObligationCount, VerifiedCommandCount: value.VerifiedCommandCount, VerifiedObligationCount: value.VerifiedObligationCount, VerifiedPathCount: value.VerifiedPathCount, VerifiedCommandIDs: stringsToIDs(value.VerifiedCommandIDs), VerifiedObligationIDs: stringsToIDs(value.VerifiedObligationIDs), VerifiedPathIDs: stringsToIDs(value.VerifiedPathIDs), Digest: value.Digest}
}
func marshalReceipt(value Receipt) ([]byte, error) {
	return json.Marshal(wireReceiptFrom(value))
}
func (value Input) MarshalJSON() ([]byte, error) { return json.Marshal(wireInputFrom(value)) }
func (value *Input) UnmarshalJSON(data []byte) error {
	var wire wireInput
	if err := decodeStrict(data, &wire); err != nil {
		return err
	}
	*value = inputFromWire(wire)
	return nil
}
func (value Receipt) MarshalJSON() ([]byte, error) {
	return marshalReceipt(canonicalReceipt(value))
}
func (value *Receipt) UnmarshalJSON(data []byte) error {
	var wire wireReceipt
	if err := decodeStrict(data, &wire); err != nil {
		return err
	}
	*value = receiptFromWire(wire)
	if value.Digest != value.expectedDigest() {
		return fmt.Errorf("receipt digest does not match canonical receipt")
	}
	return nil
}
func EncodeInput(value Input) ([]byte, error) { return json.Marshal(value) }
