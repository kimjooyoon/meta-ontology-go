package selfimprovementtermination

import (
	"fmt"
	"reflect"
)

type classification struct {
	decision, resolution, status, reason, finalState, claimState string
	period, stateCount, repeatedStates                           int
	terminationProven                                            bool
}

func Evaluate(input Input) (Receipt, error) {
	if err := ValidateInput(input); err != nil {
		return Receipt{}, err
	}
	class := effectiveClassification(input)
	receipt := Receipt{
		Schema: ReceiptSchema, Metaprogram: Metaprogram, Repository: input.Repository,
		Subject: input.Subject, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice, Stage: input.Stage,
		Status: class.status, Resolution: class.resolution, Decision: class.decision,
		Reason: class.reason, Source: input.Source, UpstreamDecision: input.UpstreamDecision,
		InputDigest: digestJSON(input), TraceDigest: digestJSON(input.Trace),
		Observations:  append([]Observation(nil), input.Trace...),
		Interventions: append([]Intervention(nil), input.Interventions...),
		Outcome: OutcomeSummary{ObservedSteps: len(input.Trace), MaxSteps: input.MaxSteps,
			StateCount: class.stateCount, RepeatedStates: class.repeatedStates,
			DetectedPeriod: class.period, FinalState: class.finalState,
			TerminationProven: class.terminationProven, ClaimState: class.claimState},
		Authority: Authority{ReadOnly: true},
	}
	receipt.ClaimTransitions = transitions(class, len(input.Trace))
	receipt.Indicators = indicators(input.Interventions)
	satisfied := countSatisfied(receipt.Indicators)
	receipt.Conformance = ConformanceSummary{Satisfied: satisfied, Total: IndicatorTotal, BasisPoints: basisPoints(satisfied, IndicatorTotal), Aggregation: ConformanceAggregation}
	receipt = seal(receipt)
	if err := ValidateReceipt(receipt, input); err != nil {
		return Receipt{}, fmt.Errorf("termination receipt: %w", err)
	}
	return receipt, nil
}

func ValidateInput(input Input) error {
	if input.Schema != InputSchema || input.Repository == "" || input.Subject != Consumer ||
		input.Producer != Producer || input.Consumer != Consumer || input.MetaOperation != MetaOperation ||
		input.ProofChoice != ProofChoice || input.Stage != TraceStage || input.MaxSteps < 1 ||
		input.MaxSteps > MaxTraceSteps || input.Source.Path != SourcePath || input.Source.CaseID == "" ||
		!validDigest(input.Source.SourceDigest) || !validDigest(input.Source.SemanticDigest) ||
		!validDigest(input.Source.CaseProgramDigest) || input.UpstreamDecision == "" {
		return invalid("identity, source causality, or budget is not bound")
	}
	if len(input.Trace) == 0 || len(input.Trace) > input.MaxSteps {
		return invalid("trace length is outside the declared budget")
	}
	for index, observation := range input.Trace {
		if observation.Stage != TraceStage || observation.Step != index+1 || observation.BeforeRank < 0 ||
			observation.AfterRank < 0 || !validDigest(observation.BeforeState) || !validDigest(observation.AfterState) {
			return invalid("step %d is malformed", index+1)
		}
		if index > 0 && input.Trace[index-1].AfterState != observation.BeforeState {
			return invalid("step %d breaks the state chain", index+1)
		}
		if observation.BeforeState == observation.AfterState {
			if observation.BeforeRank != observation.AfterRank || observation.Decision != UpstreamNoChange || observation.Reason != ReasonNoChange {
				return invalid("step %d claims an unbound no-change", index+1)
			}
		} else if observation.Decision != UpstreamChanged || observation.Reason != ReasonStateChanged {
			return invalid("step %d claims an unbound change", index+1)
		}
	}
	knownUpstream := input.UpstreamDecision == UpstreamNoChange || input.UpstreamDecision == UpstreamChanged
	if knownUpstream && input.UpstreamDecision != input.Trace[len(input.Trace)-1].Decision {
		return invalid("upstream decision disagrees with the final observed step")
	}
	if len(input.Interventions) != IndicatorTotal {
		return invalid("intervention denominator is not fixed at %d", IndicatorTotal)
	}
	for index, intervention := range input.Interventions {
		if intervention.Schema != InterventionSchema || intervention.Stage != InterventionStage ||
			intervention.Step != index+1 || !validDigest(intervention.SourceBeforeDigest) ||
			!validDigest(intervention.SourceAfterDigest) || !validDigest(intervention.SemanticBeforeDigest) ||
			!validDigest(intervention.SemanticAfterDigest) || !intervention.SourceChanged ||
			intervention.SourceBeforeDigest != input.Source.SourceDigest || intervention.SourceAfterDigest == intervention.SourceBeforeDigest ||
			intervention.SemanticBeforeDigest != input.Source.SemanticDigest {
			return invalid("intervention %d is not source-bound", index+1)
		}
		baselineClass := classify(input)
		expectedBaseline := interventionOutcome(input.Source.SourceDigest, input.Source.SemanticDigest, baselineClass, len(input.Trace))
		if !reflect.DeepEqual(intervention.Baseline, expectedBaseline) {
			return invalid("intervention %d baseline outcome is not source-bound", index+1)
		}
		if !validInterventionOutcome(intervention.Intervened) ||
			intervention.Intervened.SourceDigest != intervention.SourceAfterDigest ||
			intervention.Intervened.SemanticDigest != intervention.SemanticAfterDigest {
			return invalid("intervention %d intervened outcome is not source-bound", index+1)
		}
		switch intervention.ID {
		case "semantic-trace":
			if intervention.SemanticAfterDigest == intervention.SemanticBeforeDigest || !intervention.SemanticChanged || intervention.Reason != "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST" {
				return invalid("semantic intervention is not semantic")
			}
			if !canonicalOpenTransition(intervention.Intervened.ClaimTransitions) {
				return invalid("semantic intervention does not carry a canonical open claim transition")
			}
		case "nonsemantic-comment":
			if intervention.SemanticAfterDigest != intervention.SemanticBeforeDigest || intervention.SemanticChanged || intervention.Reason != "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST" {
				return invalid("comment intervention changed semantic meaning")
			}
			if !sameOutcome(intervention.Baseline, intervention.Intervened) {
				return invalid("comment intervention changed the subject outcome")
			}
		default:
			return invalid("intervention %d is unknown", index+1)
		}
	}
	return nil
}

func classify(input Input) classification {
	states := []string{input.Trace[0].BeforeState}
	for _, observation := range input.Trace {
		if observation.AfterState != states[len(states)-1] {
			states = append(states, observation.AfterState)
		}
	}
	cycleStart, period := repeatedState(states)
	repeatedStates := 0
	if cycleStart >= 0 {
		repeatedStates = 1
	}
	final := input.Trace[len(input.Trace)-1]
	if input.UpstreamDecision != UpstreamNoChange && input.UpstreamDecision != UpstreamChanged {
		return classification{DecisionFailClosed, ResolutionLower, ReceiptFailClosed, ReasonDecisionUnknown, final.AfterState, ClaimOpen, period, len(states), repeatedStates, false}
	}
	if cycleStart >= 0 {
		return classification{DecisionCycle, ResolutionExact, ReceiptBound, ReasonCycle, final.AfterState, ClaimRefuted, period, len(states), repeatedStates, false}
	}
	if input.UpstreamDecision == UpstreamNoChange && final.Decision == UpstreamNoChange {
		return classification{DecisionFixedPoint, ResolutionExact, ReceiptBound, ReasonNoChange, final.AfterState, ClaimDischarged, 0, len(states), repeatedStates, true}
	}
	diverging := len(input.Trace) == input.MaxSteps
	for _, observation := range input.Trace {
		diverging = diverging && observation.Decision == UpstreamChanged && observation.AfterRank > observation.BeforeRank
	}
	if diverging {
		return classification{DecisionDivergence, ResolutionLower, ReceiptBound, ReasonDivergence, final.AfterState, ClaimOpen, 0, len(states), repeatedStates, false}
	}
	return classification{DecisionInProgress, ResolutionLower, ReceiptBound, ReasonInProgress, final.AfterState, ClaimOpen, 0, len(states), repeatedStates, false}
}

func effectiveClassification(input Input) classification {
	class := classify(input)
	if len(input.Interventions) > 0 {
		semantic := input.Interventions[0]
		if semantic.SemanticChanged && semantic.Baseline.Decision == semantic.Intervened.Decision &&
			semantic.Baseline.Resolution == semantic.Intervened.Resolution {
			class = classification{DecisionFailClosed, ResolutionLower, ReceiptFailClosed, ReasonDigestOnly,
				class.finalState, ClaimOpen, class.period, class.stateCount, class.repeatedStates, false}
		}
	}
	return class
}

func repeatedState(states []string) (int, int) {
	for left := 0; left < len(states); left++ {
		for right := left + 1; right < len(states); right++ {
			if states[left] == states[right] {
				return left, right - left
			}
		}
	}
	return -1, 0
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
}

func transitions(class classification, finalStep int) []ClaimTransition {
	finalFrom, finalTo := ClaimOpen, ClaimOpen
	switch class.claimState {
	case ClaimDischarged:
		finalTo = ClaimDischarged
	case ClaimRefuted:
		finalTo = ClaimRefuted
	}
	return []ClaimTransition{
		{Stage: ClaimStage, Step: 0, From: ClaimOpen, To: ClaimOpen, Reason: "TRACE_BOUND"},
		{Stage: ClaimStage, Step: finalStep, From: finalFrom, To: finalTo, Reason: class.reason},
	}
}

func interventionOutcome(sourceDigest, semanticDigest string, class classification, finalStep int) InterventionOutcome {
	return InterventionOutcome{SourceDigest: sourceDigest, SemanticDigest: semanticDigest,
		Decision: class.decision, Resolution: class.resolution,
		ClaimTransitions: transitions(class, finalStep)}
}

func validInterventionOutcome(outcome InterventionOutcome) bool {
	return validDigest(outcome.SourceDigest) && validDigest(outcome.SemanticDigest) &&
		outcome.Decision != "" && outcome.Resolution != "" && len(outcome.ClaimTransitions) == 2
}

func canonicalOpenTransition(transitions []ClaimTransition) bool {
	return len(transitions) == 2 && transitions[0].Stage == ClaimStage && transitions[0].Step == 0 &&
		transitions[0].From == ClaimOpen && transitions[0].To == ClaimOpen && transitions[0].Reason == "TRACE_BOUND" &&
		transitions[1].Stage == ClaimStage && transitions[1].From == ClaimOpen && transitions[1].To == ClaimOpen &&
		transitions[1].Step > 0 && transitions[1].Reason != ""
}

func sameOutcome(left, right InterventionOutcome) bool {
	return left.Decision == right.Decision && left.Resolution == right.Resolution &&
		reflect.DeepEqual(left.ClaimTransitions, right.ClaimTransitions)
}

func ValidateReceipt(receipt Receipt, input Input) error {
	class := effectiveClassification(input)
	want := receiptForValidation(input, class)
	want.ReceiptDigest, want.ReplayDigest = receipt.ReceiptDigest, receipt.ReplayDigest
	if !reflect.DeepEqual(receipt, want) {
		return fmt.Errorf("receipt does not match source-bound outcome")
	}
	if !validDigest(receipt.ReceiptDigest) || receipt.ReceiptDigest != sealWithoutDigest(receipt).ReceiptDigest ||
		!validDigest(receipt.ReplayDigest) || receipt.ReplayDigest != replayDigest(receipt) {
		return fmt.Errorf("receipt digest is not sealed")
	}
	return nil
}

func receiptForValidation(input Input, class classification) Receipt {
	receipt := Receipt{
		Schema: ReceiptSchema, Metaprogram: Metaprogram, Repository: input.Repository,
		Subject: input.Subject, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice, Stage: input.Stage,
		Status: class.status, Resolution: class.resolution, Decision: class.decision,
		Reason: class.reason, Source: input.Source, UpstreamDecision: input.UpstreamDecision,
		InputDigest: digestJSON(input), TraceDigest: digestJSON(input.Trace),
		Observations: append([]Observation(nil), input.Trace...), Interventions: append([]Intervention(nil), input.Interventions...),
		ClaimTransitions: transitions(class, len(input.Trace)), Outcome: OutcomeSummary{
			ObservedSteps: len(input.Trace), MaxSteps: input.MaxSteps, StateCount: class.stateCount,
			RepeatedStates: class.repeatedStates, DetectedPeriod: class.period, FinalState: class.finalState,
			TerminationProven: class.terminationProven, ClaimState: class.claimState,
		}, Conformance: ConformanceSummary{Satisfied: IndicatorTotal, Total: IndicatorTotal, BasisPoints: 10000, Aggregation: ConformanceAggregation},
		Authority: Authority{ReadOnly: true}, Indicators: indicators(input.Interventions),
	}
	receipt.Conformance.Satisfied = countSatisfied(receipt.Indicators)
	receipt.Conformance.BasisPoints = basisPoints(receipt.Conformance.Satisfied, receipt.Conformance.Total)
	return receipt
}

func sealWithoutDigest(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	receipt.ReplayDigest = ""
	receipt.ReceiptDigest = digestJSON(receipt)
	return receipt
}

func replayDigest(receipt Receipt) string {
	return digestJSON(struct {
		InputDigest, TraceDigest, ReceiptDigest string
	}{receipt.InputDigest, receipt.TraceDigest, receipt.ReceiptDigest})
}
