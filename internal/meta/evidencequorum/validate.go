package evidencequorum

import "fmt"

func Validate(report Report) error {
	canonical := CanonicalContract()
	if report.Schema != ReportSchema || report.Scope != Scope || !validHead(report.HeadSHA) ||
		report.SourcePath != canonical.SourcePath || report.SourceEntry != canonical.SourceEntry ||
		!validDigest(report.SourceDigest) || !validDigest(report.ProducerReceiptDigest) ||
		!validDigest(report.UnknownProducerReceiptDigest) || report.ContractDigest != digestJSON(canonical) {
		return fmt.Errorf("evidence quorum identity mismatch: source_path=%q source_entry=%q source_digest=%q producer_receipt_digest=%q unknown_producer_receipt_digest=%q contract_digest=%q expected_contract_digest=%q", report.SourcePath, report.SourceEntry, report.SourceDigest, report.ProducerReceiptDigest, report.UnknownProducerReceiptDigest, report.ContractDigest, digestJSON(canonical))
	}
	want := Summary{CasesSatisfied: 5, CasesTotal: 5, ClaimsTotal: 5, DischargedClaims: 1,
		OpenClaims: 2, RefutedClaims: 1, UnknownClaims: 1, RawEvidenceTotal: 12, IndependentGroupsTotal: 11,
		DuplicateEvidenceTotal: 1, ConflictCases: 1, QuorumSatisfiedCases: 1,
		LowerResolutionCases: 3, MinimumIndependentGroups: 3}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Reason != "EVIDENCE_QUORUM_CONTRACT_SATISFIED" || report.Summary != want ||
		len(report.Cases) != 5 || len(report.Indicators) != 9 || len(report.Proofs) != 5 ||
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
		if definition.ProducerDecision == DecisionUnknown &&
			(claim.Coordinate.Stage != "UNKNOWN" || claim.Coordinate.Step != "UNKNOWN" || claim.Coordinate.Reason != "QUORUM_EVIDENCE_UNKNOWN") {
			return fmt.Errorf("evidence quorum unknown coordinate mismatch")
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
