package denominatorevolution

const (
	GuardrailForbiddenEstimate = "gooo.guardrail.denominator.forbidden-estimate.v1"
	GuardrailRepositoryWrites  = "gooo.guardrail.denominator.repository-writes.v1"
)

func summarize(cases []CaseResult, base Denominator, repositoryWrites int) Summary {
	summary := Summary{CasesTotal: len(cases), FixedDenominatorNumerator: len(base.Obligations), FixedDenominatorDenominator: DenominatorSize, LegalAdvanceDenominator: 1, UnauthorizedRejectionDenominator: 1, UnknownPredecessorDenominator: 1, AdditionReasonDenominator: 1, DeletionReasonDenominator: 1, Guardrails: makeGuardrails(0, repositoryWrites)}
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
	noEstimate := guardrailOrFail(summary, GuardrailForbiddenEstimate)
	readOnly := guardrailOrFail(summary, GuardrailRepositoryWrites)
	values := []struct {
		id, class, proof, operation      string
		stage, step, reason              string
		numerator, denominator, expected int
		guardrail                        *Guardrail
	}{
		{"gooo.metric.denominator.fixed-members.v1", "OUTCOME", "FOUNDATION", "bind-fixed-denominator", "FOUNDATION", "pin-versioned-member-set", "version and exact members are the measurement basis", summary.FixedDenominatorNumerator, summary.FixedDenominatorDenominator, DenominatorSize, nil},
		{"gooo.metric.denominator.legal-advance.v1", "OUTCOME", "COHERENCE", "accept-authorized-denominator-advance", "DECIDE", "apply-migration-receipt", "only a bound receipt accepts a successor", summary.LegalAdvanceNumerator, summary.LegalAdvanceDenominator, 1, nil},
		{"gooo.metric.denominator.unauthorized-rejection.v1", "GUARDRAIL", "REGRESSION", "reject-unreceipted-denominator-change", "DECIDE", "reject-missing-receipt", "a changed basis without a receipt is blocked", summary.UnauthorizedRejectionNumerator, summary.UnauthorizedRejectionDenominator, 1, nil},
		{"gooo.metric.denominator.unknown-predecessor.v1", "GUARDRAIL", "FOUNDATION", "fail-closed-unknown-predecessor", "RESOLVE", "lookup-predecessor", "unknown predecessor remains unknown", summary.UnknownPredecessorNumerator, summary.UnknownPredecessorDenominator, 1, nil},
		{"gooo.metric.denominator.addition-reasons.v1", "DRIVER", "COHERENCE", "classify-change-reason", "PROPOSE", "record-add-delete-reason", "every accepted addition has a reason", summary.AdditionReasonNumerator, summary.AdditionReasonDenominator, 1, nil},
		{"gooo.metric.denominator.deletion-reasons.v1", "DRIVER", "COHERENCE", "classify-change-reason", "PROPOSE", "record-add-delete-reason", "every accepted deletion has a reason", summary.DeletionReasonNumerator, summary.DeletionReasonDenominator, 1, nil},
		{"gooo.metric.denominator.no-estimate-claim.v1", "GUARDRAIL", "REGRESSION", "reject-aggregate-estimate", "GUARD", "reject-forbidden-claim", "no forbidden estimate claim is emitted", noEstimate.ConformanceNumerator, noEstimate.ConformanceDenominator, 1, noEstimate},
		{"gooo.metric.denominator.read-only.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-migration", "EFFECT", "observe-zero-writes", "the experiment has no mutation authority", readOnly.ConformanceNumerator, readOnly.ConformanceDenominator, 1, readOnly},
	}
	result := make([]Indicator, len(values))
	for index, value := range values {
		result[index] = Indicator{MetricID: value.id, Class: value.class, ProofChoice: value.proof, MetaOperation: value.operation, Coordinate: Coordinate{Stage: value.stage, Step: value.step, Reason: value.reason}, Numerator: value.numerator, Denominator: value.denominator, ExpectedNumerator: value.expected, Satisfied: value.numerator == value.expected && value.denominator > 0, Guardrail: value.guardrail}
	}
	return result
}

func makeGuardrails(forbiddenEstimateObserved, repositoryWrites int) []Guardrail {
	return []Guardrail{
		newGuardrail(GuardrailForbiddenEstimate, forbiddenEstimateObserved, 0),
		newGuardrail(GuardrailRepositoryWrites, repositoryWrites, 0),
	}
}

func newGuardrail(id string, observed, allowedMax int) Guardrail {
	conforms := observed <= allowedMax
	numerator := 0
	if conforms {
		numerator = 1
	}
	return Guardrail{ID: id, Direction: "AT_MOST", Observed: observed, AllowedMax: allowedMax, ConformanceNumerator: numerator, ConformanceDenominator: 1, Conforms: conforms}
}

func guardrailsConform(values []Guardrail) bool {
	expected := makeGuardrails(0, 0)
	if len(values) != len(expected) {
		return false
	}
	for index := range expected {
		if values[index] != expected[index] {
			return false
		}
	}
	return true
}

func guardrailFor(summary Summary, id string) *Guardrail {
	for index := range summary.Guardrails {
		if summary.Guardrails[index].ID == id {
			return &summary.Guardrails[index]
		}
	}
	return nil
}

func guardrailOrFail(summary Summary, id string) *Guardrail {
	if value := guardrailFor(summary, id); value != nil {
		return value
	}
	fallback := newGuardrail(id, 1, 0)
	return &fallback
}

func hasUnsatisfied(values []Indicator) bool {
	for _, value := range values {
		if !value.Satisfied {
			return true
		}
	}
	return false
}
