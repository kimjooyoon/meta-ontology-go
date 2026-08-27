package selfimprovementtermination

import (
	"reflect"
	"strings"
	"testing"
)

func TestClassifiesTerminationWitnesses(t *testing.T) {
	cases := []struct {
		name, decision, reason string
		input                  Input
		period                 int
		proven                 bool
	}{
		{"fixed point", DecisionFixedPoint, "NO_CHANGE_FIXED_POINT_OBSERVED", fixedPointInput(), 0, true},
		{"two cycle", DecisionCycle, "REPEATED_STATE_CYCLE_OBSERVED", cycleInput(), 2, false},
		{"in progress", DecisionInProgress, "TRACE_ENDED_BEFORE_TERMINATION", progressInput(), 0, false},
		{"divergence possible", DecisionDivergence, "STRICTLY_GROWING_BOUNDARY_NO_FIXED_POINT", divergenceInput(), 0, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			receipt, err := Evaluate(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Decision != test.decision || receipt.Reason != test.reason ||
				receipt.Summary.DetectedPeriod != test.period || receipt.Summary.TerminationProven != test.proven ||
				receipt.Summary.Satisfied != IndicatorTotal || receipt.Summary.Total != IndicatorTotal ||
				receipt.Summary.BasisPoints != 10000 || !receipt.Authority.ReadOnly {
				t.Fatalf("receipt = %#v", receipt)
			}
			replay, err := Evaluate(test.input)
			if err != nil || !reflect.DeepEqual(receipt, replay) {
				t.Fatalf("replay diverged: %v %#v", err, replay)
			}
		})
	}
}

func TestRejectsUnsubstantiatedFixedPointClaim(t *testing.T) {
	input := progressInput()
	input.Trace[0].Decision = DecisionFixedPoint
	if _, err := Evaluate(input); err == nil {
		t.Fatal("unsubstantiated fixed-point claim was accepted")
	}
}

func TestCycleTakesPrecedenceOverLaterNoChange(t *testing.T) {
	input := cycleInput()
	input.MaxSteps = 4
	input.Trace = append(input.Trace, Observation{
		Stage: TraceStage, Step: 3, BeforeState: state("a"), AfterState: state("a"),
		BeforeRank: 1, AfterRank: 1, Decision: "NO_CHANGE", Reason: "NO_CHANGE_FIXED_POINT_OBSERVED",
	})
	receipt, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionCycle || receipt.Summary.DetectedPeriod != 2 {
		t.Fatalf("cycle was hidden by later no-change: %#v", receipt)
	}
}

func baseInput(maxSteps int, trace []Observation) Input {
	return Input{Schema: InputSchema, Repository: "kimjooyoon/meta-ontology-go", Subject: Consumer,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: ProofChoice,
		Stage: TraceStage, MaxSteps: maxSteps, Trace: trace}
}

func state(value string) string { return "sha256:" + strings.Repeat(value, 64) }

func changed(step int, before, after string, beforeRank, afterRank int) Observation {
	return Observation{Stage: TraceStage, Step: step, BeforeState: state(before), AfterState: state(after),
		BeforeRank: beforeRank, AfterRank: afterRank, Decision: "CHANGED", Reason: "METAPROGRAM_STATE_CHANGED"}
}

func fixedPointInput() Input {
	return baseInput(4, []Observation{{Stage: TraceStage, Step: 1, BeforeState: state("a"), AfterState: state("a"),
		BeforeRank: 4, AfterRank: 4, Decision: "NO_CHANGE", Reason: "NO_CHANGE_FIXED_POINT_OBSERVED"}})
}

func cycleInput() Input {
	return baseInput(4, []Observation{changed(1, "a", "b", 1, 2), changed(2, "b", "a", 2, 1)})
}

func progressInput() Input {
	return baseInput(4, []Observation{changed(1, "a", "b", 3, 4), changed(2, "b", "c", 4, 4)})
}

func divergenceInput() Input {
	return baseInput(2, []Observation{changed(1, "a", "b", 1, 2), changed(2, "b", "c", 2, 3)})
}
