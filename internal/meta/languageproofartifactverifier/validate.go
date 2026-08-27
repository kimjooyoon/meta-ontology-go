package languageproofartifactverifier

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Producer != ProducerID || report.Consumer != ConsumerID ||
		!validHead(report.HeadSHA) || report.ContractDigest != digestValue(CanonicalContract()) ||
		report.RecipeDigest != digestValue(CanonicalRecipe()) || report.RecipeVersion != CanonicalRecipe().Version || !validDigest(report.IndependenceDigest) {
		return fmt.Errorf("proof-carrying report identity mismatch")
	}
	if report.ConformanceDecision != "PASS" || report.ConformanceResolution != "EXACT" || report.ConformanceReason != "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" ||
		report.SubjectArtifactDecision != "CARRIED" || report.SubjectArtifactResolution != "EVIDENCE_ATTACHED" ||
		report.SubjectArtifactReason != "PROOF_CARRYING_ARTIFACT_EMITTED" || report.ArtifactUseAuthority != "READ_ONLY_CONSUMPTION" ||
		len(report.Cases) != CaseTotal || len(report.Indicators) != len(MetricIDs()) || len(report.Proofs) != 3 ||
		len(report.Transitions) != TransitionTotal || len(report.Interventions) != 2 ||
		report.RepositoryWrites != report.WriteSet.RepositoryWrites || report.MutationAuthority || report.PromotionAuthority || report.SemanticAuthority {
		return fmt.Errorf("proof-carrying report shape mismatch")
	}
	if err := validateWriteSet(report.WriteSet); err != nil {
		return err
	}
	want := Summary{CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, ValidArtifacts: 1, EvidenceKindsCarried: EvidenceTotal,
		ExactEvidenceLinks: EvidenceTotal, RecipeMatches: 1, PreservedTransitions: EvidenceTotal, TransitionTotal: TransitionTotal,
		TamperedRejections: 1, CoherentTamperRejections: 1, MissingEvidenceRejections: 1, ByteOnlyDenials: 1, RecipeRejections: 1,
		LedgerDischargedClaims: 3, LedgerOpenClaims: 6, LedgerRefutedClaims: 9, SemanticInterventions: 1, NonsemanticInterventions: 1,
		ReadOnlyAuthorities: 1, ProducerDependencies: 0, ProducerImportNumerator: 0, ProducerImportDenominator: report.Summary.ProducerImportDenominator,
		CoreParserDependencies: report.Summary.CoreParserDependencies,
		GeneratedAuthority:     0, SemanticClaims: 0, RepositoryWrites: 0, MutationAuthorities: 0, PromotionAuthorities: 0, SemanticAuthorities: 0}
	if report.Summary != want {
		return fmt.Errorf("proof-carrying summary mismatch")
	}
	if report.Summary.ProducerImportDenominator <= 0 || report.Summary.ProducerImportNumerator != 0 {
		return fmt.Errorf("proof-carrying producer import ratio mismatch")
	}
	if report.Summary.CoreParserDependencies <= 0 {
		return fmt.Errorf("proof-carrying core parser dependency evidence missing")
	}
	if err := validateOpenLedger(report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying ledger mismatch")
	}
	if err := validateFinalLedger(report.Ledger, report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying ledger mismatch")
	}
	for index, entry := range report.Ledger.Entries[:EvidenceTotal] {
		if !reflect.DeepEqual(entry.EvidenceDigest, report.PriorLedger.Entries[index].EvidenceDigest) {
			return fmt.Errorf("proof-carrying ledger evidence mismatch")
		}
	}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Claims) != EvidenceTotal || item.Coordinate.Stage == "" {
			return fmt.Errorf("proof-carrying case mismatch")
		}
		for _, claim := range item.Claims {
			if claim.Provenance == "" || (claim.Status == "DISCHARGED" && claim.Resolution != "EXACT") ||
				(claim.Status == "OPEN" && claim.Resolution != "LOWER_RESOLUTION") ||
				(claim.Status == "REFUTED" && claim.Resolution != "INVARIANT_ONLY") {
				return fmt.Errorf("proof-carrying claim ledger mismatch")
			}
		}
	}
	for _, intervention := range report.Interventions {
		if intervention.Status != "SATISFIED" || !intervention.RawDigestChanged || !intervention.ConsumerDecisionPreserved {
			return fmt.Errorf("proof-carrying intervention mismatch")
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
