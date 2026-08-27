package languageproofartifactverifier

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Producer != ProducerID || report.Consumer != ConsumerID ||
		!validHead(report.HeadSHA) || report.ContractDigest != digestValue(CanonicalContract()) ||
		report.RecipeDigest != digestValue(CanonicalRecipe()) || !validDigest(report.IndependenceDigest) {
		return fmt.Errorf("proof-carrying report identity mismatch")
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" || report.Reason != "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" ||
		!report.AuthorityGranted || len(report.Cases) != CaseTotal || len(report.Indicators) != len(MetricIDs()) ||
		len(report.Proofs) != 3 || len(report.Transitions) != TransitionTotal || report.RepositoryWrites != 0 || report.MutationAuthority {
		return fmt.Errorf("proof-carrying report shape mismatch")
	}
	want := Summary{CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, ValidArtifacts: 1, EvidenceKindsCarried: EvidenceTotal,
		ExactEvidenceLinks: EvidenceTotal, RecipeMatches: 1, PreservedTransitions: TransitionTotal, TransitionTotal: TransitionTotal,
		TamperedRejections: 1, MissingEvidenceRejections: 1, ByteOnlyDenials: 1, RecipeRejections: 1,
		ProducerDependencies: 0, GeneratedAuthority: 0, SemanticClaims: 0, RepositoryWrites: 0, MutationAuthorities: 0}
	if report.Summary != want {
		return fmt.Errorf("proof-carrying summary mismatch")
	}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Claims) != EvidenceTotal || item.Coordinate.Stage == "" {
			return fmt.Errorf("proof-carrying case mismatch")
		}
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied {
			return fmt.Errorf("proof-carrying indicator mismatch")
		}
	}
	for _, proof := range report.Proofs {
		if !proof.Passed || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("proof-carrying proof mismatch")
		}
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("proof-carrying report digest mismatch")
	}
	return nil
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}
