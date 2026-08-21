package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func r4ID(value string) semantic.ID { return semantic.MustIdentity("pathclosure-r4-test://" + value) }
func r4Digest(value string) string  { return semantic.StableHashString("pathclosure-r4-test/" + value) }

type r4Fixture struct {
	input   pathclosure.R4Input
	path    pathclosure.R4Path
	records []pathclosure.R4Record
}

func completeR4Fixture() r4Fixture {
	provider, providerDigest := r4ID("provider/runner"), r4Digest("provider")
	phaseDigest := r4Digest("phase")
	root, middle, end := r4ID("node/root"), r4ID("node/middle"), r4ID("node/end")
	records := []pathclosure.R4Record{
		{ID: r4ID("record/compile"), SubjectID: root, ObjectID: middle, ProviderID: provider, ProviderDigest: providerDigest, Phase: pathclosure.R4CompilePhase, PhaseDigest: phaseDigest, ReceiptID: r4ID("receipt/compile"), Writes: true},
		{ID: r4ID("record/runtime"), SubjectID: middle, ObjectID: end, ProviderID: provider, ProviderDigest: providerDigest, Phase: pathclosure.R4RuntimePhase, PhaseDigest: phaseDigest, PredecessorID: r4ID("record/compile"), ReceiptID: r4ID("receipt/runtime"), Writes: true},
	}
	receipts := []pathclosure.R4Receipt{
		{ID: r4ID("receipt/compile"), EventID: r4ID("event/compile"), RecordID: records[0].ID, ProviderID: provider, ProviderDigest: providerDigest, Phase: pathclosure.R4CompilePhase, PhaseDigest: phaseDigest, Writes: true},
		{ID: r4ID("receipt/runtime"), EventID: r4ID("event/runtime"), RecordID: records[1].ID, ProviderID: provider, ProviderDigest: providerDigest, Phase: pathclosure.R4RuntimePhase, PhaseDigest: phaseDigest, Writes: true},
	}
	path := pathclosure.R4Path{ID: r4ID("path/main"), StartID: root, EndID: end}
	for _, record := range records {
		path.RecordIDs = append(path.RecordIDs, record.ID)
		bytesValue, err := record.CanonicalRecordBytes()
		if err != nil {
			panic(err)
		}
		path.RecordBytes = append(path.RecordBytes, string(bytesValue))
	}
	input := pathclosure.R4Input{Schema: pathclosure.R4SchemaVersion, Boundary: pathclosure.R4Boundary{RequiredPathIDs: []semantic.ID{path.ID}, Exhausted: true}, Records: records, Receipts: receipts, Paths: []pathclosure.R4Path{path}}
	return r4Fixture{input: input, path: path, records: records}
}
func cloneR4Input(value pathclosure.R4Input) pathclosure.R4Input {
	copy := value
	copy.Boundary.RequiredPathIDs = append([]semantic.ID(nil), value.Boundary.RequiredPathIDs...)
	copy.Records = append([]pathclosure.R4Record(nil), value.Records...)
	copy.Receipts = append([]pathclosure.R4Receipt(nil), value.Receipts...)
	copy.Paths = append([]pathclosure.R4Path(nil), value.Paths...)
	for index := range copy.Paths {
		copy.Paths[index].RecordIDs = append([]semantic.ID(nil), value.Paths[index].RecordIDs...)
		copy.Paths[index].RecordBytes = append([]string(nil), value.Paths[index].RecordBytes...)
	}
	return copy
}
func refreshR4RecordBytes(input *pathclosure.R4Input) {
	for pathIndex := range input.Paths {
		for recordIndex, recordID := range input.Paths[pathIndex].RecordIDs {
			for _, record := range input.Records {
				if record.ID != recordID {
					continue
				}
				bytesValue, err := record.CanonicalRecordBytes()
				if err != nil {
					panic(err)
				}
				input.Paths[pathIndex].RecordBytes[recordIndex] = string(bytesValue)
			}
		}
	}
}
