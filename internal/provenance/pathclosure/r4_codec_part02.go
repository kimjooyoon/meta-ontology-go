package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func wireR4Input(value R4Input) r4WireInput {
	records := make([]r4WireRecord, 0, len(value.Records))
	for _, record := range value.Records {
		records = append(records, record.canonicalFields())
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	receipts := make([]r4WireReceipt, 0, len(value.Receipts))
	for _, receipt := range value.Receipts {
		receipts = append(receipts, r4WireReceipt{ID: receipt.ID.String(), EventID: receipt.EventID.String(), RecordID: receipt.RecordID.String(), ProviderID: receipt.ProviderID.String(), ProviderDigest: receipt.ProviderDigest, Phase: string(receipt.Phase), PhaseDigest: receipt.PhaseDigest, ObserverID: receipt.ObserverID.String(), Writes: receipt.Writes, Effect: receipt.Effect})
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].ID < receipts[j].ID })
	paths := make([]r4WirePath, 0, len(value.Paths))
	for _, path := range value.Paths {
		ids := make([]string, 0, len(path.RecordIDs))
		for _, id := range path.RecordIDs {
			ids = append(ids, id.String())
		}
		paths = append(paths, r4WirePath{ID: path.ID.String(), StartID: path.StartID.String(), EndID: path.EndID.String(), RecordIDs: ids, RecordBytes: append([]string(nil), path.RecordBytes...)})
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].ID < paths[j].ID })
	ids := make([]string, 0, len(value.Boundary.RequiredPathIDs))
	for _, id := range sortedR4IDs(value.Boundary.RequiredPathIDs) {
		ids = append(ids, id.String())
	}
	return r4WireInput{Schema: value.Schema, Boundary: r4WireBoundary{RequiredPathIDs: ids, Exhausted: value.Boundary.Exhausted, OpenWorld: value.Boundary.OpenWorld}, Records: records, Receipts: receipts, Paths: paths}
}
func r4InputFromWire(value r4WireInput) R4Input {
	input := R4Input{Schema: value.Schema, Boundary: R4Boundary{Exhausted: value.Boundary.Exhausted, OpenWorld: value.Boundary.OpenWorld}}
	for _, id := range value.Boundary.RequiredPathIDs {
		input.Boundary.RequiredPathIDs = append(input.Boundary.RequiredPathIDs, semantic.ID(id))
	}
	for _, record := range value.Records {
		input.Records = append(input.Records, R4Record{ID: semantic.ID(record.ID), SubjectID: semantic.ID(record.SubjectID), ObjectID: semantic.ID(record.ObjectID), ProviderID: semantic.ID(record.ProviderID), ProviderDigest: record.ProviderDigest, Phase: R4Phase(record.Phase), PhaseDigest: record.PhaseDigest, PredecessorID: semantic.ID(record.PredecessorID), ReceiptID: semantic.ID(record.ReceiptID), Writes: record.Writes, Effect: record.Effect})
	}
	for _, receipt := range value.Receipts {
		input.Receipts = append(input.Receipts, R4Receipt{ID: semantic.ID(receipt.ID), EventID: semantic.ID(receipt.EventID), RecordID: semantic.ID(receipt.RecordID), ProviderID: semantic.ID(receipt.ProviderID), ProviderDigest: receipt.ProviderDigest, Phase: R4Phase(receipt.Phase), PhaseDigest: receipt.PhaseDigest, ObserverID: semantic.ID(receipt.ObserverID), Writes: receipt.Writes, Effect: receipt.Effect})
	}
	for _, path := range value.Paths {
		converted := R4Path{ID: semantic.ID(path.ID), StartID: semantic.ID(path.StartID), EndID: semantic.ID(path.EndID), RecordBytes: append([]string(nil), path.RecordBytes...)}
		for _, id := range path.RecordIDs {
			converted.RecordIDs = append(converted.RecordIDs, semantic.ID(id))
		}
		input.Paths = append(input.Paths, converted)
	}
	return input
}
