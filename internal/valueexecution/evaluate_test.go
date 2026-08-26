package valueexecution

import (
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
	if report.Cases[3].Actual != 42 || report.Improvement.Before.Satisfied != 0 || report.Improvement.After.Satisfied != 1 {
		t.Fatalf("value-level evidence is not exact: %#v", report)
	}
}

func TestCompileRejectsUnknownProgramWithoutFallback(t *testing.T) {
	_, err := Compile("unknown.gooo", valueFixture(`activity Increment(Integer) -> Integer computes "int.magic:1"`), "Increment")
	if got := Reason(err); got != ReasonProgramUnknown {
		t.Fatalf("reason = %s, want %s", got, ReasonProgramUnknown)
	}
}
