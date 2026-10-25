package generation

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func workspaceSplitInputFailures(indicators []sourcepolicy.Indicator, inputs map[string]workspacesplit.Request) []RegistrationInputFailure {
	var failures []RegistrationInputFailure
	used := map[string]bool{}
	for _, indicator := range indicators {
		if indicator.Operation != sourcepolicy.OperationSplitGooo || indicator.OperationInputDigest == "" {
			continue
		}
		key := indicator.OperationInputDigest
		used[key] = true
		request, exists := inputs[key]
		reason, state, class := "", "REFUTED", ""
		switch {
		case !exists:
			reason, state, class = "WORKSPACE_SPLIT_REQUEST_MISSING", "UNKNOWN", "DIRECT_MISSING"
		case request.Digest() != key:
			reason, state, class = "WORKSPACE_SPLIT_REQUEST_STALE", "UNKNOWN", "STALE"
		case !validWorkspaceSplitIndicator(indicator, request):
			reason = "WORKSPACE_SPLIT_REQUEST_MISMATCH"
		}
		if reason != "" {
			failures = append(failures, workspaceSplitInputFailure(indicatorID(indicator), key, state, reason, class))
		}
	}
	for key := range inputs {
		if !used[key] {
			failures = append(failures, workspaceSplitInputFailure("", key, "REFUTED", "UNREFERENCED_WORKSPACE_SPLIT_REQUEST", ""))
		}
	}
	sortOperationInputFailures(failures)
	return failures
}

func workspaceSplitInputFailure(id, digest, state, reason, class string) RegistrationInputFailure {
	return RegistrationInputFailure{IndicatorID: id, RequestDigest: digest, State: state,
		Stage: "OPERATION_INPUT", Step: "bind-workspace-split-request", Reason: reason,
		UnknownClass: class, NextOperation: "bind-exact-workspace-consumer-scope", BlockedBy: []string{}}
}
