package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	canonicalContract := CanonicalContract()
	canonicalRecipe := CanonicalRecipe()
	if report.Schema != ReportSchema || report.Producer != ProducerID || report.Consumer != ConsumerID || report.ValidationFailure != nil || !validHead(report.HeadSHA) ||
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
	if report.BundleDigest == "" || report.ConformanceDecision != "PASS" || report.ConformanceResolution != "EXACT" || report.ConformanceReason != "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" ||
		report.ConformanceCoordinate != (Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", report.ConformanceReason}) ||
		report.SubjectArtifactDecision != "CARRIED" || report.SubjectArtifactResolution != "EVIDENCE_ATTACHED" || report.SubjectArtifactReason != "PROOF_CARRYING_ARTIFACT_EMITTED" ||
		report.ArtifactUseAuthority != "READ_ONLY_CONSUMPTION" || !caseInventoryOK(report.Cases) || len(report.Indicators) != len(MetricIDs()) || len(report.Proofs) != 3 ||
		len(report.Transitions) != TransitionTotal || len(report.Interventions) != 2 || report.NetChangedPaths != report.WriteSet.NetChangedPaths || report.CapabilityMutationGranted || report.PromotionAuthority || report.SemanticAuthority {
		return fmt.Errorf("proof-carrying report shape mismatch")
	}
	if report.ConsumerReceipt == (ConsumerReceipt{}) {
		return &ValidationError{Coordinate: Coordinate{"CONSUME_BUNDLE", "consumer-receipt", "CONSUMER_RECEIPT_MISSING"}, Detail: "final report has no consumer receipt"}
	}
	if report.Checkout.HeadSHA != report.HeadSHA || report.Checkout.ActualHeadSHA != report.HeadSHA || !validHead(report.Checkout.HeadSHA) || !validDigest(report.Checkout.TreeDigest) ||
		!validDigest(report.Checkout.SourceDigest) || !validDigest(report.Checkout.OperationDigest) || !validDigest(report.Checkout.RecipeDigest) || !validDigest(report.Checkout.ContractDigest) ||
		(report.BundleDigest != "" && !validDigest(report.BundleDigest)) {
		return fmt.Errorf("proof-carrying checkout binding mismatch")
	}
	if err := validateWriteSet(report.WriteSet); err != nil {
		return err
	}
	want := expectedSummary(ProofPhaseFinal, report.Summary.ProducerImportDenominator, 1, 1)
	if report.Summary != want || report.Summary.ProducerImportDenominator <= 0 || report.Summary.ProducerImportNumerator != 0 || report.Summary.CoreParserDependencies != CoreParserDependencyInventoryTotal {
		return &ValidationError{Coordinate: Coordinate{"EVALUATE", "final-summary", "FINAL_SUMMARY_MISMATCH"}, Detail: summaryMismatchError("FINAL", report.Summary, want).Error()}
	}
	if err := validateIndicatorInventory(report, ProofPhaseFinal, 1, 1); err != nil {
		return err
	}
	if !validDigest(report.UnauthorizedConsumerTargetDigest) || report.UnauthorizedConsumerOutputExists || report.UnauthorizedConsumerOutputDigest != "" || report.UnauthorizedConsumerErrorClass != string(ConsumerErrorAttestationMismatch) || !validDigest(report.UnauthorizedConsumerErrorDigest) {
		return fmt.Errorf("proof-carrying unauthorized consumer observation mismatch")
	}
	if len(report.Counterexamples) != CounterexampleTotal || !counterexampleInventoryOK(report.Counterexamples) {
		return fmt.Errorf("proof-carrying counterexample inventory mismatch")
	}
	if err := validateOpenLedger(report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying prior ledger mismatch: %w", err)
	}
	if err := validateFinalLedger(report.Ledger, report.PriorLedger); err != nil {
		return fmt.Errorf("proof-carrying final ledger mismatch: %w", err)
	}
	if err := validateCaseResults(report, ProofPhaseFinal); err != nil {
		return err
	}
	if err := validateLedgerAgainstValidCase(report); err != nil {
		return err
	}
	if err := validateTransitions(report.Transitions); err != nil {
		return err
	}
	if err := validateTransitionsAgainstClaims(report); err != nil {
		return err
	}
	if err := validateInterventions(report.Interventions); err != nil {
		return err
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied {
			return fmt.Errorf("proof-carrying indicator mismatch")
		}
	}
	if err := validateProofInventory(report, false); err != nil {
		return err
	}
	if report.ProofSummary != proofSummary(report.Proofs, ProofPhaseFinal, report.ArtifactUseAuthority) {
		return fmt.Errorf("proof-carrying proof summary mismatch")
	}
	if report.BundleDigest != "" {
		preliminary := canonicalPreliminaryProjection(report)
		if err := validatePhaseTransitionPair(preliminary.Cases, report.Cases); err != nil {
			return err
		}
		if report.ConsumerReceipt.PreliminaryDigest != preliminary.Digest {
			return &ValidationError{Coordinate: Coordinate{"CONSUME_BUNDLE", "consumer-receipt", "PRELIMINARY_BINDING_MISMATCH"}, Detail: "consumer receipt preliminary digest does not match the canonical projection"}
		}
		if report.ConsumerReceipt.Schema != ConsumerReceiptSchema || report.ConsumerReceipt.Version != 1 || report.ConsumerReceipt.Authority != "READ_ONLY_CONSUMPTION" ||
			report.ConsumerReceipt.Producer != ProducerID || report.ConsumerReceipt.Consumer != ConsumerID || !validDigest(report.ConsumerReceipt.PreliminaryDigest) || report.ConsumerReceipt.TargetPath != "artifact.json" ||
			!report.ConsumerReceipt.OutputExists || !validDigest(report.ConsumerReceipt.TargetDigest) || !validDigest(report.ConsumerReceipt.OutputDigest) || !validDigest(report.ConsumerReceipt.AttestationDigest) ||
			report.ConsumerReceipt.AttestationDigest != attestationDigest(report) || report.ConsumerReceipt.Digest != consumerReceiptDigest(report.ConsumerReceipt) {
			return fmt.Errorf("proof-carrying consumer receipt mismatch")
		}
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("proof-carrying report digest mismatch")
	}
	return nil
}

// ValidatePreliminary accepts only the exact pre-consumer observation. It is
// intentionally stricter than a shape check: the report must already contain
// every non-consumer case, claim, ledger, transition, proof, intervention, and
// binding needed for the consumer kernel to lift it to a final report.
func ValidatePreliminary(report Report) error {
	bundleMode := report.BundleDigest != ""
	wantScope := "CURRENT_CHECKOUT_OBSERVATION"
	wantBundleMetric, wantConsumerMetric := 0, 0
	wantReason := "BUNDLE_CONSUMPTION_NOT_OBSERVED"
	if bundleMode {
		wantScope = "BUNDLE_HISTORICAL_SUBJECT_BINDING"
		wantBundleMetric = 1
		wantReason = "CONSUMER_RECHECK_NOT_OBSERVED"
	}
	if report.ConformanceDecision == "PASS" || report.ConformanceResolution == "EXACT" || report.ArtifactUseAuthority != "" || report.ConsumerReceipt != (ConsumerReceipt{}) {
		return preliminaryValidationError(Coordinate{"CONSUME_AUTHORITY", "preliminary-report", "PRELIMINARY_FINAL_AUTHORITY_PRESENT"}, fmt.Errorf("preliminary report contains final consumer authority or receipt"))
	}
	if report.Schema != ReportSchema || report.Producer != ProducerID || report.Consumer != ConsumerID || report.ValidationFailure != nil || !validHead(report.HeadSHA) ||
		report.ContractDigest != digestValue(CanonicalContract()) || report.RecipeDigest != digestValue(CanonicalRecipe()) || report.RecipeVersion != CanonicalRecipe().Version ||
		!validDigest(report.IndependenceDigest) || report.AuthorityObservation != "UNKNOWN_GLOBAL_TRANSIENT_SCOPE" || report.ArtifactUseAuthority != "" || report.ConsumerReceipt != (ConsumerReceipt{}) ||
		(bundleMode && !validDigest(report.BundleDigest)) || report.CheckoutBindingScope != wantScope ||
		report.ConformanceDecision != "FAIL_CLOSED" || report.ConformanceResolution != "LOWER_RESOLUTION" || report.ConformanceReason != wantReason ||
		report.ConformanceCoordinate != (Coordinate{"CONSUME_BUNDLE", "consumer-recheck", report.ConformanceReason}) ||
		report.PreliminaryDecision != "FAIL_CLOSED" || report.PreliminaryResolution != "LOWER_RESOLUTION" || report.PreliminaryReason != wantReason ||
		report.PreliminaryCoordinate != report.ConformanceCoordinate {
		return fmt.Errorf("proof-carrying preliminary report identity mismatch")
	}
	if report.Checkout.HeadSHA != report.HeadSHA || report.Checkout.ActualHeadSHA != report.HeadSHA || !validHead(report.Checkout.HeadSHA) || !validDigest(report.Checkout.TreeDigest) ||
		!validDigest(report.Checkout.SourceDigest) || !validDigest(report.Checkout.OperationDigest) || !validDigest(report.Checkout.RecipeDigest) || !validDigest(report.Checkout.ContractDigest) ||
		report.Checkout.SourceDigest == "" || report.Checkout.OperationDigest == "" {
		return fmt.Errorf("proof-carrying preliminary checkout binding mismatch")
	}
	if report.SubjectArtifactDecision != "CARRIED" || report.SubjectArtifactResolution != "EVIDENCE_ATTACHED" || report.SubjectArtifactReason != "PROOF_CARRYING_ARTIFACT_EMITTED" ||
		report.NetChangedPaths != 0 || report.NetChangedPaths != report.WriteSet.NetChangedPaths || report.CapabilityMutationGranted || report.PromotionAuthority || report.SemanticAuthority ||
		report.NotClaimed == nil {
		return fmt.Errorf("proof-carrying preliminary authority or subject mismatch")
	}
	if err := validateWriteSet(report.WriteSet); err != nil {
		return err
	}
	if report.WriteSet.CapabilityMutationGranted || report.WriteSet.NetChangedPaths != 0 || report.WriteSet.NetUnchanged != true || report.WriteSet.TransientUnknown != true ||
		report.WriteSet.ActualWritesObservation != "UNKNOWN" || report.WriteSet.GlobalMutationAuthority != "UNKNOWN" {
		return fmt.Errorf("proof-carrying preliminary write observation mismatch")
	}
	wantSummary := expectedSummary(ProofPhasePreliminary, report.Summary.ProducerImportDenominator, wantBundleMetric, wantConsumerMetric)
	if report.Summary != wantSummary ||
		report.Summary.ProducerImportDenominator <= 0 || report.Summary.ProducerImportNumerator != 0 || report.Summary.CoreParserDependencies != CoreParserDependencyInventoryTotal ||
		report.IndependenceDigest != digestValue(IndependenceEvidence{Schema: "gooo/language-proof-carrying-artifact-independence/v1", ProducerDependencies: report.Summary.ProducerDependencies, ProducerImportNumerator: report.Summary.ProducerImportNumerator, ProducerImportDenominator: report.Summary.ProducerImportDenominator, CoreParserDependencies: report.Summary.CoreParserDependencies}) {
		return preliminaryValidationError(Coordinate{"EVALUATE", "preliminary-summary", "PRELIMINARY_SUMMARY_MISMATCH"}, summaryMismatchError("PRELIMINARY", report.Summary, wantSummary))
	}
	if err := validateIndicatorInventory(report, ProofPhasePreliminary, wantBundleMetric, wantConsumerMetric); err != nil {
		return preliminaryValidationError(Coordinate{"EVALUATE", "preliminary-inventory", "PRELIMINARY_INDICATOR_INVENTORY_MISMATCH"}, err)
	}
	if err := validateCaseResults(report, ProofPhasePreliminary); err != nil {
		return err
	}
	if err := validateOpenLedger(report.PriorLedger); err != nil {
		return preliminaryValidationError(Coordinate{"CONSUME_LEDGER", "preliminary-ledger", "PRELIMINARY_CLAIM_LEDGER_MISMATCH"}, err)
	}
	if err := validateFinalLedger(report.Ledger, report.PriorLedger); err != nil {
		return preliminaryValidationError(Coordinate{"CONSUME_LEDGER", "preliminary-ledger", "PRELIMINARY_CLAIM_LEDGER_MISMATCH"}, err)
	}
	if err := validateLedgerAgainstValidCase(report); err != nil {
		return preliminaryValidationError(Coordinate{"CONSUME_LEDGER", "preliminary-ledger", "PRELIMINARY_CLAIM_LEDGER_MISMATCH"}, err)
	}
	if err := validateTransitions(report.Transitions); err != nil {
		return preliminaryValidationError(Coordinate{"CONSUME", "preliminary-transition-chain", "PRELIMINARY_TRANSITION_CHAIN_MISMATCH"}, err)
	}
	if err := validateTransitionsAgainstClaims(report); err != nil {
		return preliminaryValidationError(Coordinate{"CONSUME", "preliminary-transition-chain", "PRELIMINARY_TRANSITION_CHAIN_MISMATCH"}, err)
	}
	if err := validateInterventions(report.Interventions); err != nil {
		return err
	}
	if err := validatePhaseState(report.Cases, ProofPhasePreliminary); err != nil {
		return err
	}
	if err := validateProofInventory(report, true); err != nil {
		return preliminaryValidationError(Coordinate{"VERIFY_PROOF", "preliminary-proof-gate", "PRELIMINARY_PROOF_NOT_SATISFIED"}, err)
	}
	if report.ProofSummary != proofSummary(report.Proofs, ProofPhasePreliminary, report.ArtifactUseAuthority) {
		return preliminaryValidationError(Coordinate{"VERIFY_PROOF", "preliminary-proof-summary", "PRELIMINARY_PROOF_SUMMARY_MISMATCH"}, fmt.Errorf("proof-carrying proof summary mismatch"))
	}
	if len(report.Counterexamples) != CounterexampleTotal || !counterexampleInventoryOK(report.Counterexamples) {
		return fmt.Errorf("proof-carrying preliminary counterexample inventory mismatch")
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("proof-carrying preliminary report digest mismatch")
	}
	return nil
}

// canonicalPreliminaryProjection removes only the consumer observation from a
// valid final report and projects the one declared evidence-time claim state
// transition back to PRELIMINARY. Everything else that establishes the
// historical subject and its producer-side evidence remains represented in
// the projection, so the receipt must bind to this digest rather than to an
// arbitrary self-consistent claim.
func canonicalPreliminaryProjection(report Report) Report {
	preliminary := report
	preliminary.ConsumerReceipt = ConsumerReceipt{}
	preliminary.ArtifactUseAuthority = ""
	preliminary.Summary.ConsumerRechecks = 0
	preliminary.ConformanceDecision = "FAIL_CLOSED"
	preliminary.ConformanceResolution = "LOWER_RESOLUTION"
	preliminary.ConformanceReason = "CONSUMER_RECHECK_NOT_OBSERVED"
	preliminary.ConformanceCoordinate = Coordinate{"CONSUME_BUNDLE", "consumer-recheck", preliminary.ConformanceReason}
	preliminary.PreliminaryDecision = preliminary.ConformanceDecision
	preliminary.PreliminaryResolution = preliminary.ConformanceResolution
	preliminary.PreliminaryReason = preliminary.ConformanceReason
	preliminary.PreliminaryCoordinate = preliminary.ConformanceCoordinate
	preliminary.Cases = projectCasesForPhase(report.Cases, ProofPhasePreliminary)
	preliminary.Summary = summarize(preliminary.Cases, IndependenceEvidence{
		ProducerDependencies: report.Summary.ProducerDependencies, ProducerImportNumerator: report.Summary.ProducerImportNumerator,
		ProducerImportDenominator: report.Summary.ProducerImportDenominator, CoreParserDependencies: report.Summary.CoreParserDependencies,
	}, preliminary.WriteSet, preliminary.Interventions, preliminary.Ledger, preliminary.Summary.BundleOnlyVerification, 0)
	preliminary.Indicators = indicators(preliminary.Summary, ProofPhasePreliminary)
	preliminary.Proofs = proofs(preliminary, preliminary.Cases, ProofPhasePreliminary)
	preliminary.ProofSummary = proofSummary(preliminary.Proofs, ProofPhasePreliminary, preliminary.ArtifactUseAuthority)
	preliminary.Digest = reportDigest(preliminary)
	return preliminary
}

// ResealReportDigest is a fixture-only helper exposed to the CI verifier
// command. It recomputes the report envelope after a deliberate mutation;
// semantic validators still decide whether the resealed report is valid.
func ResealReportDigest(report Report) Report {
	report.Digest = reportDigest(report)
	return report
}

// ResealFinalPreliminaryDigest creates a coherent final-report fixture whose
// receipt, attestation, and outer report all agree on an intentionally chosen
// preliminary digest. Validate must still reject it unless that digest is the
// canonical preliminary projection.
func ResealFinalPreliminaryDigest(report Report, preliminaryDigest string) Report {
	receipt := report.ConsumerReceipt
	receipt.PreliminaryDigest = preliminaryDigest
	report.ConsumerReceipt = receipt
	report.ConsumerReceipt.AttestationDigest = attestationDigest(report)
	report.ConsumerReceipt.Digest = consumerReceiptDigest(report.ConsumerReceipt)
	report.Digest = reportDigest(report)
	return report
}

// ResealClaimState creates a fixture with a valid claim-state digest and
// report digest. The fixed expectation validator, rather than stale-envelope
// detection, must reject the semantic state mutation.
func ResealClaimState(report Report, caseID, claimID, state string) Report {
	for caseIndex := range report.Cases {
		if report.Cases[caseIndex].ID != caseID {
			continue
		}
		for claimIndex := range report.Cases[caseIndex].Claims {
			claim := &report.Cases[caseIndex].Claims[claimIndex]
			if claim.ID != claimID {
				continue
			}
			claim.Status = state
			switch state {
			case "DISCHARGED":
				claim.Resolution, claim.Reason, claim.Provenance = "EXACT", "CLAIM_DISCHARGED", "consumer-canonical-recipe-v2"
			case "OPEN":
				claim.Resolution, claim.Reason, claim.Provenance = "LOWER_RESOLUTION", "CLAIM_PENDING", "consumer-observation"
			case "REFUTED":
				claim.Resolution, claim.Reason, claim.Provenance = "INVARIANT_ONLY", "CLAIM_REFUTED", "consumer-observation"
			}
			claim.StateDigest = claimStateDigest(*claim)
		}
	}
	report.Digest = reportDigest(report)
	return report
}

type ValidationError struct {
	Coordinate Coordinate
	Detail     string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("proof-carrying validation rejected at stage=%s step=%s reason=%s: %s", e.Coordinate.Stage, e.Coordinate.Step, e.Coordinate.Reason, e.Detail)
}

func preliminaryValidationError(coordinate Coordinate, detail error) error {
	return &ValidationError{Coordinate: coordinate, Detail: detail.Error()}
}

type summaryMismatchDetail struct {
	Phase    string  `json:"phase"`
	Actual   Summary `json:"actual"`
	Expected Summary `json:"expected"`
}

func summaryMismatchError(phase string, actual, expected Summary) error {
	detail, err := json.Marshal(summaryMismatchDetail{Phase: phase, Actual: actual, Expected: expected})
	if err != nil {
		return fmt.Errorf("proof-carrying summary mismatch")
	}
	return fmt.Errorf("proof-carrying summary mismatch: %s", detail)
}

func expectedSummary(phase string, denominator, bundleMetric, consumerMetric int) Summary {
	claimStateTotals := fixedClaimStateTotals(phase)
	return Summary{
		CasesSatisfied: CaseTotal, CasesTotal: CaseTotal, ValidArtifacts: 1,
		EvidenceKindsCarried: EvidenceTotal, ExactEvidenceLinks: EvidenceTotal,
		RecipeMatches: 1, PreservedTransitions: EvidenceTotal + 1,
		TransitionTotal: TransitionTotal, ClaimTemplates: ClaimTemplateTotal,
		ClaimInstances: CaseTotal * ClaimTemplateTotal, AcceptedTransitions: TransitionTotal,
		CaseDischargedClaims: claimStateTotals.Discharged, CaseOpenClaims: claimStateTotals.Open, CaseRefutedClaims: claimStateTotals.Refuted,
		FinalLedgerOpenClaims: ClaimTemplateTotal, FinalLedgerDischargedClaims: ClaimTemplateTotal,
		TamperedRejections: 1, CoherentTamperRejections: 1, CoherentClaimStructureRejections: 4,
		MissingEvidenceRejections: 1, ByteOnlyDenials: 1, RecipeRejections: 1,
		RecipeOnlyRejections: 1, MissingAttachmentRejections: 1, WrongAttachmentRejections: 1,
		UnrelatedEvidenceRejections: 1, StaleHeadRejections: 1, UnauthorizedConsumerDenials: 1,
		SemanticInterventions: 1, NonsemanticInterventions: 1, ReadOnlyAuthorities: 1,
		ProducerDependencies: 0, ProducerImportNumerator: 0, ProducerImportDenominator: denominator,
		CoreParserDependencies: CoreParserDependencyInventoryTotal, NetRepositoryStateUnchanged: 1,
		UnknownAuthorityObservations: 1, BundleOnlyVerification: bundleMetric, ConsumerRechecks: consumerMetric,
		GeneratedAuthority: 0, SemanticClaims: 0, NetChangedPaths: 0,
		MutationAuthorities: 0, PromotionAuthorities: 0, SemanticAuthorities: 0,
	}
}

func validateIndicatorInventory(report Report, phase string, bundleMetric, consumerMetric int) error {
	if len(report.Indicators) != len(MetricIDs()) {
		return fmt.Errorf("proof-carrying indicator inventory size mismatch")
	}
	want := indicators(expectedSummary(phase, report.Summary.ProducerImportDenominator, bundleMetric, consumerMetric), phase)
	byID := make(map[string]Indicator, len(report.Indicators))
	for _, item := range report.Indicators {
		if _, exists := byID[item.MetricID]; exists {
			return fmt.Errorf("proof-carrying indicator inventory duplicate: %s", item.MetricID)
		}
		byID[item.MetricID] = item
	}
	for index, id := range MetricIDs() {
		item, exists := byID[id]
		if !exists {
			return fmt.Errorf("proof-carrying indicator inventory missing: %s", id)
		}
		if item != want[index] {
			return fmt.Errorf("proof-carrying indicator mismatch: %s", id)
		}
	}
	return nil
}

func validateCaseResults(report Report, phase string) error {
	if !caseInventoryOK(report.Cases) || len(report.Cases) != CaseTotal {
		return fmt.Errorf("proof-carrying case inventory mismatch")
	}
	claimIDs := make([]string, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		claimIDs = append(claimIDs, spec.ID)
	}
	for index, item := range report.Cases {
		spec := CanonicalContract().Cases[index]
		wantCoordinate, ok := caseCoordinates()[item.ID]
		if !ok || item.Status != "SATISFIED" || item.ExpectedDecision != spec.ExpectedDecision || item.ExpectedResolution != spec.ExpectedResolution || item.ExpectedReason != spec.ExpectedReason ||
			item.ObservedDecision != spec.ExpectedDecision || item.ObservedResolution != spec.ExpectedResolution || item.ObservedReason != spec.ExpectedReason || item.Coordinate != wantCoordinate {
			return fmt.Errorf("proof-carrying case mismatch: %s", item.ID)
		}
		if len(item.Claims) != ClaimTemplateTotal {
			return fmt.Errorf("proof-carrying case claim count mismatch: %s", item.ID)
		}
		seen := map[string]bool{}
		for claimIndex, claim := range item.Claims {
			if claim.ID != claimIDs[claimIndex] || seen[claim.ID] || claim.Provenance == "" || claim.StateDigest != claimStateDigest(claim) ||
				(claim.Status == "DISCHARGED" && claim.Resolution != "EXACT") || (claim.Status == "OPEN" && claim.Resolution != "LOWER_RESOLUTION") ||
				(claim.Status == "REFUTED" && claim.Resolution != "INVARIANT_ONLY") || (claim.Status != "DISCHARGED" && claim.Status != "OPEN" && claim.Status != "REFUTED") ||
				(len(claim.EvidenceDigests) == 0 && claim.EvidenceDigest != "") || (len(claim.EvidenceDigests) > 0 && claim.EvidenceDigest != claim.EvidenceDigests[0]) {
				return fmt.Errorf("proof-carrying claim evaluation mismatch: %s", item.ID)
			}
			seen[claim.ID] = true
		}
	}
	if err := validateClaimStateExpectations(phase, report.Cases); err != nil {
		return err
	}
	if valid := validCase(report.Cases); valid == nil || !validCaseClaims(*valid, report) {
		return fmt.Errorf("proof-carrying valid claim binding mismatch")
	}
	return nil
}

func caseCoordinates() map[string]Coordinate {
	return map[string]Coordinate{
		"valid-proof-carrying-artifact":      {"CONSUME_AUTHORITY", "grant-read-only-consumption", "CONSUMER_ONLY_READ_ONLY_AUTHORITY"},
		"tampered-evidence":                  {"CONSUME_EVIDENCE", "evidence-digest", "PROOF_EVIDENCE_DIGEST_MISMATCH"},
		"coherent-tamper-reconstruction":     {"CONSUME_OPERATION", "receipt", "OPERATION_RECONSTRUCTION_MISMATCH"},
		"missing-operation-evidence":         {"CONSUME_EVIDENCE", "operation-evidence", "PROOF_EVIDENCE_MISSING"},
		"bytes-only-no-authority":            {"CONSUME_INPUT", "external-evidence", "ARTIFACT_BYTES_NOT_AUTHORITY"},
		"independent-recipe-mismatch":        {"CONSUME_RECIPE", "recipe", "INDEPENDENT_RECIPE_MISMATCH"},
		"recipe-only-mismatch":               {"CONSUME_RECIPE", "recipe", "INDEPENDENT_RECIPE_MISMATCH"},
		"missing-attachment":                 {"CONSUME_INPUT", "operation-attachment", "ARTIFACT_ATTACHMENT_MISSING"},
		"wrong-attachment-digest":            {"CONSUME_OPERATION", "attachment-digest", "OPERATION_ATTACHMENT_DIGEST_MISMATCH"},
		"unrelated-evidence-tamper":          {"CONSUME_EVIDENCE", "invariant-evidence", "INVARIANT_EVIDENCE_NOT_PRESERVED"},
		"stale-head":                         {"CONSUME_IDENTITY", "head", "HEAD_BINDING_MISMATCH"},
		"unauthorized-consumer":              {"CONSUME_BUNDLE", "attestation", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED"},
		"coherent-claim-proposition-tamper":  {"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"},
		"coherent-claim-dependency-tamper":   {"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"},
		"coherent-claim-proof-choice-tamper": {"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"},
		"coherent-claim-target-tamper":       {"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"},
	}
}

func validateLedgerAgainstValidCase(report Report) error {
	valid := validCase(report.Cases)
	if valid == nil || len(valid.Claims) != ClaimTemplateTotal || len(report.PriorLedger.Entries) != ClaimTemplateTotal || len(report.Ledger.Entries) != ClaimTemplateTotal*2 {
		return fmt.Errorf("proof-carrying claim ledger subject mismatch")
	}
	for index, claim := range valid.Claims {
		prior := report.PriorLedger.Entries[index]
		if prior.ClaimID != claim.ID || prior.Proposition != claim.Proposition || prior.TargetDigest != claim.TargetDigest || !reflect.DeepEqual(prior.Dependencies, claim.Dependencies) ||
			prior.Status != "OPEN" || prior.Resolution != "LOWER_RESOLUTION" || prior.Producer != ProducerID || prior.Consumer != ConsumerID || prior.ProofChoice != claim.ProofChoice ||
			prior.MetaOperation != claim.MetaOperation || prior.Coordinate != claim.Coordinate || prior.Reason != "AWAITING_INDEPENDENT_RECHECK" || !reflect.DeepEqual(prior.EvidenceDigest, claim.EvidenceDigests) || prior.Provenance != "producer-carried-prior-ledger" {
			return fmt.Errorf("proof-carrying prior ledger subject mismatch")
		}
		final := report.Ledger.Entries[ClaimTemplateTotal+index]
		if final.ClaimID != claim.ID || final.Proposition != claim.Proposition || final.TargetDigest != claim.TargetDigest || !reflect.DeepEqual(final.Dependencies, claim.Dependencies) ||
			final.Status != "DISCHARGED" || final.Resolution != "EXACT" || final.Producer != ProducerID || final.Consumer != ConsumerID || final.ProofChoice != claim.ProofChoice ||
			final.MetaOperation != claim.MetaOperation || final.Coordinate != claim.Coordinate || final.Reason != "INDEPENDENT_SOURCE_OPERATION_RECIPE_RECHECKED" || !reflect.DeepEqual(final.EvidenceDigest, claim.EvidenceDigests) || final.Provenance != "consumer-canonical-recipe-v2" {
			return fmt.Errorf("proof-carrying final ledger subject mismatch")
		}
	}
	return nil
}

func validateTransitionsAgainstClaims(report Report) error {
	valid := validCase(report.Cases)
	if valid == nil {
		return fmt.Errorf("proof-carrying transition subject missing")
	}
	want := claimTransitions(valid.Claims)
	if !reflect.DeepEqual(report.Transitions, want) {
		return fmt.Errorf("proof-carrying transition subject mismatch")
	}
	return nil
}

func validateInterventions(interventions []InterventionResult) error {
	if len(interventions) != 2 {
		return fmt.Errorf("proof-carrying intervention inventory mismatch")
	}
	for index, item := range interventions {
		wantKind := "SEMANTIC"
		wantID := "semantic-source-intervention"
		if index == 1 {
			wantKind = "NONSEMANTIC"
			wantID = "comment-only-intervention"
		}
		if item.ID != wantID || item.Kind != wantKind || item.Status != "SATISFIED" || item.Reason == "" ||
			!validDigest(item.RawSourceDigestBefore) || !validDigest(item.RawSourceDigestAfter) || !validDigest(item.SemanticDigestBefore) || !validDigest(item.SemanticDigestAfter) ||
			!validDigest(item.OperationReceiptDigestBefore) || !validDigest(item.OperationReceiptDigestAfter) || !validDigest(item.EvidenceLinkDigestBefore) || !validDigest(item.EvidenceLinkDigestAfter) ||
			!validDigest(item.ClaimTransitionDigestBefore) || !validDigest(item.ClaimTransitionDigestAfter) || item.ConsumerDecisionBefore != "PASS" || item.ConsumerDecisionAfter != "PASS" ||
			!item.RawDigestChanged || !item.ConsumerDecisionPreserved {
			return fmt.Errorf("proof-carrying intervention mismatch: %s", item.ID)
		}
		if index == 0 {
			if !item.SemanticDigestChanged || !item.OperationReceiptChanged || !item.EvidenceLinksChanged || !item.ClaimTransitionsChanged || item.SemanticDigestPreserved {
				return fmt.Errorf("proof-carrying semantic intervention mismatch")
			}
		} else if !item.SemanticDigestPreserved || item.SemanticDigestChanged {
			return fmt.Errorf("proof-carrying nonsemantic intervention mismatch")
		}
	}
	return nil
}

func validateProofInventory(report Report, preliminary bool) error {
	if len(report.Proofs) != 3 {
		return fmt.Errorf("proof-carrying proof inventory mismatch")
	}
	phase := ProofPhaseFinal
	if preliminary {
		phase = ProofPhasePreliminary
	}
	want := proofs(report, report.Cases, phase)
	if !reflect.DeepEqual(report.Proofs, want) {
		return fmt.Errorf("proof-carrying proof projection mismatch")
	}
	for index, proof := range report.Proofs {
		if !proof.EvidenceValidated || !validDigest(proof.TargetDigest) || !validDigest(proof.ReceiptDigest) || len(proof.EvidenceDigests) == 0 {
			return fmt.Errorf("proof-carrying proof binding mismatch")
		}
		for _, digest := range proof.EvidenceDigests {
			if !validDigest(digest) {
				return fmt.Errorf("proof-carrying proof evidence mismatch")
			}
		}
		if proof.Passed != (proof.State == ProofStateDischarged) {
			return fmt.Errorf("proof-carrying proof discharge state mismatch")
		}
		if index == 0 && (!proof.EvidenceValidated || proof.ConsumerGateOpen || proof.Phase != phase || (preliminary && proof.State != ProofStateObserved) || (!preliminary && proof.State != ProofStateDischarged)) {
			return fmt.Errorf("proof-carrying foundation proof mismatch")
		}
		if preliminary {
			if index > 0 && (proof.Passed || !proof.ConsumerGateOpen || proof.Phase != ProofPhasePreliminary || proof.State != ProofStateOpen) {
				return fmt.Errorf("proof-carrying preliminary consumer proof gate mismatch")
			}
		} else if !proof.Passed || proof.ConsumerGateOpen || proof.Phase != ProofPhaseFinal || proof.State != ProofStateDischarged {
			return fmt.Errorf("proof-carrying final proof mismatch")
		}
	}
	return nil
}

func counterexampleInventoryOK(items []Counterexample) bool {
	want := []string{"bundle-not-provided", "bundle-corrupt", "unauthorized-attestation-mismatch", "consumer-target-missing", "consumer-output-absent", "proof-false", "main-indicator-38-of-40"}
	wantCoordinates := []Coordinate{
		{"CONSUME_BUNDLE", "read-bundle", "BUNDLE_CONSUMPTION_NOT_OBSERVED"},
		{"CONSUME_BUNDLE", "validate-bundle", "BUNDLE_INVALID"},
		{"CONSUME_BUNDLE", "attestation", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED"},
		{"CONSUME_BUNDLE", "target-missing", "TARGET_MISSING"},
		{"CONSUME_BUNDLE", "receipt", "CONSUMER_OUTPUT_ABSENT"},
		{"VERIFY_PROOF", "final-gate", "PROOF_NOT_SATISFIED"},
		{"EVALUATE", "final-conformance-gate", "INDICATOR_GATE_NOT_SATISFIED"},
	}
	wantResolution := []string{"LOWER_RESOLUTION", "LOWER_RESOLUTION", "INVARIANT_ONLY", "LOWER_RESOLUTION", "LOWER_RESOLUTION", "INVARIANT_ONLY", "LOWER_RESOLUTION"}
	wantClaim := []string{"consumer-authority", "consumer-authority", "consumer-authority", "consumer-authority", "consumer-authority", "proof-gate", "conformance"}
	wantTo := []string{"OPEN", "OPEN", "REFUTED", "OPEN", "OPEN", "OPEN", "OPEN"}
	wantErrorClass := []string{"BUNDLE_INVALID", "BUNDLE_INVALID", "ATTESTATION_MISMATCH", "TARGET_MISSING", "RECEIPT_MISMATCH", "PROOF_GATE", "INDICATOR_GATE"}
	if len(items) != len(want) {
		return false
	}
	for index, item := range items {
		if item.ID != want[index] || item.Decision != "FAIL_CLOSED" || item.Resolution != wantResolution[index] || item.Coordinate != wantCoordinates[index] ||
			item.ClaimID != wantClaim[index] || item.From != "CARRIED" || item.To != wantTo[index] || item.ErrorClass != wantErrorClass[index] || !validDigest(item.ErrorDigest) || item.OutputExists {
			return false
		}
	}
	return true
}

func validCaseClaims(item CaseResult, report Report) bool {
	if len(item.Claims) != ClaimTemplateTotal || !validDigest(item.SourceDigest) || !validDigest(item.OperationDigest) || !validDigest(item.SemanticDigest) || !validDigest(item.ArtifactDigest) ||
		item.SourceDigest != report.Checkout.SourceDigest || item.OperationAttachmentDigest != report.Checkout.OperationDigest || item.RecipeAttachmentDigest != report.Checkout.RecipeDigest {
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
