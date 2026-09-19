package valueexecution

import (
	"math"
	"strings"
	"testing"
	"testing/fstest"
)

func TestEvaluateProducesExactValueWitness(t *testing.T) {
	filesystem := fstest.MapFS{"main.gooo": {Data: valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`)}}
	head := strings.Repeat("a", 40)
	report := Evaluate(filesystem, "main.gooo", "Increment", head)
	if err := Validate(report, head); err != nil {
		t.Fatal(err)
	}
	if report.Scope != RegisteredValueOperationScope || report.Cases[3].Actual != 42 || report.Improvement.Before.Satisfied != 0 || report.Improvement.After.Satisfied != 1 {
		t.Fatalf("value-level evidence is not exact: %#v", report)
	}
}

func TestCompileRejectsUnknownProgramWithoutFallback(t *testing.T) {
	_, err := Compile("unknown.gooo", valueFixture(`activity Increment(Integer) -> Integer computes "int.magic:1"`), "Increment")
	if got := Reason(err); got != ReasonProgramUnknown {
		t.Fatalf("reason = %s, want %s", got, ReasonProgramUnknown)
	}
	failure, ok := FailureOf(err)
	if !ok || failure.Stage != "RESOLVE" || failure.Step != "resolve-operation-spec" {
		t.Fatalf("unknown coordinate = %#v", failure)
	}
}

func TestCompileLowersAndDefendsTypedOperationIR(t *testing.T) {
	program, err := Compile("typed.gooo", valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`), "Increment")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateOperationIR(program.Operation); err != nil || program.Operation.Spec.Effect != EffectPureValue {
		t.Fatalf("operation IR is not exact: %#v / %v", program.Operation, err)
	}
	program.Operation.Spec.Effect = "NETWORK"
	_, err = program.Execute([]int64{1})
	if got := Reason(err); got != ReasonOperationIRInvalid {
		t.Fatalf("tampered IR reason = %s, want %s", got, ReasonOperationIRInvalid)
	}
}

func TestCompileLowersRegisteredSubtractOperation(t *testing.T) {
	program, err := Compile("subtract.gooo", valueFixture(`activity Decrement(Integer) -> Integer computes "int.sub:2"`), "Decrement")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.sub" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("subtract operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{7}); err != nil || got != 5 {
		t.Fatalf("subtract execution = %d / %v, want 5 / nil", got, err)
	}
}

func TestCompileLowersRegisteredMultiplyOperationAndRejectsOverflow(t *testing.T) {
	program, err := Compile("multiply.gooo", valueFixture(`activity Multiply(Integer) -> Integer computes "int.mul:3"`), "Multiply")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.mul" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("multiply operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{7}); err != nil || got != 21 {
		t.Fatalf("multiply execution = %d / %v, want 21 / nil", got, err)
	}
	if _, err := program.Execute([]int64{1 << 62}); Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("multiply overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
}

func TestCompileLowersRegisteredDivideOperationAndRejectsInvalidDivisors(t *testing.T) {
	program, err := Compile("divide.gooo", valueFixture(`activity Divide(Integer) -> Integer computes "int.div:2"`), "Divide")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.div" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("divide operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{8}); err != nil || got != 4 {
		t.Fatalf("divide execution = %d / %v, want 4 / nil", got, err)
	}
	zero, err := Compile("divide-zero.gooo", valueFixture(`activity Divide(Integer) -> Integer computes "int.div:0"`), "Divide")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zero.Execute([]int64{8}); Reason(err) != ReasonIntegerDivisionByZero {
		t.Fatalf("divide-by-zero reason = %s, want %s", Reason(err), ReasonIntegerDivisionByZero)
	}
	overflow, err := Compile("divide-overflow.gooo", valueFixture(`activity Divide(Integer) -> Integer computes "int.div:-1"`), "Divide")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := overflow.Execute([]int64{math.MinInt64}); Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("divide overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
}

func TestCompileRejectsRuntimeBindingsWithoutAPlan(t *testing.T) {
	source := append(valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`),
		[]byte("bind Increment.result -> Increment.input\n")...)
	_, err := Compile("bound.gooo", source, "Increment")
	if got := Reason(err); got != ReasonPlanRequired {
		t.Fatalf("reason = %s, want %s", got, ReasonPlanRequired)
	}
	failure, ok := FailureOf(err)
	if !ok || failure.Stage != "PLAN" || failure.Step != "runtime-binding-plan-required" {
		t.Fatalf("plan boundary = %#v", failure)
	}
}

func TestValidateRejectsDeclaredValueScopeCohort(t *testing.T) {
	filesystem := fstest.MapFS{"main.gooo": {Data: valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`)}}
	head := strings.Repeat("a", 40)
	cases := []struct {
		name  string
		scope string
	}{
		{name: "missing", scope: ""},
		{name: "declaration-resolution-only", scope: "DECLARATION_RESOLUTION_ONLY"},
	}
	const declaredCaseCount = 2
	if len(cases) != declaredCaseCount {
		t.Fatalf("declared value scope regression cases = %d, want %d", len(cases), declaredCaseCount)
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			report := Evaluate(filesystem, "main.gooo", "Increment", head)
			report.Scope = test.scope
			report.Digest = reportDigest(report)
			if err := Validate(report, head); err == nil || err.Error() != "value witness identity is invalid" {
				t.Fatalf("scope %q validation error = %v, want value witness identity is invalid", test.scope, err)
			}
		})
	}
}
