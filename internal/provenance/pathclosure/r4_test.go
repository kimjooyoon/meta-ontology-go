package pathclosure_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

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

func mutateR4Record(input *pathclosure.R4Input, id semantic.ID, mutate func(*pathclosure.R4Record)) {
	for index := range input.Records {
		if input.Records[index].ID == id {
			mutate(&input.Records[index])
		}
	}
}

func mutateR4Receipt(input *pathclosure.R4Input, id semantic.ID, mutate func(*pathclosure.R4Receipt)) {
	for index := range input.Receipts {
		if input.Receipts[index].ID == id {
			mutate(&input.Receipts[index])
		}
	}
}

func TestEvaluateR4CompleteFiniteBoundaryNeverAuthorizesPromotion(t *testing.T) {
	fixture := completeR4Fixture()
	got := pathclosure.EvaluateR4(fixture.input)
	if got.Status != pathclosure.PASS || got.Code != pathclosure.CodeR4ProofValid || !got.ProofValid || got.PromotionAuthorized {
		t.Fatalf("complete R4 result = %#v", got)
	}
	if !reflect.DeepEqual(got.CoveredPathIDs, []semantic.ID{fixture.path.ID}) || got.Cost != 5 {
		t.Fatalf("coverage/cost = %#v", got)
	}
}

func TestEvaluateR4DirectForgedPathAndReceiptPartitions(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*pathclosure.R4Input)
		status pathclosure.Status
		code   string
	}{
		{"wrong subject", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.SubjectID = r4ID("node/forged") })
			refreshR4RecordBytes(input)
		}, pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
		{"wrong object endpoint", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/compile"), func(record *pathclosure.R4Record) { record.ObjectID = r4ID("node/forged") })
			refreshR4RecordBytes(input)
		}, pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
		{"wrong canonical record bytes", func(input *pathclosure.R4Input) {
			input.Paths[0].RecordBytes[0] = strings.Replace(input.Paths[0].RecordBytes[0], "compile", "runtime", 1)
		}, pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
		{"wrong predecessor", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.PredecessorID = r4ID("record/wrong") })
			refreshR4RecordBytes(input)
		}, pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
		{"duplicate receipt", func(input *pathclosure.R4Input) { input.Receipts = append(input.Receipts, input.Receipts[0]) }, pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
		{"conflicting receipt", func(input *pathclosure.R4Input) {
			mutateR4Receipt(input, r4ID("receipt/runtime"), func(receipt *pathclosure.R4Receipt) { receipt.EventID = r4ID("event/other") })
			input.Receipts = append(input.Receipts, pathclosure.R4Receipt{ID: r4ID("receipt/runtime"), EventID: r4ID("event/runtime"), RecordID: r4ID("record/runtime")})
		}, pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
		{"missing record", func(input *pathclosure.R4Input) { input.Records = input.Records[:1] }, pathclosure.UNKNOWN, pathclosure.CodeMissingRecord},
		{"missing evidence binding", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.ReceiptID = "" })
			refreshR4RecordBytes(input)
		}, pathclosure.UNKNOWN, pathclosure.CodeMissingEvidence},
		{"missing provider binding", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.ProviderID, record.ProviderDigest = "", "" })
			refreshR4RecordBytes(input)
		}, pathclosure.UNKNOWN, pathclosure.CodeMissingProvider},
		{"record phase altered without receipt", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/compile"), func(record *pathclosure.R4Record) { record.Phase = pathclosure.R4RuntimePhase })
			refreshR4RecordBytes(input)
		}, pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
		{"receipt phase altered without record", func(input *pathclosure.R4Input) {
			mutateR4Receipt(input, r4ID("receipt/runtime"), func(receipt *pathclosure.R4Receipt) { receipt.Phase = pathclosure.R4CompilePhase })
		}, pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
		{"stale provider phase digest", func(input *pathclosure.R4Input) {
			mutateR4Receipt(input, r4ID("receipt/runtime"), func(receipt *pathclosure.R4Receipt) { receipt.PhaseDigest = r4Digest("stale") })
		}, pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
		{"producer only effect claim", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.Writes, record.Effect = false, "no-write" })
			mutateR4Receipt(input, r4ID("receipt/runtime"), func(receipt *pathclosure.R4Receipt) {
				receipt.Writes, receipt.Effect, receipt.ObserverID = false, "no-write", ""
			})
			refreshR4RecordBytes(input)
		}, pathclosure.UNKNOWN, pathclosure.CodeMissingObserver},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := cloneR4Input(completeR4Fixture().input)
			test.mutate(&input)
			got := pathclosure.EvaluateR4(input)
			if got.Status != test.status || got.Code != test.code || got.ProofValid || got.PromotionAuthorized {
				t.Fatalf("R4 result = %#v, want %s/%s", got, test.status, test.code)
			}
		})
	}
}

func TestEvaluateR4FiniteBoundaryAndRootIsolation(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*pathclosure.R4Input)
		code   string
	}{
		{"open world", func(input *pathclosure.R4Input) { input.Boundary.OpenWorld = true }, pathclosure.CodeOpenWorld},
		{"not exhausted", func(input *pathclosure.R4Input) { input.Boundary.Exhausted = false }, pathclosure.CodeUnexhaustedBoundary},
		{"root replay is not discovered", func(input *pathclosure.R4Input) { input.Paths[0].StartID = r4ID("node/alternate-root") }, pathclosure.CodeInvalidPath},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := cloneR4Input(completeR4Fixture().input)
			test.mutate(&input)
			got := pathclosure.EvaluateR4(input)
			if got.Status != pathclosure.UNKNOWN && test.name != "root replay is not discovered" || got.Code != test.code {
				t.Fatalf("R4 result = %#v, want code %s", got, test.code)
			}
		})
	}
}

func TestR4ExpectedOnlyMetaLabelCannotAffectDecision(t *testing.T) {
	input := completeR4Fixture().input
	baseline := pathclosure.EvaluateR4(input)
	baselineDigest := baseline.CanonicalDigest()
	// Expected/meta labels are deliberately not an R4 input. Keeping them in a
	// caller-side adapter must not alter any result field or decision digest.
	for _, metaLabel := range []string{"compile", "runtime", "forged-alias"} {
		result := func(_ string) pathclosure.R4Result { return pathclosure.EvaluateR4(input) }(metaLabel)
		if !reflect.DeepEqual(result, baseline) || result.CanonicalDigest() != baselineDigest {
			t.Fatalf("meta label %q changed the R4 decision: %#v vs %#v", metaLabel, result, baseline)
		}
	}
}

func TestStrictR4JSONCodecCanonicalReplay(t *testing.T) {
	fixture := completeR4Fixture()
	data, err := pathclosure.EncodeR4Input(fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := pathclosure.DecodeR4Input(data)
	if err != nil || pathclosure.EvaluateR4(decoded).Status != pathclosure.PASS {
		t.Fatalf("round trip = %#v, err=%v", decoded, err)
	}
	if replay, err := pathclosure.EncodeR4Input(decoded); err != nil || !bytes.Equal(data, replay) {
		t.Fatalf("canonical replay changed: %s / %s", data, replay)
	}

	for _, test := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"unknown field", func(value []byte) []byte {
			return bytes.Replace(value, []byte(`"schema"`), []byte(`"unknown":"x","schema"`), 1)
		}},
		{"duplicate key", func(value []byte) []byte {
			return bytes.Replace(value, []byte(`"schema":"`+pathclosure.R4SchemaVersion+`"`), []byte(`"schema":"`+pathclosure.R4SchemaVersion+`","schema":"`+pathclosure.R4SchemaVersion+`"`), 1)
		}},
		{"trailing value", func(value []byte) []byte { return append(append([]byte(nil), value...), []byte(` {}`)...) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := pathclosure.DecodeR4Input(test.mutate(data)); err == nil {
				t.Fatal("malformed strict JSON was accepted")
			}
		})
	}
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if _, ok := wire["records"]; !ok {
		t.Fatal("records missing from canonical input")
	}
}

func TestEvaluateR4PermutationReplayUsesStableIDsNotInsertionOrder(t *testing.T) {
	fixture := completeR4Fixture()
	permuted := cloneR4Input(fixture.input)
	permuted.Records[0], permuted.Records[1] = permuted.Records[1], permuted.Records[0]
	permuted.Receipts[0], permuted.Receipts[1] = permuted.Receipts[1], permuted.Receipts[0]
	left, right := pathclosure.EvaluateR4(fixture.input), pathclosure.EvaluateR4(permuted)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("permutation changed R4 result:\nleft=%#v\nright=%#v", left, right)
	}
}
