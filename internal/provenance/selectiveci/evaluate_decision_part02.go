package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
)

// Evaluate performs no selection or execution. It verifies only the supplied
// finite named paths and returns a deterministic receipt for that evidence.
func Evaluate(raw Input) Receipt {
	input, state := normalizeInput(raw)
	if state.class != issueNone {
		if state.class == issueFailClosed {
			return makeReceipt(input, FailClosed, FullSuite, state.code)
		}
		return makeReceipt(input, Unknown, FullSuite, state.code)
	}
	if len(input.SelectedCommandIDs) == 0 || len(input.ObligationIDs) == 0 {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if len(input.SelectedCommandIDs) != len(input.ObligationIDs) {
		return makeReceipt(input, FailClosed, FullSuite, CodeAmbiguous)
	}
	if len(input.Paths) < len(input.SelectedCommandIDs) {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if len(input.Paths) > len(input.SelectedCommandIDs) {
		return makeReceipt(input, FailClosed, FullSuite, CodeAmbiguous)
	}
	if issue := validateBindings(input); issue.class != issueNone {
		return makeReceipt(input, issueStatus(issue.class), FullSuite, issue.code)
	}
	if issue := validateChains(input); issue.class != issueNone {
		return makeReceipt(input, issueStatus(issue.class), FullSuite, issue.code)
	}
	closure := pathclosure.Evaluate(input.InferencePath, requirements(input))
	if closure.Status == pathclosure.UNKNOWN {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if closure.Status != pathclosure.PASS {
		return makeReceipt(input, FailClosed, FullSuite, CodeMalformed)
	}
	return makeReceipt(input, Verified, NoFallback, CodeVerified)
}
func issueStatus(class issueClass) DecisionStatus {
	if class == issueFailClosed {
		return FailClosed
	}
	return Unknown
}
