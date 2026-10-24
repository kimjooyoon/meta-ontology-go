package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func requiresSourcePrerequisite(operation sourcepolicy.Operation) bool {
	return operation == sourcepolicy.OperationSplitGo || operation == sourcepolicy.OperationExtractFunction
}

func validExtractionPrerequisite(observed CandidatePrerequisite) bool {
	if observed.Stage == "" || observed.Step == "" || observed.Reason == "" || observed.NextOperation == "" {
		return false
	}
	if observed.State == "READY" {
		return observed.Reason == "EXTRACTION_PLAN_READY" && observed.NextOperation == "execute-observed-extraction" &&
			observed.Stage == "prepare-extraction" && observed.Step == "stage-without-writing" &&
			observed.UnknownClass == "" && len(observed.BlockedBy) == 0 && len(observed.Diagnostics) == 0
	}
	if observed.Reason == "EXTRACTION_PLAN_READY" {
		return false
	}
	if observed.State == "BLOCKED" {
		return observed.UnknownClass == "KNOWN_CONTRADICTION"
	}
	return observed.State == "UNKNOWN" && validObservationFailureDecision(ObservationFailure{
		Decision: observed.State, UnknownClass: observed.UnknownClass})
}
