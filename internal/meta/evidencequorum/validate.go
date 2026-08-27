package evidencequorum

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != Scope || !validHead(report.HeadSHA) ||
		report.SourcePath != CanonicalContract().SourcePath || !validDigest(report.SourceDigest) ||
		report.ContractDigest != digestJSON(CanonicalContract()) {
		return fmt.Errorf("evidence quorum identity mismatch")
	}
	want := Summary{CasesSatisfied: 4, CasesTotal: 4, ClaimsTotal: 4, DischargedClaims: 1,
		OpenClaims: 2, RefutedClaims: 1, RawEvidenceTotal: 12, IndependentGroupsTotal: 11,
		DuplicateEvidenceTotal: 1, ConflictCases: 1, QuorumSatisfiedCases: 1,
		LowerResolutionCases: 2, MinimumIndependentGroups: 3}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Reason != "EVIDENCE_QUORUM_CONTRACT_SATISFIED" || report.Summary != want ||
		len(report.Cases) != 4 || len(report.Indicators) != 8 || len(report.Proofs) != 4 ||
		report.RepositoryWrites != 0 || report.MutationAuthority || report.Summary.ConfidenceAggregated {
		return fmt.Errorf("evidence quorum report shape mismatch")
	}
	definitions := CanonicalContract().Cases
	for index, item := range report.Cases {
		definition := definitions[index]
		if item.ID != definition.ID || item.Status != "SATISFIED" ||
			item.ObservedDecision != definition.ExpectedDecision || item.ObservedResolution != definition.ExpectedResolution ||
			item.ObservedReason != definition.ExpectedReason || len(item.Claims) != 1 || len(item.Claims[0].Transitions) != 1 {
			return fmt.Errorf("evidence quorum case mismatch")
		}
		claim := item.Claims[0]
		if claim.ID != CanonicalContract().Claim.ID || claim.Producer == "" || claim.Consumer == "" ||
			claim.MetaOperation == "" || claim.ProofChoice == "" || claim.Status != definition.ExpectedStatus ||
			claim.Coordinate.Stage == "" || claim.Coordinate.Step == "" || claim.Coordinate.Reason == "" ||
			claim.Transitions[0].From != "OPEN" || claim.Transitions[0].To != claim.Status {
			return fmt.Errorf("evidence quorum claim transition mismatch")
		}
	}
	for _, item := range report.Indicators {
		if !item.Satisfied || item.Value != item.Target {
			return fmt.Errorf("evidence quorum indicator mismatch")
		}
	}
	for _, proof := range report.Proofs {
		if !proof.Passed || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("evidence quorum proof mismatch")
		}
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("evidence quorum digest mismatch")
	}
	return nil
}
