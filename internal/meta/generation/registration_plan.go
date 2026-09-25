package generation

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

const ReasonRegistrationInput Reason = "REGISTRATION_INPUT_UNBOUND"

func Build(baseSHA, headSHA string, report sourcepolicy.Report) Plan {
	return BuildWithRegistrationInputs(baseSHA, headSHA, report, nil)
}

// BuildWithRegistrationInputs does not lower the independent-operation floor.
// Input digests are already part of the source indicators and canonical input.
func BuildWithRegistrationInputs(baseSHA, headSHA string, report sourcepolicy.Report,
	inputs map[string]syntaxregistration.Request) Plan {
	plan := buildWithoutRegistrationInputs(baseSHA, headSHA, report)
	failures := registrationInputFailures(report.Indicators, inputs)
	if len(failures) != 0 {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonRegistrationInput
		plan.Selected = []Action{}
		plan.IndicatorDecisionLedger = nil
		plan.RegistrationInputFailures = failures
		for _, failure := range failures {
			if failure.State == "REFUTED" {
				plan.Decision = DecisionRejected
			}
			if failure.IndicatorID != "" {
				plan.UnknownIndicatorIDs = append(plan.UnknownIndicatorIDs, failure.IndicatorID)
			}
		}
		sort.Strings(plan.UnknownIndicatorIDs)
		return finish(plan)
	}
	changed := false
	for index := range plan.Selected {
		action := &plan.Selected[index]
		if action.Operation == sourcepolicy.OperationRegisterSyntax {
			request := inputs[action.SourceIndicator.OperationInputDigest]
			action.RegistrationRequest = cloneRegistrationRequest(&request)
			changed = true
		}
	}
	if changed {
		return attachPlanIndicatorDecisionLedger(plan, normalizeIndicators(report.Indicators))
	}
	return plan
}

func registrationInputFailures(indicators []sourcepolicy.Indicator, inputs map[string]syntaxregistration.Request) []RegistrationInputFailure {
	var failures []RegistrationInputFailure
	used := make(map[string]bool)
	for _, indicator := range indicators {
		if indicator.Operation != sourcepolicy.OperationRegisterSyntax {
			if indicator.OperationInputDigest != "" {
				failures = append(failures, registrationInputFailure(indicatorID(indicator),
					indicator.OperationInputDigest, "REFUTED", "OPERATION_INPUT_FORBIDDEN", ""))
			}
			continue
		}
		key := indicator.OperationInputDigest
		used[key] = true
		request, exists := inputs[key]
		switch {
		case !exists:
			failures = append(failures, registrationInputFailure(indicatorID(indicator), key,
				"UNKNOWN", "REGISTRATION_REQUEST_MISSING", "DIRECT_MISSING"))
		case syntaxregistration.RequestDigest(request) != key:
			failures = append(failures, registrationInputFailure(indicatorID(indicator), key,
				"UNKNOWN", "REGISTRATION_REQUEST_STALE", "STALE"))
		case !registrationRequestKnown(request) || request.Case.Path != indicator.Subject || !validRegistrationIndicator(indicator):
			failures = append(failures, registrationInputFailure(indicatorID(indicator), key,
				"REFUTED", "REGISTRATION_REQUEST_CONTRACT_MISMATCH", ""))
		}
	}
	for key := range inputs {
		if !used[key] {
			failures = append(failures, registrationInputFailure("", key,
				"REFUTED", "UNREFERENCED_REGISTRATION_REQUEST", ""))
		}
	}
	sort.Slice(failures, func(left, right int) bool {
		a, b := failures[left], failures[right]
		return a.IndicatorID+"\x00"+a.RequestDigest+"\x00"+a.Reason <
			b.IndicatorID+"\x00"+b.RequestDigest+"\x00"+b.Reason
	})
	return failures
}

func registrationInputFailure(indicatorID, requestDigest, state, reason, class string) RegistrationInputFailure {
	return RegistrationInputFailure{IndicatorID: indicatorID, RequestDigest: requestDigest,
		State: state, Stage: "OPERATION_INPUT", Step: "bind-registration-request",
		Reason: reason, UnknownClass: class, NextOperation: "observe-and-bind-exact-registration-input",
		BlockedBy: []string{}}
}
