package metainvocation

func claimsFor(decision, sourceDigest string, checks []PlannedCheck, unknowns []UnknownCause, failureReason string) []Claim {
	sourceClaim := Claim{
		ID: "source-program-binding", Statement: "the invoked entry is bound to the declared Gooo meta program",
		Status: ClaimDischarged, Stage: "SOURCE_BINDING", Step: "compile-meta-program", Reason: "SOURCE_BINDING_PROVED",
		Evidence: []string{sourceDigest}, DependsOn: []string{},
	}
	ruleClaim := Claim{
		ID: "rule-evidence-completeness", Statement: "every changed path has registered rule evidence",
		Status: ClaimDischarged, Stage: "RULE_SELECTION", Step: "classify-change", Reason: "RULE_EVIDENCE_COMPLETE",
		Evidence: []string{}, DependsOn: []string{"source-program-binding"},
	}
	for _, check := range checks {
		for _, reason := range check.Reasons {
			ruleClaim.Evidence = append(ruleClaim.Evidence, reason.ID)
		}
	}
	planClaim := Claim{
		ID: "ci-plan-decision", Statement: "the selected checks are authorized by complete rule evidence",
		Status: ClaimDischarged, Stage: "PLAN_AUTHORIZATION", Step: "require-rule-evidence", Reason: "PLAN_AUTHORIZED",
		Evidence: []string{}, DependsOn: []string{"rule-evidence-completeness"},
	}
	if decision == DecisionPass {
		for _, check := range checks {
			planClaim.Evidence = append(planClaim.Evidence, check.ID)
		}
	}
	if decision == DecisionClosed {
		ruleClaim.Status = ClaimRefuted
		ruleClaim.Stage = "INPUT_VALIDATION"
		ruleClaim.Step = "validate-change-set"
		ruleClaim.Reason = failureReason
		planClaim.Status = ClaimOpen
		planClaim.Reason = "DEPENDENCY_BLOCKED"
	}
	if decision == DecisionUnknown {
		ruleClaim.Status = ClaimOpen
		ruleClaim.Reason = "UNKNOWN_AT_RULE_SELECTION"
		for _, unknown := range unknowns {
			ruleClaim.Evidence = append(ruleClaim.Evidence, unknown.Stage+":"+unknown.Step+":"+unknown.Reason+":"+unknown.File)
		}
		planClaim.Status = ClaimOpen
		planClaim.Reason = "DEPENDENCY_BLOCKED"
	}
	return []Claim{sourceClaim, ruleClaim, planClaim}
}
