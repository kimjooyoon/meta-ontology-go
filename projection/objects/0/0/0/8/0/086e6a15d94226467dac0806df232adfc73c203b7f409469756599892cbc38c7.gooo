package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"strings"
	"testing"
)

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
