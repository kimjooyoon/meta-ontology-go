package selfimprovementtermination

import "fmt"

type classification struct {
	decision, reason, finalState string
	period, stateCount           int
	hasCycle, diverging          bool
}

func Evaluate(input Input) (Receipt, error) {
	if err := ValidateInput(input); err != nil {
		return Receipt{}, err
	}
	class := classify(input.Trace, input.MaxSteps)
	receipt := Receipt{
		Schema: ReceiptSchema, Metaprogram: Metaprogram, Repository: input.Repository,
		Subject: input.Subject, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice, Stage: input.Stage,
		Status: ReceiptBound, Resolution: ResolutionExact, Decision: class.decision,
		Reason: class.reason, InputDigest: digestJSON(input), TraceDigest: digestJSON(input.Trace),
		Observations: append([]Observation(nil), input.Trace...), Summary: Summary{
			ObservedSteps: len(input.Trace), MaxSteps: input.MaxSteps, StateCount: class.stateCount,
			DetectedPeriod: class.period, FinalState: class.finalState,
			TerminationProven: class.decision == DecisionFixedPoint,
		}, Authority: Authority{ReadOnly: true},
	}
	receipt.ClaimTransitions = transitions(class, len(input.Trace))
	receipt.Indicators = indicators(input, class)
	receipt.Summary.Total = len(receipt.Indicators)
	for _, indicator := range receipt.Indicators {
		if indicator.Satisfied {
			receipt.Summary.Satisfied++
		}
	}
	receipt.Summary.BasisPoints = basisPoints(receipt.Summary.Satisfied, receipt.Summary.Total)
	receipt = seal(receipt)
	if err := ValidateReceipt(receipt, input); err != nil {
		return Receipt{}, fmt.Errorf("termination receipt: %w", err)
	}
	return receipt, nil
}

func ValidateInput(input Input) error {
	if input.Schema != InputSchema || input.Repository == "" || input.Subject != Consumer ||
		input.Producer != Producer || input.Consumer != Consumer ||
		input.MetaOperation != MetaOperation || input.ProofChoice != ProofChoice ||
		input.Stage != TraceStage || input.MaxSteps < 1 || input.MaxSteps > MaxTraceSteps {
		return invalid("identity or budget is not bound")
	}
	if len(input.Trace) == 0 || len(input.Trace) > input.MaxSteps {
		return invalid("trace length is outside the declared budget")
	}
	for index, observation := range input.Trace {
		if observation.Stage != TraceStage || observation.Step != index+1 ||
			observation.BeforeRank < 0 || observation.AfterRank < 0 ||
			!validDigest(observation.BeforeState) || !validDigest(observation.AfterState) {
			return invalid("step %d is malformed", index+1)
		}
		if index > 0 && input.Trace[index-1].AfterState != observation.BeforeState {
			return invalid("step %d breaks the state chain", index+1)
		}
		if observation.BeforeState == observation.AfterState {
			if observation.BeforeRank != observation.AfterRank || observation.Decision != "NO_CHANGE" ||
				observation.Reason != "NO_CHANGE_FIXED_POINT_OBSERVED" {
				return invalid("step %d claims an unbound no-change", index+1)
			}
		} else if observation.Decision != "CHANGED" || observation.Reason != "METAPROGRAM_STATE_CHANGED" {
			return invalid("step %d claims an unbound change", index+1)
		}
	}
	return nil
}

func classify(trace []Observation, maxSteps int) classification {
	states := []string{trace[0].BeforeState}
	for _, observation := range trace {
		if observation.AfterState != states[len(states)-1] {
			states = append(states, observation.AfterState)
		}
	}
	cycleStart, period := repeatedState(states)
	final := trace[len(trace)-1]
	if cycleStart >= 0 {
		return classification{DecisionCycle, "REPEATED_STATE_CYCLE_OBSERVED", final.AfterState, period, len(states), true, false}
	}
	if final.Decision == "NO_CHANGE" {
		return classification{DecisionFixedPoint, "NO_CHANGE_FIXED_POINT_OBSERVED", final.AfterState, 0, len(states), false, false}
	}
	diverging := len(trace) == maxSteps
	for _, observation := range trace {
		diverging = diverging && observation.Decision == "CHANGED" && observation.AfterRank > observation.BeforeRank
	}
	if diverging {
		return classification{DecisionDivergence, "STRICTLY_GROWING_BOUNDARY_NO_FIXED_POINT", final.AfterState, 0, len(states), false, true}
	}
	return classification{DecisionInProgress, "TRACE_ENDED_BEFORE_TERMINATION", final.AfterState, 0, len(states), false, false}
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
	return []ClaimTransition{
		{Stage: ClaimStage, Step: 0, From: "UNPROVEN", To: "OBSERVED", Reason: "TRACE_BOUND"},
		{Stage: ClaimStage, Step: finalStep, From: "OBSERVED", To: class.decision, Reason: class.reason},
	}
}

func ValidateReceipt(receipt Receipt, input Input) error {
	if receipt.Schema != ReceiptSchema || receipt.Metaprogram != Metaprogram || receipt.Status != ReceiptBound ||
		receipt.Resolution != ResolutionExact || receipt.Repository != input.Repository ||
		receipt.Subject != input.Subject || receipt.InputDigest != digestJSON(input) ||
		receipt.TraceDigest != digestJSON(input.Trace) || receipt.Authority != (Authority{ReadOnly: true}) ||
		!validDigest(receipt.ReceiptDigest) || !validDigest(receipt.ReplayDigest) ||
		receipt.ReceiptDigest != sealWithoutDigest(receipt).ReceiptDigest ||
		receipt.ReplayDigest != digestJSON(struct {
			InputDigest, TraceDigest, ReceiptDigest string
		}{receipt.InputDigest, receipt.TraceDigest, receipt.ReceiptDigest}) {
		return fmt.Errorf("receipt identity or digest is not bound")
	}
	if len(receipt.Indicators) != IndicatorTotal || receipt.Summary.Total != IndicatorTotal ||
		receipt.Summary.Satisfied != IndicatorTotal || receipt.Summary.BasisPoints != 10000 {
		return fmt.Errorf("receipt denominator is not fixed at %d", IndicatorTotal)
	}
	for _, indicator := range receipt.Indicators {
		if !indicator.Satisfied || indicator.Producer != Producer || indicator.Consumer != Consumer ||
			indicator.MetaOperation != MetaOperation || indicator.ProofChoice != ProofChoice {
			return fmt.Errorf("indicator %s is not independently bound", indicator.ID)
		}
	}
	return nil
}

func sealWithoutDigest(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	receipt.ReplayDigest = ""
	receipt.ReceiptDigest = digestJSON(receipt)
	return receipt
}
