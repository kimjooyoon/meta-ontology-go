package selfimprovementtermination

import (
	"reflect"
	"testing"
)

func TestClassifiesTerminationWitnesses(t *testing.T) {
	cases := []struct {
		name, decision, resolution, claim, reason string
		input                                     Input
		period, states                            int
		proven                                    bool
	}{
		{"fixed point", DecisionFixedPoint, ResolutionExact, ClaimDischarged, ReasonNoChange, fixedPointInput(), 0, 1, true},
		{"two cycle", DecisionCycle, ResolutionExact, ClaimRefuted, ReasonCycle, cycleInput(), 2, 3, false},
		{"in progress", DecisionInProgress, ResolutionLower, ClaimOpen, ReasonInProgress, progressInput(), 0, 3, false},
		{"divergence possible", DecisionDivergence, ResolutionLower, ClaimOpen, ReasonDivergence, divergenceInput(), 0, 3, false},
		{"unknown upstream", DecisionFailClosed, ResolutionLower, ClaimOpen, ReasonDecisionUnknown, unknownInput(), 0, 1, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			receipt, err := Evaluate(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Decision != test.decision || receipt.Resolution != test.resolution ||
				receipt.Reason != test.reason || receipt.Outcome.ClaimState != test.claim ||
				receipt.Outcome.DetectedPeriod != test.period || receipt.Outcome.StateCount != test.states ||
				receipt.Outcome.TerminationProven != test.proven || receipt.Conformance.Satisfied != IndicatorTotal ||
				receipt.Conformance.Total != IndicatorTotal || receipt.Conformance.BasisPoints != 10000 || !receipt.Authority.ReadOnly {
				t.Fatalf("receipt = %#v", receipt)
			}
			replay, err := Evaluate(test.input)
			if err != nil || !reflect.DeepEqual(receipt, replay) {
				t.Fatalf("replay diverged: %v %#v", err, replay)
			}
		})
	}
}

func TestUnknownUpstreamEmitsLocalizedFailClosedReceipt(t *testing.T) {
	receipt, err := Evaluate(unknownInput())
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != ReceiptFailClosed || receipt.Resolution != ResolutionLower || receipt.Reason != ReasonDecisionUnknown || receipt.Outcome.ClaimState != ClaimOpen {
		t.Fatalf("unknown upstream was not fail-closed: %#v", receipt)
	}
}

func TestCycleTakesPrecedenceOverLaterNoChange(t *testing.T) {
	input := cycleInput()
	input.MaxSteps = 4
	input.Trace = append(input.Trace, Observation{
		Stage: TraceStage, Step: 3, BeforeState: state("a"), AfterState: state("a"),
		BeforeRank: 1, AfterRank: 1, Decision: UpstreamNoChange, Reason: ReasonNoChange,
	})
	receipt, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionCycle || receipt.Outcome.DetectedPeriod != 2 || receipt.Outcome.ClaimState != ClaimRefuted {
		t.Fatalf("cycle was hidden by later no-change: %#v", receipt)
	}
}

func baseInput(maxSteps int, trace []Observation) Input {
	sourceDigest := digestBytes([]byte("source"))
	semanticDigest := digestBytes([]byte("semantic"))
	return Input{Schema: InputSchema, Repository: "kimjooyoon/meta-ontology-go", Subject: Consumer,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: ProofChoice,
		Stage: TraceStage, Source: SourceCausality{Path: SourcePath, SourceDigest: sourceDigest,
			SemanticDigest: semanticDigest, CaseID: "test", CaseProgramDigest: digestBytes([]byte("program"))},
		UpstreamDecision: UpstreamChanged, MaxSteps: maxSteps, Trace: trace,
		Interventions: []Intervention{
			{ID: "semantic-trace", Schema: InterventionSchema, Stage: InterventionStage, Step: 1,
				Reason: "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST", SourceBeforeDigest: sourceDigest,
				SourceAfterDigest: digestBytes([]byte("semantic-source")), SemanticBeforeDigest: semanticDigest,
				SemanticAfterDigest: digestBytes([]byte("semantic-after")), SourceChanged: true, SemanticChanged: true},
			{ID: "nonsemantic-comment", Schema: InterventionSchema, Stage: InterventionStage, Step: 2,
				Reason: "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST", SourceBeforeDigest: sourceDigest,
				SourceAfterDigest: digestBytes([]byte("comment-source")), SemanticBeforeDigest: semanticDigest,
				SemanticAfterDigest: semanticDigest, SourceChanged: true, SemanticChanged: false},
		},
	}
}

func state(value string) string { return stateDigest(value) }

func changed(step int, before, after string, beforeRank, afterRank int) Observation {
	return Observation{Stage: TraceStage, Step: step, BeforeState: state(before), AfterState: state(after),
		BeforeRank: beforeRank, AfterRank: afterRank, Decision: UpstreamChanged, Reason: ReasonStateChanged}
}

func fixedPointInput() Input {
	input := baseInput(4, []Observation{{Stage: TraceStage, Step: 1, BeforeState: state("a"), AfterState: state("a"),
		BeforeRank: 4, AfterRank: 4, Decision: UpstreamNoChange, Reason: ReasonNoChange}})
	input.UpstreamDecision = UpstreamNoChange
	return input
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

func unknownInput() Input {
	input := fixedPointInput()
	input.UpstreamDecision = "FUTURE_DECISION"
	return input
}
