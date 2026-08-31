package languagedebug

func Observe(data []byte, breakpoint string) Receipt {
	execution, ok := decodeExecution(data)
	if !ok || breakpoint == "" {
		return seal(Receipt{
			Schema: "gooo/language-debug-receipt/v1", Decision: DecisionFailClosed,
			Reason: "DEBUG_EXECUTION_UNKNOWN", Resolution: ResolutionLower,
			State: StateRejected, Breakpoint: breakpoint, Trace: []Event{},
			NonClaims: CanonicalNonClaims(),
		})
	}
	receipt := fromExecution(execution, breakpoint)
	for index, event := range execution.Events {
		if event.Kind != breakpoint {
			continue
		}
		current := event
		receipt.Decision = DecisionPass
		receipt.Reason = "DEBUG_BREAKPOINT_REACHED"
		receipt.State = StatePaused
		receipt.CurrentEvent = &current
		receipt.Trace = append([]Event(nil), execution.Events[:index+1]...)
		receipt.RemainingEvents = len(execution.Events) - index - 1
		return seal(receipt)
	}
	receipt.Decision = DecisionFailClosed
	receipt.Reason = "DEBUG_BREAKPOINT_NOT_REACHED"
	receipt.State = StateRejected
	receipt.Trace = append([]Event(nil), execution.Events...)
	return seal(receipt)
}

func fromExecution(execution executionReceipt, breakpoint string) Receipt {
	return Receipt{
		Schema: "gooo/language-debug-receipt/v1", Resolution: ResolutionExact,
		Filename: execution.Filename, SourceDigest: execution.SourceDigest,
		SemanticDigest: execution.SemanticDigest, ExecutionDigest: execution.Digest,
		Entry: execution.Entry, Breakpoint: breakpoint, Diagnostics: execution.Diagnostics,
		Effects: execution.Effects, NonClaims: CanonicalNonClaims(), Trace: []Event{},
	}
}
