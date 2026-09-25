package languagesourcebindingpromotion

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != Scope || !validHead(report.HeadSHA) ||
		report.ContractDigest != digestJSON(CanonicalContract()) || !validDigest(report.PolicySourceDigest) ||
		!validDigest(report.PolicyArtifactDigest) || !validDigest(report.IndependenceDigest) {
		return fmt.Errorf("source binding promotion identity mismatch")
	}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Reason != "SOURCE_BINDING_PROMOTION_CONTRACT_SATISFIED" || len(report.Cases) != 5 ||
		len(report.Indicators) != 10 || len(report.Proofs) != 3 || report.RepositoryWrites != 0 || report.MutationAuthority {
		return fmt.Errorf("source binding promotion report shape mismatch")
	}
	want := Summary{CasesSatisfied: 5, CasesTotal: 5, ExactPromotions: 1, ExactClaims: 3,
		DirectUnknowns: 3, DependencyBlocked: 4, LinkRefutations: 1, PolicyReplays: 1}
	if report.Summary != want {
		return fmt.Errorf("source binding promotion summary mismatch")
	}
	claimIDs := []string{"structural-source-execution", "independent-source-binding", "source-binding-promotion"}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Claims) != 3 {
			return fmt.Errorf("source binding promotion case mismatch")
		}
		for index, claim := range item.Claims {
			if claim.ID != claimIDs[index] || claim.Status == "" || claim.Reason == "" || claim.Coordinate.Stage == "" {
				return fmt.Errorf("source binding promotion claim mismatch")
			}
		}
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied {
			return fmt.Errorf("source binding promotion indicator mismatch")
		}
	}
	for _, proof := range report.Proofs {
		if !proof.Passed || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("source binding promotion proof mismatch")
		}
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("source binding promotion digest mismatch")
	}
	return nil
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestJSON(report)
}
