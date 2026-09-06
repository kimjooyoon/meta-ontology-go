package valueexecution

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecuteResultUsesActualGoooAndRegisteredOperation(t *testing.T) {
	filename, source := actualValueFixture(t)
	program, err := Compile(filename, source, "Increment")
	if err != nil {
		t.Fatal(err)
	}
	calls := instrumentApply(&program)

	result, err := program.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Fatalf("registered Apply calls = %d, want 1", *calls)
	}
	value, err := result.Integer()
	if err != nil || value != IntegerValue(42) {
		t.Fatalf("result integer = %d, err=%v, want 42", value, err)
	}
	evidence := result.Evidence()
	if evidence.ProducerActivityID != "valuewitness://activity/increment" ||
		evidence.ProducerActivity != "Increment" || evidence.OutputEntityID != "gooo://value-witness/entity/integer" ||
		evidence.OutputEntity != IntegerEntity || evidence.SourceDigest != digestBytes(source) ||
		evidence.SemanticFingerprint != program.SemanticFingerprint || evidence.OperationSpecDigest != program.Operation.SpecDigest ||
		evidence.Value != 42 || !validDigest(evidence.ResultDigest) {
		t.Fatalf("result evidence = %#v", evidence)
	}
	if err := program.ValidateProducedResult(result); err != nil {
		t.Fatalf("compiled producer rejected its result: %v", err)
	}
}

func TestExecuteResultReplayHasEqualContentDigestButTwoApplyCalls(t *testing.T) {
	program := compiledActualValueProgram(t)
	calls := instrumentApply(&program)

	first, err := program.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	second, err := program.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 2 {
		t.Fatalf("registered Apply calls = %d, want 2", *calls)
	}
	if first.Evidence().ResultDigest != second.Evidence().ResultDigest {
		t.Fatalf("replay result digests differ: %q vs %q", first.Evidence().ResultDigest, second.Evidence().ResultDigest)
	}
	if !first.Valid() || !second.Valid() {
		t.Fatal("successful replay did not produce valid handles")
	}
}

func TestProducedResultRejectsZeroAndJSONForgedHandles(t *testing.T) {
	program := compiledActualValueProgram(t)
	issued, err := program.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(issued.Evidence())
	if err != nil {
		t.Fatal(err)
	}
	var forged ProducedResult
	if err := json.Unmarshal(raw, &forged); err == nil {
		t.Fatal("JSON evidence unexpectedly decoded into ProducedResult")
	}
	if forged.Valid() {
		t.Fatal("JSON-forged result handle reported valid")
	}
	err = program.ValidateProducedResult(forged)
	if Reason(err) != ReasonResultHandleInvalid {
		t.Fatalf("JSON-forged reason = %s, want %s", Reason(err), ReasonResultHandleInvalid)
	}
	zeroFailure, ok := FailureOf(err)
	if !ok || zeroFailure.Stage != "RESULT" || zeroFailure.Step != "validate-produced-result" {
		t.Fatalf("zero-destination structured failure = %#v", zeroFailure)
	}

	for name, payload := range map[string][]byte{
		"evidence":     raw,
		"null":         []byte("null"),
		"empty-object": []byte("{}"),
	} {
		t.Run(name, func(t *testing.T) {
			alreadyIssued := issued
			err := json.Unmarshal(payload, &alreadyIssued)
			if err == nil {
				t.Fatal("JSON unexpectedly decoded into an existing ProducedResult")
			}
			if alreadyIssued.Valid() {
				t.Fatal("existing ProducedResult remained usable after rejected JSON decode")
			}
			failure, ok := FailureOf(err)
			if !ok || failure.Code != ReasonResultHandleInvalid || failure.Stage != "RESULT" || failure.Step != "unmarshal-produced-result" {
				t.Fatalf("structured JSON failure = %#v", failure)
			}
		})
	}
}

func TestExecuteResultRejectsCopiedProgramWithoutPrivateAuthorityBeforeApply(t *testing.T) {
	program := compiledActualValueProgram(t)
	fake := Program{
		Activity: program.Activity, Text: program.Text, Operation: program.Operation,
		SourceDigest: program.SourceDigest, SemanticFingerprint: program.SemanticFingerprint,
		ModelProgram: program.ModelProgram, implementation: program.implementation,
	}
	calls := instrumentApply(&fake)

	result, err := fake.ExecuteResult([]int64{41})
	if Reason(err) != ReasonProgramAuthorityInvalid {
		t.Fatalf("reason = %s, want %s", Reason(err), ReasonProgramAuthorityInvalid)
	}
	if result.Valid() || *calls != 0 {
		t.Fatalf("copied program issued result=%#v after %d Apply calls", result.Evidence(), *calls)
	}
	failure, ok := FailureOf(err)
	if !ok || failure.Stage != "EXECUTE" || failure.Step != "validate-result-authority" {
		t.Fatalf("structured failure = %#v", failure)
	}
}

func TestExecuteResultRejectsPublicBindingMutationBeforeApply(t *testing.T) {
	mutations := map[string]func(*Program){
		"activity":               func(program *Program) { program.Activity = "Other" },
		"text":                   func(program *Program) { program.Text = "int.add:2" },
		"source-digest":          func(program *Program) { program.SourceDigest = digestBytes([]byte("other-source")) },
		"semantic-fingerprint":   func(program *Program) { program.SemanticFingerprint = "other-semantic" },
		"model-program":          func(program *Program) { program.ModelProgram = "int.add:2" },
		"operation-activity":     func(program *Program) { program.Operation.Activity = "Other" },
		"operation-input-entity": func(program *Program) { program.Operation.InputEntities[0] = "Other" },
		"operation-spec-digest":  func(program *Program) { program.Operation.SpecDigest = digestBytes([]byte("other-spec")) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			program := compiledActualValueProgram(t)
			calls := instrumentApply(&program)
			mutate(&program)

			result, err := program.ExecuteResult([]int64{41})
			if Reason(err) != ReasonProgramAuthorityMismatch {
				t.Fatalf("reason = %s, want %s", Reason(err), ReasonProgramAuthorityMismatch)
			}
			if result.Valid() || *calls != 0 {
				t.Fatalf("mutated program issued result=%#v after %d Apply calls", result.Evidence(), *calls)
			}
		})
	}
}

func TestProducedResultRejectsChangedSourceOrigin(t *testing.T) {
	filename, source := actualValueFixture(t)
	producer, err := Compile(filename, source, "Increment")
	if err != nil {
		t.Fatal(err)
	}
	changedSource := append(append([]byte(nil), source...), '\n')
	other, err := Compile(filename, changedSource, "Increment")
	if err != nil {
		t.Fatal(err)
	}
	result, err := other.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid() {
		t.Fatal("changed-source producer did not issue a valid result")
	}
	if result.Evidence().ProducerActivityID != "valuewitness://activity/increment" || result.Evidence().SourceDigest == producer.SourceDigest {
		t.Fatalf("changed-source evidence = %#v, producer source digest = %q", result.Evidence(), producer.SourceDigest)
	}
	if err := producer.ValidateProducedResult(result); Reason(err) != ReasonResultProducerMismatch {
		t.Fatalf("reason = %s, want %s", Reason(err), ReasonResultProducerMismatch)
	}
}

func TestProducedResultRejectsDifferentActivityInSameSource(t *testing.T) {
	filename, source := actualValueFixture(t)
	twoActivitySource := []byte(strings.Replace(
		string(source),
		`activity Increment(Integer) -> Integer computes "int.add:1"`,
		`activity Increment(Integer) -> Integer computes "int.add:1"
activity Repeat(Integer) -> Integer computes "int.add:1"`,
		1,
	))
	increment, err := Compile(filename, twoActivitySource, "Increment")
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := Compile(filename, twoActivitySource, "Repeat")
	if err != nil {
		t.Fatal(err)
	}
	calls := instrumentApply(&repeat)
	result, err := repeat.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 1 || !result.Valid() {
		t.Fatalf("same-source second activity calls=%d valid=%t", *calls, result.Valid())
	}
	evidence := result.Evidence()
	if evidence.SourceDigest != increment.SourceDigest || evidence.ProducerActivity != "Repeat" || evidence.ProducerActivityID != "valuewitness://activity/repeat" {
		t.Fatalf("same-source activity evidence = %#v, increment source digest = %q", evidence, increment.SourceDigest)
	}
	if err := increment.ValidateProducedResult(result); Reason(err) != ReasonResultProducerMismatch {
		t.Fatalf("reason = %s, want %s", Reason(err), ReasonResultProducerMismatch)
	}
}

func TestExecuteResultOverflowReturnsNoUsableHandle(t *testing.T) {
	program := compiledActualValueProgram(t)
	calls := instrumentApply(&program)

	result, err := program.ExecuteResult([]int64{math.MaxInt64})
	if Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
	if *calls != 1 || result.Valid() {
		t.Fatalf("overflow result valid=%t after %d Apply calls", result.Valid(), *calls)
	}
	if err := program.ValidateProducedResult(result); Reason(err) != ReasonResultHandleInvalid {
		t.Fatalf("invalid overflow result reason = %s, want %s", Reason(err), ReasonResultHandleInvalid)
	}
}

func TestProducedResultEvidenceIsDetachedFromHandleAuthority(t *testing.T) {
	program := compiledActualValueProgram(t)
	result, err := program.ExecuteResult([]int64{41})
	if err != nil {
		t.Fatal(err)
	}
	want := result.Evidence()
	mutated := want
	mutated.ProducerActivityID = "other-activity"
	mutated.OutputEntityID = "other-entity"
	mutated.SourceDigest = digestBytes([]byte("other-source"))
	mutated.Value = 999
	mutated.ResultDigest = digestBytes([]byte("other-result"))
	if err := program.ValidateProducedResult(result); err != nil {
		t.Fatalf("evidence copy mutation changed authority: %v", err)
	}
	if got := result.Evidence(); got != want {
		t.Fatalf("evidence copy mutated handle authority: got=%#v want=%#v", got, want)
	}
}

func actualValueFixture(t *testing.T) (string, []byte) {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve result test path")
	}
	filename := filepath.Join(filepath.Dir(currentFile), "..", "..", "examples", "language-value-witness", "main.gooo")
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return filename, source
}

func compiledActualValueProgram(t *testing.T) Program {
	t.Helper()
	filename, source := actualValueFixture(t)
	program, err := Compile(filename, source, "Increment")
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func instrumentApply(program *Program) *int {
	original := program.implementation.Apply
	calls := 0
	program.implementation.Apply = func(input, operand int64) (int64, error) {
		calls++
		return original(input, operand)
	}
	return &calls
}
