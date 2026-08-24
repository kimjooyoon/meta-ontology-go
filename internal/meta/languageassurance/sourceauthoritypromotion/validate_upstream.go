package sourceauthoritypromotion

func validateUpstream(doc upstreamDocument) (Evidence, bool) {
	evidence := Evidence{DenominatorID: doc.DenominatorID, DenominatorDigest: doc.DenominatorDigest,
		CasesPassed: doc.Summary.CasesPassed, CasesTotal: doc.Summary.CasesTotal, CoverageBPS: doc.Summary.CoverageBPS}
	if doc.Schema != UpstreamSchema || doc.DenominatorID != UpstreamDenominator || doc.DenominatorDigest != UpstreamDigest || doc.Decision != "PASS" || doc.Resolution != ResolutionExact {
		return evidence, false
	}
	if doc.RepositoryWrites != 0 || doc.PromotionCreditBPS != 0 || len(doc.Cases) != 3 {
		return evidence, false
	}
	s := doc.Summary
	if s.CasesTotal != 3 || s.CasesPassed != 3 || s.ExactAllow != 1 || s.FailClosed != 2 || s.CoverageBPS != 10000 {
		return evidence, false
	}
	exact := findCase(doc.Cases, "exact")
	digest := findCase(doc.Cases, "digest-mismatch")
	authority := findCase(doc.Cases, "authority-mismatch")
	if !caseMatches(exact, "SATISFIED", ResolutionExact, "ALLOW", "SOURCE_SNAPSHOT_EXACT", true) {
		return evidence, false
	}
	if !caseMatches(digest, "UNKNOWN", ResolutionInvariantOnly, DecisionBlock, "SOURCE_DIGEST_MISMATCH", true) {
		return evidence, false
	}
	if !caseMatches(authority, "UNKNOWN", ResolutionInvariantOnly, DecisionBlock, "AUTHORITY_SCOPE_MISMATCH", false) {
		return evidence, false
	}
	if !validSnapshot(exact.Receipt.Snapshot) || !validSnapshot(digest.Receipt.Snapshot) || !validIndicatorSplit(exact.Receipt.Indicators) {
		return evidence, false
	}
	snapshot := exact.Receipt.Snapshot
	evidence.Repository, evidence.Revision, evidence.SnapshotDigest = snapshot.Authority.Repository, snapshot.Authority.Revision, snapshot.Digest
	return evidence, true
}

func findCase(cases []upstreamCase, id string) *upstreamCase {
	for index := range cases {
		if cases[index].ID == id {
			return &cases[index]
		}
	}
	return nil
}

func caseMatches(item *upstreamCase, observation, resolution, enforcement, reason string, snapshot bool) bool {
	if item == nil || !item.Passed || item.ExpectedObservation != observation || item.ExpectedResolution != resolution || item.ExpectedEnforcement != enforcement || item.ExpectedReason != reason {
		return false
	}
	r := item.Receipt
	return r.Observation == observation && r.Resolution == resolution && r.Enforcement == enforcement && r.Reason == reason && r.RepositoryWrites == 0 && r.PromotionCreditBPS == 0 && (r.Snapshot != nil) == snapshot
}

func validSnapshot(snapshot *upstreamSnapshot) bool {
	return snapshot != nil && snapshot.Authority.Repository == "cosmos72/gomacro" && snapshot.Authority.Revision == "cf0d4bf32da393dbda97e3572f216731013ffa55" && snapshot.Authority.Path == "README.md" && snapshot.Selection.StartLine == 1 && snapshot.Selection.EndLine == 1 && snapshot.Bytes == 77 && snapshot.Digest == "sha256:29362aa311de0f24c66f41cc65a8b6ffd996baf37e048b5a72db63172aae5bf2"
}

func validIndicatorSplit(indicators []upstreamIndicator) bool {
	counts := map[string]int{}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
		counts[indicator.Class]++
		counts[indicator.ProofChoice]++
	}
	return len(indicators) == 6 && counts["OUTCOME"] == 1 && counts["DRIVER"] == 2 && counts["GUARDRAIL"] == 3 && counts["FOUNDATION"] == 3 && counts["COHERENCE"] == 2 && counts["REGRESSION"] == 1
}
