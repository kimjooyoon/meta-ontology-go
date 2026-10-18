package bodycodegen

import "fmt"

// PureCallBudgetError locates the first source call site that crosses the
// conservative static budget. It describes a visited prefix, not a complete
// closure or a measured runtime count. Ordinals include primitive call sites,
// matching PureCallEdge. Loop multiplicities may saturate above the limit.
type PureCallBudgetError struct {
	Caller             string `json:"caller"`
	CallerID           string `json:"caller_id"`
	Callee             string `json:"callee"`
	CalleeID           string `json:"callee_id"`
	CallOrdinal        int    `json:"call_ordinal"`
	Limit              int    `json:"limit"`
	PrefixBefore       int    `json:"prefix_before"`
	Multiplicity       int    `json:"multiplicity"`
	MultiplicityCapped bool   `json:"multiplicity_capped"`
	CalleeBound        int    `json:"callee_bound"`
	PrefixAfter        int    `json:"prefix_after"`
}

func (e *PureCallBudgetError) Error() string {
	return fmt.Sprintf("pure activity closure exceeds %d calls per invocation: %s (%s), call %d to %s (%s): "+
		"prefix %d + multiplicity %d * (1 + callee bound %d) = %d; multiplicity capped: %t",
		e.Limit, e.Caller, e.CallerID, e.CallOrdinal, e.Callee, e.CalleeID,
		e.PrefixBefore, e.Multiplicity, e.CalleeBound, e.PrefixAfter, e.MultiplicityCapped)
}

func (r *pureCallResolver) budgetError(caller, callee string, ordinal, before, multiplicity, calleeBound int) error {
	return &PureCallBudgetError{
		Caller: caller, CallerID: r.ids[caller], Callee: callee, CalleeID: r.ids[callee],
		CallOrdinal: ordinal, Limit: pureCallLimit, PrefixBefore: before,
		Multiplicity: multiplicity, MultiplicityCapped: multiplicity > pureCallLimit,
		CalleeBound: calleeBound, PrefixAfter: before + multiplicity*(1+calleeBound),
	}
}
