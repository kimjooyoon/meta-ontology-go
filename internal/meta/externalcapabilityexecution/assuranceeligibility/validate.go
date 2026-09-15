package assuranceeligibility

func unknownDecision(value evidence) bool {
	status, resolution, count := targetObligation(value.Assurance)
	if count != 1 || (status != "NOT_IMPLEMENTED" && status != "OPERATING") ||
		(resolution != "NONE" && resolution != ResolutionExact) {
		return true
	}
	parent := value.ParentReport.Decision
	capability, suite := value.CapabilityReport.Decision, value.CapabilitySuite.Decision
	return (parent != DecisionFailClosed && parent != "FIXED_POINT") ||
		(capability != "CAPABILITY_EXECUTABLE" && capability != DecisionFailClosed) ||
		(suite != "CAPABILITY_EXECUTABLE" && suite != DecisionFailClosed)
}

func validAssurance(value assuranceReport) bool {
	status, resolution, count := targetObligation(value)
	summary := value.Summary
	return value.Schema == "gooo/language-assurance-report/v1" && value.SubjectSHA == AssuranceSubject &&
		value.DenominatorID == "gooo/language-assurance-denominator/v1" &&
		value.DenominatorDigest == AssuranceDenominator && value.ReportDigest == AssuranceReport &&
		summary.DenominatorTotal == 12 && summary.Operating == 11 && summary.NotImplemented == 1 &&
		summary.ImplementationCoverageBPS == 9166 && summary.EvidenceGroupsObserved == 3 &&
		summary.EvidenceGroupsTotal == 3 && summary.SnapshotBindingsObserved == 3 &&
		summary.SnapshotBindingsRequired == 3 && summary.RawReconstructionsObserved == 1 &&
		summary.RawReconstructionsRequired == 1 && summary.UnknownTopDecisions == 0 &&
		summary.UnresolvedIndicators == 0 && summary.ViolatedGuardrails == 0 &&
		summary.RepositoryWrites == 0 && count == 1 && status == "NOT_IMPLEMENTED" && resolution == "NONE"
}

func targetObligation(value assuranceReport) (string, string, int) {
	var status, resolution string
	count := 0
	for _, obligation := range value.Obligations {
		if obligation.MetricID == MetricID {
			status, resolution, count = obligation.Status, obligation.Resolution, count+1
		}
	}
	return status, resolution, count
}

func validSubject(subject string, value evidence) bool {
	return value.CapabilityReport.SubjectSHA == subject &&
		value.CapabilityObservation.SubjectSHA == subject && value.CapabilitySuite.SubjectSHA == subject
}

func validReference(value evidence) bool {
	parent, capability := value.ParentObservation, value.CapabilityObservation.Reference
	return parent.GoVersion == GoVersion && parent.Reference.Available && parent.Reference.BindingExact &&
		parent.Reference.URL == ReferenceURL && parent.Reference.Commit == ReferenceCommit &&
		parent.Reference.Tree == ReferenceTree && capability.RepositoryURL == ReferenceURL &&
		capability.CommitSHA == ReferenceCommit && capability.TreeSHA == ReferenceTree &&
		capability.GoVersion == GoVersion
}
