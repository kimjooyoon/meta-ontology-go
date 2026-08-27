package denominatorevolution

func summarize(cases []CaseResult, base Denominator) Summary {
	summary := Summary{CasesTotal: len(cases), FixedDenominatorNumerator: len(base.Obligations), FixedDenominatorDenominator: DenominatorSize, LegalAdvanceDenominator: 1, UnauthorizedRejectionDenominator: 1, UnknownPredecessorDenominator: 1, AdditionReasonDenominator: 1, DeletionReasonDenominator: 1, ForbiddenEstimateDenominator: 1}
	for _, value := range cases {
		if value.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		switch value.ID {
		case "legal-advance":
			if value.ObservedDecision == "ADVANCE" {
				summary.LegalAdvanceNumerator = 1
			}
			if value.Receipt != nil && len(value.Receipt.Additions) == 1 {
				summary.AdditionReasonNumerator = 1
			}
			if value.Receipt != nil && len(value.Receipt.Deletions) == 1 {
				summary.DeletionReasonNumerator = 1
			}
		case "unauthorized-change":
			if value.ObservedDecision == "BLOCK" && value.ObservedReason == "MIGRATION_RECEIPT_MISSING" {
				summary.UnauthorizedRejectionNumerator = 1
			}
		case "unknown-predecessor":
			if value.ObservedDecision == "FAIL_CLOSED" && value.ObservedReason == "PREDECESSOR_DIGEST_UNKNOWN" {
				summary.UnknownPredecessorNumerator = 1
			}
		}
	}
	return summary
}

func makeIndicators(summary Summary) []Indicator {
	values := []struct {
		id, class, proof, operation      string
		stage, step, reason              string
		numerator, denominator, expected int
	}{
		{"gooo.metric.denominator.fixed-members.v1", "OUTCOME", "FOUNDATION", "bind-fixed-denominator", "FOUNDATION", "pin-versioned-member-set", "version and exact members are the measurement basis", summary.FixedDenominatorNumerator, summary.FixedDenominatorDenominator, DenominatorSize},
		{"gooo.metric.denominator.legal-advance.v1", "OUTCOME", "COHERENCE", "accept-authorized-denominator-advance", "DECIDE", "apply-migration-receipt", "only a bound receipt accepts a successor", summary.LegalAdvanceNumerator, summary.LegalAdvanceDenominator, 1},
		{"gooo.metric.denominator.unauthorized-rejection.v1", "GUARDRAIL", "REGRESSION", "reject-unreceipted-denominator-change", "DECIDE", "reject-missing-receipt", "a changed basis without a receipt is blocked", summary.UnauthorizedRejectionNumerator, summary.UnauthorizedRejectionDenominator, 1},
		{"gooo.metric.denominator.unknown-predecessor.v1", "GUARDRAIL", "FOUNDATION", "fail-closed-unknown-predecessor", "RESOLVE", "lookup-predecessor", "unknown predecessor remains unknown", summary.UnknownPredecessorNumerator, summary.UnknownPredecessorDenominator, 1},
		{"gooo.metric.denominator.addition-reasons.v1", "DRIVER", "COHERENCE", "classify-change-reason", "PROPOSE", "record-add-delete-reason", "every accepted addition has a reason", summary.AdditionReasonNumerator, summary.AdditionReasonDenominator, 1},
		{"gooo.metric.denominator.deletion-reasons.v1", "DRIVER", "COHERENCE", "classify-change-reason", "PROPOSE", "record-add-delete-reason", "every accepted deletion has a reason", summary.DeletionReasonNumerator, summary.DeletionReasonDenominator, 1},
		{"gooo.metric.denominator.no-estimate-claim.v1", "GUARDRAIL", "REGRESSION", "reject-aggregate-estimate", "GUARD", "reject-forbidden-claim", "no improvement rate or aggregate estimate is counted", summary.ForbiddenEstimateNumerator, summary.ForbiddenEstimateDenominator, 0},
		{"gooo.metric.denominator.read-only.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-migration", "EFFECT", "observe-zero-writes", "the experiment has no mutation authority", 0, 1, 0},
	}
	result := make([]Indicator, len(values))
	for index, value := range values {
		result[index] = Indicator{MetricID: value.id, Class: value.class, ProofChoice: value.proof, MetaOperation: value.operation, Coordinate: Coordinate{Stage: value.stage, Step: value.step, Reason: value.reason}, Numerator: value.numerator, Denominator: value.denominator, ExpectedNumerator: value.expected, Satisfied: value.numerator == value.expected && value.denominator > 0}
	}
	return result
}

func hasUnsatisfied(values []Indicator) bool {
	for _, value := range values {
		if !value.Satisfied {
			return true
		}
	}
	return false
}
