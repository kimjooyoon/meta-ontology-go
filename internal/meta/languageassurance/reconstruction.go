package languageassurance

type candidate struct {
	Decision   string
	Reason     string
	Resolution Resolution
}

func completeReconstruction(subjectSHA string, transaction Transaction, summary Summary, findings []Finding) (Summary, []Finding, candidate) {
	summary.UnresolvedIndicators = unresolved(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths, summary.ExactSnapshotBindingBPS)
	core := candidateFor(summary)
	expected := expectedRawReceipt(subjectSHA, transaction, summary, core)
	rawFindings := detectRawReconstructionMismatches(transaction.RawReconstructions, expected)
	rawBPS, rawPaths := observeRawReconstructions(transaction.RawReconstructions, len(rawFindings))
	summary.RawReconstructionSummary = RawReconstructionSummary{
		RawReconstructionsObserved: len(transaction.RawReconstructions), RawReconstructionsRequired: 1,
		RawReconstructionBPS: rawBPS, RawReconstructionMismatchPaths: rawPaths,
	}
	summary.UnresolvedIndicators = unresolved(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths, summary.ExactSnapshotBindingBPS, rawBPS)
	findings = append(findings, rawFindings...)
	sortFindings(findings)
	if rawBPS == nil {
		return summary, findings, candidate{CandidateFailClosed, ReasonEvidenceUnknown, ResolutionUnknown}
	}
	if *rawBPS < 10000 {
		return summary, findings, candidate{CandidateBlock, ReasonRawMismatch, ResolutionExact}
	}
	return summary, findings, core
}

func candidateFor(summary Summary) candidate {
	decision, reason, resolution := decide(summary)
	return candidate{decision, reason, resolution}
}

func expectedRawReceipt(subjectSHA string, transaction Transaction, summary Summary, candidate candidate) RawReconstructionReceipt {
	return RawReconstructionReceipt{
		Schema: RawReconstructionSchema, VerifierID: RawVerifierID, SubjectSHA: subjectSHA,
		DenominatorDigest: digest(Denominator()), RawTransactionDigest: rawTransactionDigest(transaction),
		Observation: rawObservation(summary, candidate),
	}
}

func rawObservation(summary Summary, candidate candidate) RawObservation {
	return RawObservation{
		EvidenceGroupsObserved: summary.EvidenceGroupsObserved, EvidenceGroupsTotal: summary.EvidenceGroupsTotal,
		SelfMintingPaths: summary.SelfMintingPaths, RoleConflictPaths: summary.RoleConflictPaths,
		UnknownLaunderingPaths: summary.UnknownLaunderingPaths, UnknownTopDecisions: summary.UnknownTopDecisions,
		SnapshotBindingsObserved: summary.SnapshotBindingsObserved, SnapshotBindingsRequired: summary.SnapshotBindingsRequired,
		ExactSnapshotBindingBPS: summary.ExactSnapshotBindingBPS, SnapshotMismatchPaths: summary.SnapshotMismatchPaths,
		CandidateDecision: candidate.Decision, CandidateReason: candidate.Reason, CandidateResolution: candidate.Resolution,
	}
}
