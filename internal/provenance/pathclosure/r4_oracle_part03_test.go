package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"testing"
)

func TestR4IndependentLiteralOracleAndFairFiniteGraphBaseline(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*pathclosure.R4Input)
	}{
		{"complete", nil},
		{"wrong subject", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.SubjectID = r4ID("node/forged") })
			refreshR4RecordBytes(input)
		}},
		{"wrong object endpoint", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/compile"), func(record *pathclosure.R4Record) { record.ObjectID = r4ID("node/forged") })
			refreshR4RecordBytes(input)
		}},
		{"wrong predecessor", func(input *pathclosure.R4Input) {
			mutateR4Record(input, r4ID("record/runtime"), func(record *pathclosure.R4Record) { record.PredecessorID = r4ID("record/wrong") })
			refreshR4RecordBytes(input)
		}},
		{"missing record", func(input *pathclosure.R4Input) { input.Records = input.Records[:1] }},
		{"stale provider phase digest", func(input *pathclosure.R4Input) {
			mutateR4Receipt(input, r4ID("receipt/runtime"), func(receipt *pathclosure.R4Receipt) { receipt.PhaseDigest = r4Digest("stale") })
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := cloneR4Input(completeR4Fixture().input)
			if test.mutate != nil {
				test.mutate(&input)
			}
			wantStatus, wantCode := literalR4Oracle(test.name)
			got := pathclosure.EvaluateR4(input)
			if got.Status != wantStatus || got.Code != wantCode {
				t.Fatalf("production=%s/%s literal=%s/%s", got.Status, got.Code, wantStatus, wantCode)
			}
			baseline := plainFiniteGraphBaseline(input)
			if test.name == "complete" && !baseline {
				t.Fatal("plain finite graph baseline rejected the complete graph")
			}
			if test.name != "stale provider phase digest" && test.name != "missing record" && test.name != "complete" && baseline {
				t.Fatal("plain finite graph baseline accepted a forged topology")
			}
			if test.name == "stale provider phase digest" && !baseline {
				t.Fatal("fair graph baseline modeled a receipt/phase rule it does not own")
			}
		})
	}
}
