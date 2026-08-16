package pathclosure_test

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// literalR4Oracle is intentionally a closed table: it is not derived from
// EvaluateR4 and cannot learn a new production false positive by reuse.
func literalR4Oracle(name string) (pathclosure.Status, string) {
	return map[string]struct {
			status pathclosure.Status
			code   string
		}{
			"complete":                     {pathclosure.PASS, pathclosure.CodeR4ProofValid},
			"wrong subject":                {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong object endpoint":        {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong canonical record bytes": {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong predecessor":            {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"duplicate receipt":            {pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
			"conflicting receipt":          {pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
			"missing record":               {pathclosure.UNKNOWN, pathclosure.CodeMissingRecord},
			"missing evidence binding":     {pathclosure.UNKNOWN, pathclosure.CodeMissingEvidence},
			"missing provider binding":     {pathclosure.UNKNOWN, pathclosure.CodeMissingProvider},
			"compile labeled runtime":      {pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
			"runtime labeled compile":      {pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
			"stale provider phase digest":  {pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
			"producer only effect claim":   {pathclosure.UNKNOWN, pathclosure.CodeMissingObserver},
		}[name].status, map[string]string{
			"complete": pathclosure.CodeR4ProofValid, "wrong subject": pathclosure.CodeInvalidPath, "wrong object endpoint": pathclosure.CodeInvalidPath,
			"wrong canonical record bytes": pathclosure.CodeInvalidPath, "wrong predecessor": pathclosure.CodeInvalidPath, "duplicate receipt": pathclosure.CodeConflictingReceipt,
			"conflicting receipt": pathclosure.CodeConflictingReceipt, "missing record": pathclosure.CodeMissingRecord, "missing evidence binding": pathclosure.CodeMissingEvidence,
			"missing provider binding": pathclosure.CodeMissingProvider, "compile labeled runtime": pathclosure.CodePhaseMismatch, "runtime labeled compile": pathclosure.CodePhaseMismatch,
			"stale provider phase digest": pathclosure.CodePhaseMismatch, "producer only effect claim": pathclosure.CodeMissingObserver,
		}[name]
}

// plainFiniteGraphBaseline is deliberately weaker than R4: it checks only a
// finite ordered graph and never pretends to validate receipts or phases.
func plainFiniteGraphBaseline(input pathclosure.R4Input) bool {
	if input.Boundary.OpenWorld || !input.Boundary.Exhausted || len(input.Boundary.RequiredPathIDs) == 0 {
		return false
	}
	records := map[semantic.ID]pathclosure.R4Record{}
	for _, record := range input.Records {
		records[record.ID] = record
	}
	paths := map[semantic.ID]pathclosure.R4Path{}
	for _, path := range input.Paths {
		paths[path.ID] = path
	}
	for _, required := range input.Boundary.RequiredPathIDs {
		path, ok := paths[required]
		if !ok || len(path.RecordIDs) == 0 || len(path.RecordIDs) != len(path.RecordBytes) {
			return false
		}
		var previous pathclosure.R4Record
		for index, recordID := range path.RecordIDs {
			record, ok := records[recordID]
			if !ok {
				return false
			}
			if index == 0 {
				if record.PredecessorID != "" || record.SubjectID != path.StartID {
					return false
				}
			} else if record.PredecessorID != previous.ID || previous.ObjectID != record.SubjectID {
				return false
			}
			if index == len(path.RecordIDs)-1 && record.ObjectID != path.EndID {
				return false
			}
			previous = record
		}
	}
	return true
}

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
