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
	wantScope := "CURRENT_CHECKOUT_OBSERVATION"
	if report.BundleDigest != "" {
		wantScope = "BUNDLE_HISTORICAL_SUBJECT_BINDING"
	}
	if report.CheckoutBindingScope != wantScope {
		return fmt.Errorf("proof-carrying checkout binding scope mismatch")
	}
	if report.ConformanceDecision != "PASS" || report.ConformanceResolution != "EXACT" || report.ConformanceReason != "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" ||
		report.SubjectArtifactDecision != "CARRIED" || report.SubjectArtifactResolution != "EVIDENCE_ATTACHED" || report.SubjectArtifactReason != "PROOF_CARRYING_ARTIFACT_EMITTED" ||
		report.ArtifactUseAuthority != "READ_ONLY_CONSUMPTION" || !caseInventoryOK(report.Cases) || len(report.Indicators) != len(MetricIDs()) || len(report.Proofs) != 3 ||
		len(report.Transitions) != TransitionTotal || len(report.Interventions) != 2 || report.NetChangedPaths != report.WriteSet.NetChangedPaths || report.CapabilityMutationGranted || report.PromotionAuthority || report.SemanticAuthority {
		return fmt.Errorf("proof-carrying report shape mismatch")
	}
	if report.Checkout.HeadSHA != report.HeadSHA || !validHead(report.Checkout.HeadSHA) || !validDigest(report.Checkout.TreeDigest) ||
		!validDigest(report.Checkout.SourceDigest) || !validDigest(report.Checkout.OperationDigest) || !validDigest(report.Checkout.RecipeDigest) || !validDigest(report.Checkout.ContractDigest) ||
		(report.BundleDigest != "" && !validDigest(report.BundleDigest)) {
		return fmt.Errorf("proof-carrying checkout binding mismatch")
	}
	if err := validateWriteSet(report.WriteSet); err != nil {
		return err
	}
	bundleMetricValue := report.Summary.BundleOnlyVerification
	consumerMetricValue := report.Summary.ConsumerRechecks
	if bundleMetricValue != 0 && bundleMetricValue != 1 || consumerMetricValue != 0 && consumerMetricValue != 1 || report.BundleDigest != "" && (bundleMetricValue != 1 || consumerMetricValue != 1) {
		return fmt.Errorf("proof-carrying bundle metric scope mismatch")
	}
	want := Summary{CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, ValidArtifacts: 1, EvidenceKindsCarried: EvidenceTotal, ExactEvidenceLinks: EvidenceTotal,
		RecipeMatches: 1, PreservedTransitions: EvidenceTotal + 1, TransitionTotal: TransitionTotal, ClaimTemplates: ClaimTemplateTotal, ClaimInstances: CaseTotal * ClaimTemplateTotal,
		AcceptedTransitions: TransitionTotal, CaseDischargedClaims: 43, CaseOpenClaims: 20, CaseRefutedClaims: 17, FinalLedgerOpenClaims: ClaimTemplateTotal, FinalLedgerDischargedClaims: ClaimTemplateTotal,
		TamperedRejections: 1, CoherentTamperRejections: 1, CoherentClaimStructureRejections: 4, MissingEvidenceRejections: 1, ByteOnlyDenials: 1, RecipeRejections: 1, RecipeOnlyRejections: 1,
		MissingAttachmentRejections: 1, WrongAttachmentRejections: 1, UnrelatedEvidenceRejections: 1, StaleHeadRejections: 1, UnauthorizedConsumerDenials: 1,
		SemanticInterventions: 1, NonsemanticInterventions: 1, ReadOnlyAuthorities: 1, ProducerDependencies: 0, ProducerImportNumerator: 0,
		ProducerImportDenominator: report.Summary.ProducerImportDenominator, CoreParserDependencies: CoreParserDependencyInventoryTotal,
		NetRepositoryStateUnchanged: 1, UnknownAuthorityObservations: 1, BundleOnlyVerification: bundleMetricValue, ConsumerRechecks: consumerMetricValue,
		GeneratedAuthority: 0, SemanticClaims: 0, NetChangedPaths: 0, MutationAuthorities: 0, PromotionAuthorities: 0, SemanticAuthorities: 0}
	if report.Summary != want || report.Summary.ProducerImportDenominator <= 0 || report.Summary.ProducerImportNumerator != 0 || report.Summary.CoreParserDependencies != CoreParserDependencyInventoryTotal {
		return fmt.Errorf("proof-carrying summary mismatch")
	}
	if !validDigest(report.UnauthorizedConsumerTargetDigest) || !validDigest(report.UnauthorizedConsumerOutputDigest) {
		return fmt.Errorf("proof-carrying unauthorized consumer observation mismatch")
	}
	if err := validateOpenLedger(report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying prior ledger mismatch: %w", err)
	}
	if err := validateFinalLedger(report.Ledger, report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying final ledger mismatch: %w", err)
	}
	claimIDs := make([]string, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		claimIDs = append(claimIDs, spec.ID)
	}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Claims) != ClaimTemplateTotal || item.Coordinate.Stage == "" || item.ObservedDecision != item.ExpectedDecision || item.ObservedResolution != item.ExpectedResolution || item.ObservedReason != item.ExpectedReason {
			return fmt.Errorf("proof-carrying case mismatch")
		}
		seen := map[string]bool{}
		for index, claim := range item.Claims {
			if index >= len(claimIDs) || claim.ID != claimIDs[index] || seen[claim.ID] || claim.Provenance == "" || claim.StateDigest != claimStateDigest(claim) ||
				(claim.Status == "DISCHARGED" && claim.Resolution != "EXACT") || (claim.Status == "OPEN" && claim.Resolution != "LOWER_RESOLUTION") ||
				(claim.Status == "REFUTED" && claim.Resolution != "INVARIANT_ONLY") {
				return fmt.Errorf("proof-carrying claim evaluation mismatch")
			}
			seen[claim.ID] = true
		}
		if item.ID == "valid-proof-carrying-artifact" && !validCaseClaims(item, report) {
			return fmt.Errorf("proof-carrying valid claim binding mismatch")
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
			allowedOpen := report.BundleDigest == "" && (indicator.MetricID == "gooo.metric.language.proof-carrying-artifact-bundle-only.v3" || indicator.MetricID == "gooo.metric.language.proof-carrying-artifact-consumer-recheck.v3") && indicator.Value == 0 && indicator.Target == 1
			if !allowedOpen {
				return fmt.Errorf("proof-carrying indicator mismatch")
			}
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

func validCaseClaims(item CaseResult, report Report) bool {
	if len(item.Claims) != ClaimTemplateTotal || item.SourceDigest == "" || item.OperationDigest == "" {
		return false
	}
	for index, spec := range claimSpecs() {
		claim := item.Claims[index]
		target := spec.TargetDigest
		switch spec.ID {
		case "source-bytes-bound":
			target = item.SourceDigest
		case "operation-receipt-bound":
			target = item.OperationDigest
		case "recipe-match":
			target = report.RecipeDigest
		}
		if claim.ID != spec.ID || claim.Proposition != spec.Proposition || claim.TargetDigest != target || !equalStrings(claim.Dependencies, spec.Dependencies) ||
			claim.ProofChoice != spec.ProofChoice || claim.MetaOperation != spec.MetaOperation || claim.Status != "DISCHARGED" || claim.Resolution != "EXACT" || len(claim.EvidenceDigests) == 0 || claim.EvidenceDigest != claim.EvidenceDigests[0] {
			return false
		}
	}
	return true
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validateTransitions(transitions []ClaimTransition) error {
	previous := ""
	claimIDs := []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match", "consumer-authority"}
	for index, transition := range transitions {
		if transition.ClaimID != claimIDs[index] || transition.ClaimID == "" || transition.Proposition == "" || transition.StateDigest == "" || !validDigest(transition.PriorStateDigest) || (!validDigest(transition.TargetDigest) && transition.TargetDigest != "READ_ONLY_CONSUMPTION") ||
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
