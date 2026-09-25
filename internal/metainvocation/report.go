package metainvocation

func buildReport(program Program, entry, caseID, inputDigest, decision, resolution string, checks []PlannedCheck, unknowns []UnknownCause, failureReason string) Report {
	if checks == nil {
		checks = []PlannedCheck{}
	}
	if unknowns == nil {
		unknowns = []UnknownCause{}
	}
	plan := sealPlan(CheckPlan{Schema: PlanSchema, CaseID: caseID, InputDigest: inputDigest, Checks: checks})
	evidenceDigests := []string{program.SourceDigest}
	for _, check := range checks {
		for _, reason := range check.Reasons {
			evidenceDigests = append(evidenceDigests, digest(reason))
		}
	}
	receipt := sealReceipt(VerificationReceipt{
		Schema: ReceiptSchema, SubjectDigest: plan.Digest, Decision: decision, Resolution: resolution,
		EvidenceDigests: evidenceDigests, Unknowns: unknowns,
	})
	report := Report{
		Schema: ReportSchema, Decision: decision, Resolution: resolution, CaseID: caseID, Entry: entry,
		SourceDigest: program.SourceDigest, InputDigest: inputDigest, Plan: plan, Receipt: receipt,
		Unknowns: unknowns, Claims: claimsFor(decision, program.SourceDigest, checks, unknowns, failureReason),
		Effects: Effects{}, NotClaimed: []string{
			"check-execution", "full-language-semantic-correctness", "general-build-planning",
			"comparative-performance", "production-or-external-effects",
		},
	}
	return sealReport(report)
}
