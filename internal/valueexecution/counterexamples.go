package valueexecution

import "math"

type counterexample struct {
	id       string
	expected string
	run      func() error
}

func executeCounterexamples(filename string) []CounterexampleResult {
	program, _ := Compile(filename, valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`), "Increment")
	fixtures := counterexampleFixtures(filename, program)
	results := make([]CounterexampleResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		actual := Reason(fixture.run())
		replay := Reason(fixture.run())
		results = append(results, CounterexampleResult{
			ID: fixture.id, ExpectedReason: fixture.expected, ActualReason: actual, ReplayReason: replay,
			Passed: actual == fixture.expected, ReplayMatched: actual == replay,
		})
	}
	return results
}

func counterexampleFixtures(filename string, program Program) []counterexample {
	return []counterexample{
		{"missing-program", ReasonProgramMissing, compileCounterexample(filename, `activity Increment(Integer) -> Integer`)},
		{"unknown-operation", ReasonProgramUnknown, compileCounterexample(filename, `activity Increment(Integer) -> Integer computes "int.subtract:1"`)},
		{"malformed-operand", ReasonOperandInvalid, compileCounterexample(filename, `activity Increment(Integer) -> Integer computes "int.add:not-an-int"`)},
		{"unsupported-signature", ReasonSignatureArityUnsupported, compileCounterexample(filename, `activity Increment(Integer, Integer) -> Integer computes "int.add:1"`)},
		{"unresolved-reference", ReasonSemanticBindingFailed, compileCounterexample(filename, `activity Increment(Missing) -> Integer computes "int.add:1"`)},
		{"syntax-error", ReasonSourceParseFailed, compileCounterexample(filename, `activity Increment(Integer) ->`)},
		{"input-arity", ReasonInputArityMismatch, executeCounterexample(program, nil)},
		{"integer-overflow", ReasonIntegerOverflow, executeCounterexample(program, []int64{math.MaxInt64})},
	}
}

func compileCounterexample(filename, activity string) func() error {
	return func() error {
		_, err := Compile(filename, valueFixture(activity), "Increment")
		return err
	}
}

func executeCounterexample(program Program, inputs []int64) func() error {
	return func() error {
		_, err := program.Execute(inputs)
		return err
	}
}
