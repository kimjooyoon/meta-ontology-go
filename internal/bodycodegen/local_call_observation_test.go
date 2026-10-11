package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func observedCallsFixture(t *testing.T, body string) Result {
	t.Helper()
	source := strings.Replace(pureCallsFixture,
		"if input < 0 { return Twice(-input) } else { return Add(Twice(input), 3) }", body, 1)
	result, err := Generate("observed.gooo", []byte(source), "Score")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestLocalCallObservationRetainsFailurePrefix(t *testing.T) {
	r := observedCallsFixture(t, "return Add(Twice(input), 1 / input)")
	cases := []IRBodyFillTestCase{{Input: 1, Expected: 3}, {Input: 0}, {Input: 2}}
	rows, passed, calls, err := evaluateIntegerCasesObserved(context.Background(), []byte(r.Source), r.Report, cases)
	if err == nil || !strings.Contains(err.Error(), "zero divisor") || rows != nil || passed != 0 {
		t.Fatal("failed scoring contract changed", rows, passed, err)
	}
	if calls.Status != "RECORDED" || calls.EvaluationCompleted || calls.CasesStarted != 2 || calls.CasesCompleted != 1 ||
		calls.EvaluationError != err.Error() || len(calls.Calls) != 3 {
		t.Fatal("failed case or prior successful case exposure was lost", calls)
	}
	for i, want := range []string{"[1]", "[2,1]", "[0]"} {
		if string(calls.Calls[i].Inputs) != want || calls.Calls[i].ProgramSHA256 == "" ||
			calls.Calls[i].CaseIndex != []int{0, 0, 1}[i] {
			t.Fatal("argument evaluation order or exact typed inputs changed", calls.Calls)
		}
	}
	if calls.Calls[0].ActivityID != calls.Calls[2].ActivityID || calls.Calls[0].ActivityID == calls.Calls[1].ActivityID {
		t.Fatal("callee stable identities were not preserved", calls.Calls)
	}
}

func TestLocalCallObservationPreservesRepeatedAndLargeIntegerInputs(t *testing.T) {
	r := observedCallsFixture(t, "return Add(Twice(input), Twice(input))")
	cases := []IRBodyFillTestCase{{Input: 9007199254740993, Expected: 36028797018963972}}
	rows, passed, calls, err := evaluateIntegerCasesObserved(context.Background(), []byte(r.Source), r.Report, cases)
	if err != nil || passed != 1 || len(rows) != 1 || !calls.EvaluationCompleted || len(calls.Calls) != 3 {
		t.Fatal(calls, err)
	}
	if string(calls.Calls[0].Inputs) != "[9007199254740993]" || !reflect.DeepEqual(calls.Calls[0], calls.Calls[1]) ||
		string(calls.Calls[2].Inputs) != "[18014398509481986,18014398509481986]" {
		t.Fatal("integer precision or repeated observations lost", calls.Calls)
	}
	encoded, _ := json.Marshal(calls)
	var decoded LocalCallObservation
	if err := json.Unmarshal(encoded, &decoded); err != nil || !reflect.DeepEqual(*calls, decoded) {
		t.Fatal("observation does not survive saved JSON", err)
	}
}

func TestLocalCallObservationSnapshotsRecordValues(t *testing.T) {
	source := []byte(`package observation
namespace observation
entity Integer id "observation://integer"
entity Box id "observation://box" fields {
 field amount id "observation://amount" type integer required one
 field note id "observation://note" type string optional one
 field active id "observation://active" type boolean optional one
}
activity Copy(Box, Integer) -> Integer computes "let copy = input0; copy.amount = input1; return copy.amount"
activity Run(Integer) -> Integer computes "let text = \"\"; let active = false; let item = Box{amount: input, note: &text, active: &active}; let first = Copy(item, 0); item.note = nil; item.active = nil; return first + Copy(item, input)"
`)
	r, err := Generate("record.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	_, passed, calls, err := evaluateIntegerCasesObserved(context.Background(), []byte(r.Source), r.Report,
		[]IRBodyFillTestCase{{Input: 7, Expected: 7}})
	if err != nil || passed != 1 || calls.Status != "RECORDED" || len(calls.Calls) != 2 {
		t.Fatal(calls, err)
	}
	for i, want := range []string{`[{"active":false,"amount":7,"note":""},0]`, `[{"active":null,"amount":7,"note":null},7]`} {
		if string(calls.Calls[i].Inputs) != want {
			t.Fatal("call entry value aliased a later assignment or lost optional presence", calls.Calls)
		}
	}
}

func TestLocalCallObservationBoundsAndCancellation(t *testing.T) {
	r := observedCallsFixture(t, "return Twice(input)")
	cases := []IRBodyFillTestCase{{Input: 1, Expected: 2}}
	for _, tc := range []struct{ name, reason string }{{"records", "CALL_RECORD_LIMIT"}, {"bytes", "CALL_INPUT_BYTE_LIMIT"}, {"identity", "CALL_IDENTITY_UNAVAILABLE"}} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := newLocalCallRecorder([]byte(r.Source), r.Report, cases)
			switch tc.name {
			case "records":
				recorder.observation.Calls = make([]LocalCallInput, localCallRecordLimit)
			case "bytes":
				recorder.bytes = localCallByteLimit
			case "identity":
				clear(recorder.identities)
			}
			_, passed, err := evaluateIntegerCasesRecorded(context.Background(), []byte(r.Source), r.Report.Activity, cases, recorder)
			if err != nil || passed != 1 || recorder.observation.Status != "UNKNOWN" || recorder.observation.Reason != tc.reason {
				t.Fatal("bounded observation changed scoring or claimed complete capture", recorder.observation, err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, calls, err := evaluateIntegerCasesObserved(ctx, []byte(r.Source), r.Report, cases)
	if err != context.Canceled || calls.Status != "UNKNOWN" || calls.Reason != "CALLER_CANCELLED" || len(calls.Calls) != 0 {
		t.Fatal("cancellation gained input evidence", calls, err)
	}
}
