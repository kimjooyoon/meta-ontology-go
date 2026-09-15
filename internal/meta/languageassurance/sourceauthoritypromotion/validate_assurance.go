package sourceauthoritypromotion

func validateAssurance(doc assuranceDocument) (Baseline, bool, string) {
	baseline := Baseline{DenominatorID: doc.DenominatorID, DenominatorDigest: doc.DenominatorDigest,
		Total: doc.Summary.DenominatorTotal, Operating: doc.Summary.Operating,
		NotImplemented: doc.Summary.NotImplemented, CoverageBPS: doc.Summary.ImplementationCoverageBPS}
	if doc.Schema != AssuranceSchema || doc.DenominatorID != AssuranceDenominator || doc.DenominatorDigest != AssuranceDigest {
		return baseline, false, ReasonAssuranceBoundary
	}
	if doc.AssuranceDecision != "PARTIAL" || doc.CandidateDecision != "ALLOW_LIMITED" {
		return baseline, false, ReasonAssuranceBoundary
	}
	if len(doc.Denominator) != 12 || len(doc.Obligations) != 12 {
		return baseline, false, ReasonAssuranceBoundary
	}
	definitions := 0
	for _, definition := range doc.Denominator {
		if definition.MetricID != SourceMetric {
			continue
		}
		definitions++
		if definition.Priority != "P1" || definition.Class != "DRIVER" || definition.ProofChoice != "FOUNDATION" || definition.RequiredMetaOperation != SourceOperation {
			return baseline, false, ReasonAssuranceBoundary
		}
	}
	operating, missing, sourceStates := 0, 0, 0
	for _, obligation := range doc.Obligations {
		switch {
		case obligation.Status == "OPERATING" && obligation.Resolution == ResolutionExact:
			operating++
		case obligation.Status == "NOT_IMPLEMENTED" && obligation.Resolution == "NONE":
			missing++
		default:
			return baseline, false, ReasonAssuranceBoundary
		}
		if obligation.MetricID == SourceMetric {
			sourceStates++
			if obligation.Status != "NOT_IMPLEMENTED" || obligation.Resolution != "NONE" {
				return baseline, false, ReasonBaselineState
			}
		}
	}
	if definitions != 1 || sourceStates != 1 || operating != 6 || missing != 6 {
		return baseline, false, ReasonAssuranceBoundary
	}
	s := doc.Summary
	if s.DenominatorTotal != 12 || s.Operating != 6 || s.NotImplemented != 6 || s.ImplementationCoverageBPS != 5000 || s.UnknownTopDecisions != 0 || s.UnresolvedIndicators != 0 || s.ViolatedGuardrails != 0 || s.RepositoryWrites != 0 {
		return baseline, false, ReasonAssuranceBoundary
	}
	return baseline, true, ""
}
