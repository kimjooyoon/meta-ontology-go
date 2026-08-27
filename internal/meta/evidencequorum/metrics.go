package evidencequorum

func indicators(report Report) []Indicator {
	return []Indicator{
		indicator("gooo.metric.meta.evidence-quorum.fixed-case-coverage.v1", "coverage", "FOUNDATION", "fix-denominator", report.Summary.CasesSatisfied, report.Summary.CasesTotal),
		indicator("gooo.metric.meta.evidence-quorum.independent-quorum.v1", "quorum", "FOUNDATION", "count-origin-groups", report.Summary.QuorumSatisfiedCases, 1),
		indicator("gooo.metric.meta.evidence-quorum.duplicate-not-independent.guardrail.v1", "guardrail", "REGRESSION", "collapse-origin-replicas", report.Summary.DuplicateEvidenceTotal, 1),
		indicator("gooo.metric.meta.evidence-quorum.conflict-fail-closed.guardrail.v1", "guardrail", "REGRESSION", "refute-conflicting-claim", report.Summary.ConflictCases, 1),
		indicator("gooo.metric.meta.evidence-quorum.insufficient-lower-resolution.guardrail.v1", "guardrail", "REGRESSION", "lower-insufficient-claim", report.Summary.LowerResolutionCases, 2),
		indicator("gooo.metric.meta.evidence-quorum.confidence-aggregation.guardrail.v1", "guardrail", "REGRESSION", "ignore-confidence-average", boolInt(!report.Summary.ConfidenceAggregated), 1),
		indicator("gooo.metric.meta.evidence-quorum.claim-transitions.v1", "trace", "COHERENCE", "record-claim-transitions", report.Summary.ClaimsTotal, report.Summary.CasesTotal),
		indicator("gooo.metric.meta.evidence-quorum.observer-writes.guardrail.v1", "guardrail", "REGRESSION", "preserve-read-only-evaluation", boolInt(report.Summary.RepositoryWrites == 0 && !report.Summary.MutationAuthority), 1),
	}
}

func indicator(metricID, class, choice, operation string, value, target int) Indicator {
	return Indicator{MetricID: metricID, Class: class, ProofChoice: choice, MetaOperation: operation,
		Value: value, Target: target, Satisfied: value == target}
}

func proofs(report Report) []Proof {
	if len(report.Cases) != 4 {
		return nil
	}
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "discharge-independent-quorum", EvidenceDigest: digestJSON(report.Cases[0]), Passed: report.Cases[0].Status == "SATISFIED"},
		{Choice: "REGRESSION", MetaOperation: "reject-same-origin-replica", EvidenceDigest: digestJSON(report.Cases[1]), Passed: report.Cases[1].Status == "SATISFIED"},
		{Choice: "REGRESSION", MetaOperation: "refute-conflicting-evidence", EvidenceDigest: digestJSON(report.Cases[2]), Passed: report.Cases[2].Status == "SATISFIED"},
		{Choice: "COHERENCE", MetaOperation: "lower-insufficient-evidence", EvidenceDigest: digestJSON(report.Cases[3]), Passed: report.Cases[3].Status == "SATISFIED"},
	}
}

func summarize(cases []CaseResult, contract Contract) Summary {
	summary := Summary{CasesTotal: contract.FixedCaseDenominator,
		MinimumIndependentGroups: contract.MinimumIndependentGroups}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		summary.ClaimsTotal += len(item.Claims)
		summary.RawEvidenceTotal += item.RawEvidence
		summary.IndependentGroupsTotal += item.IndependentGroups
		summary.DuplicateEvidenceTotal += item.DuplicateEvidence
		if item.ConflictGroups > 0 {
			summary.ConflictCases++
		}
		if item.ObservedDecision == DecisionPass {
			summary.QuorumSatisfiedCases++
		}
		if item.ObservedResolution == ResolutionLower {
			summary.LowerResolutionCases++
		}
		if len(item.Claims) == 1 {
			switch item.Claims[0].Status {
			case StatusDischarged:
				summary.DischargedClaims++
			case StatusOpen:
				summary.OpenClaims++
			case StatusRefuted:
				summary.RefutedClaims++
			}
		}
	}
	return summary
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
