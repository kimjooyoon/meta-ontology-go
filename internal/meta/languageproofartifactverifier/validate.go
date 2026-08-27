package languageproofartifactverifier

import "fmt"

func Validate(report Report) error {
	canonicalContract := CanonicalContract()
	canonicalRecipe := CanonicalRecipe()
	if report.Schema != ReportSchema || report.Producer != ProducerID || report.Consumer != ConsumerID || !validHead(report.HeadSHA) ||
		report.ContractDigest != digestValue(canonicalContract) || report.RecipeDigest != digestValue(canonicalRecipe) || report.RecipeVersion != canonicalRecipe.Version ||
		!validDigest(report.IndependenceDigest) || report.AuthorityObservation != "UNKNOWN_GLOBAL_TRANSIENT_SCOPE" {
		return fmt.Errorf("proof-carrying report identity mismatch")
	}
	if report.ConformanceDecision != "PASS" || report.ConformanceResolution != "EXACT" || report.ConformanceReason != "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" ||
		report.SubjectArtifactDecision != "CARRIED" || report.SubjectArtifactResolution != "EVIDENCE_ATTACHED" || report.SubjectArtifactReason != "PROOF_CARRYING_ARTIFACT_EMITTED" ||
		report.ArtifactUseAuthority != "READ_ONLY_CONSUMPTION" || len(report.Cases) != CaseTotal || len(report.Indicators) != len(MetricIDs()) || len(report.Proofs) != 3 ||
		len(report.Transitions) != TransitionTotal || len(report.Interventions) != 2 || report.RepositoryWrites != report.WriteSet.RepositoryWrites || report.MutationAuthority || report.PromotionAuthority || report.SemanticAuthority {
		return fmt.Errorf("proof-carrying report shape mismatch")
	}
	if report.Checkout.HeadSHA != report.HeadSHA || report.Checkout.ActualHeadSHA != report.HeadSHA || !validHead(report.Checkout.HeadSHA) || !validDigest(report.Checkout.TreeDigest) ||
		!validDigest(report.Checkout.SourceDigest) || !validDigest(report.Checkout.OperationDigest) || !validDigest(report.Checkout.RecipeDigest) || !validDigest(report.Checkout.ContractDigest) ||
		(report.BundleDigest != "" && !validDigest(report.BundleDigest)) {
		return fmt.Errorf("proof-carrying checkout binding mismatch")
	}
	if err := validateWriteSet(report.WriteSet); err != nil {
		return err
	}
	want := Summary{CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, ValidArtifacts: 1, EvidenceKindsCarried: EvidenceTotal, ExactEvidenceLinks: EvidenceTotal,
		RecipeMatches: 1, PreservedTransitions: EvidenceTotal + 1, TransitionTotal: TransitionTotal, ClaimTemplates: ClaimTemplateTotal, ClaimInstances: CaseTotal * ClaimTemplateTotal,
		AcceptedTransitions: TransitionTotal, CaseDischargedClaims: 34, CaseOpenClaims: 13, CaseRefutedClaims: 13, FinalLedgerOpenClaims: ClaimTemplateTotal, FinalLedgerDischargedClaims: ClaimTemplateTotal,
		TamperedRejections: 1, CoherentTamperRejections: 1, MissingEvidenceRejections: 1, ByteOnlyDenials: 1, RecipeRejections: 1, RecipeOnlyRejections: 1,
		MissingAttachmentRejections: 1, WrongAttachmentRejections: 1, UnrelatedEvidenceRejections: 1, StaleHeadRejections: 1, UnauthorizedConsumerDenials: 1,
		SemanticInterventions: 1, NonsemanticInterventions: 1, ReadOnlyAuthorities: 1, ProducerDependencies: 0, ProducerImportNumerator: 0,
		ProducerImportDenominator: report.Summary.ProducerImportDenominator, CoreParserDependencies: report.Summary.CoreParserDependencies,
		NetRepositoryStateUnchanged: 1, BundleOnlyVerification: report.Summary.BundleOnlyVerification, ConsumerRechecks: report.Summary.ConsumerRechecks,
		GeneratedAuthority: 0, SemanticClaims: 0, RepositoryWrites: 0, MutationAuthorities: 0, PromotionAuthorities: 0, SemanticAuthorities: 0}
	if report.Summary != want || report.Summary.ProducerImportDenominator <= 0 || report.Summary.ProducerImportNumerator != 0 || report.Summary.CoreParserDependencies <= 0 {
		return fmt.Errorf("proof-carrying summary mismatch")
	}
	if err := validateOpenLedger(report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying prior ledger mismatch: %w", err)
	}
	if err := validateFinalLedger(report.Ledger, report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying final ledger mismatch: %w", err)
	}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Claims) != ClaimTemplateTotal || item.Coordinate.Stage == "" {
			return fmt.Errorf("proof-carrying case mismatch")
		}
		for _, claim := range item.Claims {
			if claim.ID == "" || claim.Proposition == "" || claim.StateDigest != claimStateDigest(claim) || claim.Provenance == "" ||
				(claim.Status == "DISCHARGED" && claim.Resolution != "EXACT") || (claim.Status == "OPEN" && claim.Resolution != "LOWER_RESOLUTION") ||
				(claim.Status == "REFUTED" && claim.Resolution != "INVARIANT_ONLY") {
				return fmt.Errorf("proof-carrying claim evaluation mismatch")
			}
		}
	}
	if err := validateTransitions(report.Transitions); err != nil {
		return err
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
		if !proof.Passed || !validDigest(proof.TargetDigest) || !validDigest(proof.ReceiptDigest) || len(proof.EvidenceDigests) == 0 {
			return fmt.Errorf("proof-carrying proof mismatch")
		}
		for _, digest := range proof.EvidenceDigests {
			if !validDigest(digest) {
				return fmt.Errorf("proof-carrying proof evidence mismatch")
			}
		}
	}
	if report.BundleDigest != "" {
		if report.ConsumerReceipt.Schema != ConsumerReceiptSchema || report.ConsumerReceipt.Version != 1 || report.ConsumerReceipt.Authority != "READ_ONLY_CONSUMPTION" ||
			!validDigest(report.ConsumerReceipt.TargetDigest) || !validDigest(report.ConsumerReceipt.OutputDigest) || !validDigest(report.ConsumerReceipt.AttestationDigest) ||
			report.ConsumerReceipt.AttestationDigest != attestationDigest(report) || report.ConsumerReceipt.Digest != consumerReceiptDigest(report.ConsumerReceipt) {
			return fmt.Errorf("proof-carrying consumer receipt mismatch")
		}
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("proof-carrying report digest mismatch")
	}
	return nil
}

func validateTransitions(transitions []ClaimTransition) error {
	previous := ""
	for index, transition := range transitions {
		if transition.ClaimID == "" || transition.Proposition == "" || transition.StateDigest == "" || !validDigest(transition.TargetDigest) && transition.TargetDigest != "READ_ONLY_CONSUMPTION" ||
			len(transition.EvidenceDigest) == 0 || transition.PreviousDigest != previous || transition.Digest != transitionDigest(transition) {
			return fmt.Errorf("proof-carrying transition chain mismatch")
		}
		for _, digest := range transition.EvidenceDigest {
			if !validDigest(digest) {
				return fmt.Errorf("proof-carrying transition evidence mismatch")
			}
		}
		if index < EvidenceTotal && (transition.Capability != "ARTIFACT_TRANSPORT" || transition.From != "CARRIED" || transition.To != "PRESERVED") {
			return fmt.Errorf("proof-carrying transport transition mismatch")
		}
		if index == EvidenceTotal && (transition.ClaimID != "consumer-authority" || transition.Capability != "ARTIFACT_USE" || transition.From != "NONE" || transition.To != "READ_ONLY_CONSUMPTION") {
			return fmt.Errorf("proof-carrying authority transition mismatch")
		}
		previous = transition.Digest
	}
	return nil
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}
