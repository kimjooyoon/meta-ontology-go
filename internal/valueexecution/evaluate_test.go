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

func TestCompileLowersRegisteredModuloOperationAndRejectsZeroDivisor(t *testing.T) {
	program, err := Compile("modulo.gooo", valueFixture(`activity Remainder(Integer) -> Integer computes "int.mod:3"`), "Remainder")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.mod" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("modulo operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{8}); err != nil || got != 2 {
		t.Fatalf("modulo execution = %d / %v, want 2 / nil", got, err)
	}
	zero, err := Compile("modulo-zero.gooo", valueFixture(`activity Remainder(Integer) -> Integer computes "int.mod:0"`), "Remainder")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zero.Execute([]int64{8}); Reason(err) != ReasonIntegerModuloByZero {
		t.Fatalf("modulo-by-zero reason = %s, want %s", Reason(err), ReasonIntegerModuloByZero)
	}
	minimum, err := Compile("modulo-minimum.gooo", valueFixture(`activity Remainder(Integer) -> Integer computes "int.mod:-1"`), "Remainder")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := minimum.Execute([]int64{math.MinInt64}); err != nil || got != 0 {
		t.Fatalf("minimum modulo execution = %d / %v, want 0 / nil", got, err)
	}
}

func TestCompileLowersRegisteredNegateOperationAndRejectsInvalidOperandAndMinimum(t *testing.T) {
	program, err := Compile("negate.gooo", valueFixture(`activity Negate(Integer) -> Integer computes "int.neg:0"`), "Negate")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.neg" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("negate operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{7}); err != nil || got != -7 {
		t.Fatalf("negate execution = %d / %v, want -7 / nil", got, err)
	}
	if _, err := program.Execute([]int64{math.MinInt64}); Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("negate overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
	invalid, err := Compile("negate-invalid.gooo", valueFixture(`activity Negate(Integer) -> Integer computes "int.neg:1"`), "Negate")
	if err == nil || invalid.Operation.Spec.ID != "" || Reason(err) != ReasonOperationIRInvalid {
		t.Fatalf("invalid negate operand = %#v / %v, want operation IR invalid", invalid, err)
	}
}

func TestCompileLowersRegisteredAbsoluteOperationAndRejectsInvalidOperandAndMinimum(t *testing.T) {
	program, err := Compile("absolute.gooo", valueFixture(`activity Absolute(Integer) -> Integer computes "int.abs:0"`), "Absolute")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.abs" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("absolute operation contract is not closed: %#v", program.Operation.Spec)
	}
	if got, err := program.Execute([]int64{-7}); err != nil || got != 7 {
		t.Fatalf("absolute execution = %d / %v, want 7 / nil", got, err)
	}
	if _, err := program.Execute([]int64{math.MinInt64}); Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("absolute overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
	invalid, err := Compile("absolute-invalid.gooo", valueFixture(`activity Absolute(Integer) -> Integer computes "int.abs:1"`), "Absolute")
	if err == nil || invalid.Operation.Spec.ID != "" || Reason(err) != ReasonOperationIRInvalid {
		t.Fatalf("invalid absolute operand = %#v / %v, want operation IR invalid", invalid, err)
	}
}

func TestCompileLowersRegisteredSignOperationAndRejectsInvalidOperand(t *testing.T) {
	program, err := Compile("sign.gooo", valueFixture(`activity Sign(Integer) -> Integer computes "int.sign:0"`), "Sign")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.sign" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("sign operation contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct {
		input int64
		want  int64
	}{
		{input: -7, want: -1},
		{input: 0, want: 0},
		{input: 7, want: 1},
	} {
		if got, err := program.Execute([]int64{test.input}); err != nil || got != test.want {
			t.Fatalf("sign execution for %d = %d / %v, want %d / nil", test.input, got, err, test.want)
		}
	}
	invalid, err := Compile("sign-invalid.gooo", valueFixture(`activity Sign(Integer) -> Integer computes "int.sign:1"`), "Sign")
	if err == nil || invalid.Operation.Spec.ID != "" || Reason(err) != ReasonOperationIRInvalid {
		t.Fatalf("invalid sign operand = %#v / %v, want operation IR invalid", invalid, err)
	}
}

func TestCompileLowersRegisteredMaximumOperation(t *testing.T) {
	program, err := Compile("maximum.gooo", valueFixture(`activity AtLeastZero(Integer) -> Integer computes "int.max:0"`), "AtLeastZero")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.max" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("maximum operation contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct{ input, want int64 }{{-7, 0}, {7, 7}} {
		got, err := program.Execute([]int64{test.input})
		if err != nil || got != test.want {
			t.Fatalf("maximum execution = %d / %v, want %d / nil", got, err, test.want)
		}
	}
}

func TestCompileLowersRegisteredMinimumOperation(t *testing.T) {
	program, err := Compile("minimum.gooo", valueFixture(`activity AtMostZero(Integer) -> Integer computes "int.min:0"`), "AtMostZero")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.min" || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("minimum operation contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct{ input, want int64 }{{-7, -7}, {7, 0}} {
		got, err := program.Execute([]int64{test.input})
		if err != nil || got != test.want {
			t.Fatalf("minimum execution = %d / %v, want %d / nil", got, err, test.want)
		}
	}
}

func TestCompileLowersTypedBooleanOperation(t *testing.T) {
	source := []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Integer id \"gooo://value-witness/entity/integer\"\n" +
		"entity Boolean id \"gooo://value-witness/entity/boolean\"\n\n" +
		"activity IsZero(Integer) -> Boolean computes \"int.iszero:0\"\n")
	program, err := Compile("iszero.gooo", source, "IsZero")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.iszero" || program.Operation.Spec.OutputEntity != BooleanEntity {
		t.Fatalf("boolean operation contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct {
		input int64
		want  bool
	}{{0, true}, {7, false}} {
		result, err := program.ExecuteResult([]int64{test.input})
		if err != nil {
			t.Fatalf("execute iszero(%d): %v", test.input, err)
		}
		got, err := result.Boolean()
		if err != nil || got != test.want {
			t.Fatalf("iszero(%d) = %t / %v, want %t / nil", test.input, got, err, test.want)
		}
	}
}

func TestCompileLowersBooleanNotOperation(t *testing.T) {
	source := []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Boolean id \"gooo://value-witness/entity/boolean\"\n\n" +
		"activity Not(Boolean) -> Boolean computes \"bool.not:0\"\n")
	program, err := Compile("not.gooo", source, "Not")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "bool.not" || program.Operation.Spec.InputEntities[0] != BooleanEntity || program.Operation.Spec.OutputEntity != BooleanEntity {
		t.Fatalf("boolean not contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct {
		input int64
		want  bool
	}{{0, true}, {1, false}} {
		result, err := program.ExecuteResult([]int64{test.input})
		if err != nil {
			t.Fatalf("execute not(%d): %v", test.input, err)
		}
		got, err := result.Boolean()
		if err != nil || got != test.want {
			t.Fatalf("not(%d) = %t / %v, want %t / nil", test.input, got, err, test.want)
		}
	}
}

func TestCompileLowersBooleanAndOperationAndRejectsNonBooleanOperand(t *testing.T) {
	source := []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Boolean id \"gooo://value-witness/entity/boolean\"\n\n" +
		"activity And(Boolean) -> Boolean computes \"bool.and:1\"\n")
	program, err := Compile("and.gooo", source, "And")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "bool.and" || program.Operation.Spec.InputEntities[0] != BooleanEntity || program.Operation.Spec.OutputEntity != BooleanEntity {
		t.Fatalf("boolean and contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct {
		input int64
		want  bool
	}{{0, false}, {1, true}} {
		result, err := program.ExecuteResult([]int64{test.input})
		if err != nil {
			t.Fatalf("execute and(%d): %v", test.input, err)
		}
		got, err := result.Boolean()
		if err != nil || got != test.want {
			t.Fatalf("and(%d) = %t / %v, want %t / nil", test.input, got, err, test.want)
		}
	}
	invalid, err := Compile("and-invalid.gooo", []byte(strings.Replace(string(source), "bool.and:1", "bool.and:2", 1)), "And")
	if err == nil || invalid.Operation.Spec.ID != "" || Reason(err) != ReasonOperationIRInvalid {
		t.Fatalf("invalid boolean operand = %#v / %v, want operation IR invalid", invalid, err)
	}
}

func TestCompileLowersRegisteredEqualOperationToBoolean(t *testing.T) {
	source := []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Integer id \"gooo://value-witness/entity/integer\"\n" +
		"entity Boolean id \"gooo://value-witness/entity/boolean\"\n\n" +
		"activity IsSeven(Integer) -> Boolean computes \"int.eq:7\"\n")
	program, err := Compile("equal.gooo", source, "IsSeven")
	if err != nil {
		t.Fatal(err)
	}
	if program.Operation.Spec.ID != "int.eq" || program.Operation.Spec.InputEntities[0] != IntegerEntity || program.Operation.Spec.OutputEntity != BooleanEntity || program.Operation.Spec.Effect != EffectPureValue || program.Operation.Spec.Determinism != Deterministic {
		t.Fatalf("equal operation contract is not closed: %#v", program.Operation.Spec)
	}
	for _, test := range []struct {
		input int64
		want  bool
	}{{7, true}, {6, false}} {
		result, err := program.ExecuteResult([]int64{test.input})
		if err != nil {
			t.Fatalf("execute equal(%d): %v", test.input, err)
		}
		got, err := result.Boolean()
		if err != nil || got != test.want {
			t.Fatalf("equal(%d) = %t / %v, want %t / nil", test.input, got, err, test.want)
		}
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
