package selfimprovementtermination

import "fmt"

func indicators(input Input, class classification) []Indicator {
	decision := class.decision
	fixed := decision == DecisionFixedPoint
	cycle := decision == DecisionCycle
	divergence := decision == DecisionDivergence
	inProgress := decision == DecisionInProgress
	finalNoChange := input.Trace[len(input.Trace)-1].Decision == "NO_CHANGE"
	return []Indicator{
		indicator("gooo.termination.input-schema.v1", "INPUT_SCHEMA_BOUND", true),
		indicator("gooo.termination.identity-bound.v1", "PRODUCER_CONSUMER_BOUND", true),
		indicator("gooo.termination.bounded-trace.v1", "TRACE_WITHIN_FIXED_BUDGET", len(input.Trace) <= input.MaxSteps),
		indicator("gooo.termination.state-chain.v1", "CONTIGUOUS_STATE_CHAIN", true),
		indicator("gooo.termination.no-change-branch.v1", "NO_CHANGE_BRANCH_EXACT", fixed == (finalNoChange && !class.hasCycle)),
		indicator("gooo.termination.cycle-branch.v1", "CYCLE_BRANCH_EXACT", cycle == class.hasCycle),
		indicator("gooo.termination.divergence-branch.v1", "DIVERGENCE_IS_ONLY_POSSIBLE", divergence == class.diverging),
		indicator("gooo.termination.progress-branch.v1", "IN_PROGRESS_BRANCH_EXACT", inProgress == (!finalNoChange && !class.hasCycle && !class.diverging)),
		indicator("gooo.termination.claim-transition.v1", "CLAIM_TRANSITION_BOUND", true),
		indicator("gooo.termination.read-only-authority.v1", "READ_ONLY_NO_PROMOTION", true),
	}
}

func indicator(id, reason string, satisfied bool) Indicator {
	return Indicator{
		ID: id, Route: "TERMINATION", Producer: Producer, Consumer: Consumer,
		MetaOperation: MetaOperation, ProofChoice: ProofChoice, Stage: ClaimStage,
		Step: 0, Reason: reason, Value: fmt.Sprint(satisfied), Limit: "true", Satisfied: satisfied,
	}
}
