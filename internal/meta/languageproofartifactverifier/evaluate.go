package languageproofartifactverifier

import (
	"errors"
	"reflect"
)

func Evaluate(input Input) Report {
	canonicalContract := CanonicalContract()
	canonicalRecipe := CanonicalRecipe()
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Producer: ProducerID, Consumer: ConsumerID,
		ContractDigest: digestValue(canonicalContract), RecipeDigest: digestValue(canonicalRecipe), RecipeVersion: canonicalRecipe.Version,
		IndependenceDigest: digestValue(input.Independence), BundleDigest: input.BundleDigest, Checkout: input.Checkout,
		AuthorityObservation: "UNKNOWN_GLOBAL_TRANSIENT_SCOPE", NotClaimed: []string{
			"full compiler semantic correctness", "cryptographic signer authenticity", "external side effects", "global mutation authority", "authority from artifact bytes alone"}}
	if input.BundleDigest == "" {
		report.CheckoutBindingScope = "CURRENT_CHECKOUT_OBSERVATION"
	} else {
		report.CheckoutBindingScope = "BUNDLE_HISTORICAL_SUBJECT_BINDING"
	}
	if !reflect.DeepEqual(input.Contract, canonicalContract) || !validHead(input.HeadSHA) ||
		input.Independence.Schema != "gooo/language-proof-carrying-artifact-independence/v1" ||
		input.Independence.ProducerDependencies != input.Independence.ProducerImportNumerator ||
		input.Independence.ProducerImportNumerator != 0 || input.Independence.ProducerImportDenominator <= 0 || input.Independence.CoreParserDependencies != CoreParserDependencyInventoryTotal ||
		!validCheckout(input, canonicalContract) {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_CONTRACT_OR_INDEPENDENCE_MISMATCH")
	}
	validArtifact, err := decodeStrict[Artifact](input.ValidArtifact)
	if err != nil || validateWriteSet(validArtifact.WriteSet) != nil || !reflect.DeepEqual(validArtifact.WriteSet, input.WriteSet) {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_WRITE_SET_NOT_BOUND")
	}

	phase := ProofPhasePreliminary
	if input.ConsumerReceiptProvided {
		phase = ProofPhaseFinal
	}
	results := make([]CaseResult, 0, len(input.Contract.Cases))
	for _, definition := range input.Contract.Cases {
		var observed observation
		if definition.InputKind == "UNAUTHORIZED_CONSUMER" {
			observed = unauthorizedConsumerObservation(input, phase)
		} else {
			artifact, source, operation, recipe := caseInput(input, definition.InputKind)
			observed = verifyArtifact(artifact, source, operation, recipe, input.HeadSHA, phase)
		}
		status := "NOT_SATISFIED"
		if observed.Decision == definition.ExpectedDecision && observed.Resolution == definition.ExpectedResolution && observed.Reason == definition.ExpectedReason {
			status = "SATISFIED"
		}
		results = append(results, CaseResult{ID: definition.ID, Status: status,
			ExpectedDecision: definition.ExpectedDecision, ExpectedResolution: definition.ExpectedResolution, ExpectedReason: definition.ExpectedReason,
			ObservedDecision: observed.Decision, ObservedResolution: observed.Resolution, ObservedReason: observed.Reason,
			ProofChoice: definition.ProofChoice, MetaOperation: definition.MetaOperation, Coordinate: observed.Coordinate,
			Claims: observed.Claims, ArtifactDigest: observed.ArtifactDigest, SourceDigest: observed.SourceDigest,
			SemanticDigest: observed.SemanticDigest, OperationDigest: observed.OperationDigest, OperationAttachmentDigest: observed.OperationAttachmentDigest, RecipeAttachmentDigest: observed.RecipeAttachmentDigest, ConsumerTargetDigest: observed.ConsumerTargetDigest, ConsumerOutputDigest: observed.ConsumerOutputDigest, ConsumerOutputExists: observed.ConsumerOutputExists, ConsumerErrorClass: observed.ConsumerErrorClass, ConsumerErrorDigest: observed.ConsumerErrorDigest})
	}

	interventions := evaluateInterventions(input.Interventions, input.HeadSHA, phase)
	report.WriteSet = input.WriteSet
	report.NetChangedPaths = input.WriteSet.NetChangedPaths
	report.CapabilityMutationGranted = input.WriteSet.CapabilityMutationGranted
	report.PromotionAuthority = false
	report.SemanticAuthority = false
	report.Interventions = interventions
	report.Cases = results
	if artifact, decodeErr := decodeStrict[Artifact](input.ValidArtifact); decodeErr == nil {
		report.SubjectArtifactDecision, report.SubjectArtifactResolution, report.SubjectArtifactReason = artifact.Decision, artifact.Resolution, artifact.Reason
		report.PriorLedger = artifact.PriorLedger
	}
	report.Ledger = report.PriorLedger
	if validCase := validCase(results); validCase != nil {
		report.Ledger = dischargeLedger(report.PriorLedger, results)
	}
	bundleOnlyVerification, consumerRechecks := verificationMode(input.BundleDigest, input.ConsumerReceiptProvided)
	report.Summary = summarize(results, input.Independence, input.WriteSet, interventions, report.Ledger, bundleOnlyVerification, consumerRechecks)
	report.Transitions = transitions(results)
	for _, item := range results {
		if item.ID == "unauthorized-consumer" {
			report.UnauthorizedConsumerTargetDigest = item.ConsumerTargetDigest
			report.UnauthorizedConsumerOutputDigest = item.ConsumerOutputDigest
			report.UnauthorizedConsumerOutputExists = item.ConsumerOutputExists
			report.UnauthorizedConsumerErrorClass = item.ConsumerErrorClass
			report.UnauthorizedConsumerErrorDigest = item.ConsumerErrorDigest
		}
	}
	structuralGate := report.Summary.CasesSatisfied == CaseTotal && report.Summary.ValidArtifacts == 1 && report.Summary.PreservedTransitions == EvidenceTotal+1 &&
		report.Summary.TamperedRejections == 1 && report.Summary.CoherentTamperRejections == 1 && report.Summary.MissingEvidenceRejections == 1 &&
		report.Summary.ByteOnlyDenials == 1 && report.Summary.RecipeRejections == 1 && report.Summary.RecipeOnlyRejections == 1 &&
		report.Summary.MissingAttachmentRejections == 1 && report.Summary.WrongAttachmentRejections == 1 && report.Summary.UnrelatedEvidenceRejections == 1 &&
		report.Summary.StaleHeadRejections == 1 && report.Summary.UnauthorizedConsumerDenials == 1 && report.Summary.CoherentClaimStructureRejections == 4 && report.Summary.ProducerDependencies == 0 &&
		report.Summary.SemanticInterventions == 1 && report.Summary.NonsemanticInterventions == 1 && report.Summary.ReadOnlyAuthorities == 1 &&
		report.Summary.FinalLedgerOpenClaims == ClaimTemplateTotal && report.Summary.FinalLedgerDischargedClaims == ClaimTemplateTotal &&
		report.Summary.ProducerImportNumerator == 0 && report.Summary.ProducerImportDenominator > 0 && report.Summary.NetRepositoryStateUnchanged == 1 &&
		report.Summary.UnknownAuthorityObservations == 1 &&
		report.Summary.GeneratedAuthority == 0 && report.Summary.NetChangedPaths == 0 && report.Summary.MutationAuthorities == 0 &&
		report.Summary.PromotionAuthorities == 0 && report.Summary.SemanticAuthorities == 0
	report.PreliminaryDecision, report.PreliminaryResolution, report.PreliminaryReason = "FAIL_CLOSED", "LOWER_RESOLUTION", "BUNDLE_CONSUMPTION_NOT_OBSERVED"
	report.PreliminaryCoordinate = Coordinate{"CONSUME_BUNDLE", "consumer-recheck", report.PreliminaryReason}
	if structuralGate && input.BundleDigest != "" {
		// Build the candidate attestation privately so the receipt binds the
		// exact PASS authority statement. The public conformance decision below
		// is assigned only after indicators and proofs have been evaluated.
		candidate := consumerAttestedReport(report)
		if input.ConsumerReceiptProvided {
			candidate.ConsumerReceipt = input.ConsumerReceipt
			report.PreliminaryDecision, report.PreliminaryResolution, report.PreliminaryReason = candidate.ConformanceDecision, candidate.ConformanceResolution, candidate.ConformanceReason
			report.PreliminaryCoordinate = candidate.ConformanceCoordinate
			report.ConsumerReceipt = candidate.ConsumerReceipt
		} else {
			report.PreliminaryDecision, report.PreliminaryResolution, report.PreliminaryReason = "FAIL_CLOSED", "LOWER_RESOLUTION", "CONSUMER_RECHECK_NOT_OBSERVED"
			report.PreliminaryCoordinate = Coordinate{"CONSUME_BUNDLE", "consumer-recheck", report.PreliminaryReason}
		}
	}
	// The producer records targets from its observations. Validator-owned
	// phase expectations are applied only by Validate/ValidatePreliminary.
	report.Indicators = observedIndicators(report.Summary)
	proofReport := report
	proofPhase := ProofPhasePreliminary
	if structuralGate && input.BundleDigest != "" && input.ConsumerReceiptProvided {
		proofReport = consumerAttestedReport(proofReport)
		proofPhase = ProofPhaseFinal
	}
	report.Proofs = proofs(proofReport, results, proofPhase)
	report.Counterexamples = fixedCounterexamples(input, report)
	indicatorsOK := allIndicatorsSatisfied(report.Indicators)
	proofsOK := allProofsPassed(report.Proofs)
	consumerGate := false
	if input.ConsumerReceiptProvided {
		consumerGate = input.BundleDigest != "" && consumerReceiptOK(proofReport, input.UnauthorizedBundle)
	}
	// Final conformance is deliberately calculated last. A PASS therefore
	// cannot survive an unsatisfied indicator, proof, binding, or consumer gate.
	report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "FAIL_CLOSED", "LOWER_RESOLUTION", "BUNDLE_CONSUMPTION_NOT_OBSERVED"
	report.ConformanceCoordinate = Coordinate{"CONSUME_BUNDLE", "consumer-recheck", report.ConformanceReason}
	if input.BundleDigest != "" {
		report.ConformanceReason = "CONSUMER_RECHECK_NOT_OBSERVED"
		report.ConformanceCoordinate = Coordinate{"CONSUME_BUNDLE", "consumer-recheck", report.ConformanceReason}
		if input.ConsumerReceiptProvided {
			report.ConformanceReason = "PROOF_CARRYING_ARTIFACT_CONTRACT_VIOLATED"
			report.ConformanceCoordinate = Coordinate{"EVALUATE", "final-conformance-gate", report.ConformanceReason}
		}
	}
	if structuralGate && indicatorsOK && proofsOK && consumerGate {
		report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		report.ConformanceCoordinate = Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", report.ConformanceReason}
		report.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
	}
	report.ProofSummary = proofSummary(report.Proofs, proofPhase, report.ArtifactUseAuthority)
	report.Digest = reportDigest(report)
	return report
}

func allIndicatorsSatisfied(indicators []Indicator) bool {
	if len(indicators) != len(MetricIDs()) {
		return false
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}

func allProofsPassed(proofs []Proof) bool {
	if len(proofs) != 3 {
		return false
	}
	for _, proof := range proofs {
		if proof.Phase != ProofPhaseFinal || proof.State != ProofStateDischarged || !proof.EvidenceValidated || !proof.Passed || proof.ConsumerGateOpen {
			return false
		}
	}
	return true
}

func validCase(cases []CaseResult) *CaseResult {
	for index := range cases {
		if cases[index].ID == "valid-proof-carrying-artifact" && cases[index].ObservedDecision == "PASS" {
			return &cases[index]
		}
	}
	return nil
}

func unauthorizedConsumerObservation(input Input, phase string) observation {
	artifact, err := decodeStrict[Artifact](input.ValidArtifact)
	if err != nil {
		return failure("INVARIANT_ONLY", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", "CONSUME_BUNDLE", "attestation")
	}
	if input.UnauthorizedBundleError != "" || input.UnauthorizedBundle.Digest == "" || len(input.UnauthorizedConsumer) == 0 {
		result := failure("LOWER_RESOLUTION", "BUNDLE_CONSUMPTION_NOT_OBSERVED", "CONSUME_BUNDLE", "read-bundle")
		result.ConsumerErrorClass = string(ConsumerErrorBundleInvalid)
		result.ConsumerErrorDigest = consumerError(ConsumerErrorBundleInvalid, input.UnauthorizedBundleError).Digest()
		return result
	}
	base := verifyArtifact(input.ValidArtifact, input.Source, input.Operation, input.Recipe, input.HeadSHA, phase)
	if base.Decision != "PASS" {
		return base
	}
	var unauthorized Report
	if err := decodeInto(input.UnauthorizedConsumer, &unauthorized); err != nil {
		result := failure("INVARIANT_ONLY", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", "CONSUME_BUNDLE", "attestation")
		result.ConsumerErrorClass = string(ConsumerErrorAttestationMismatch)
		errorValue := consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
		result.ConsumerErrorDigest = errorValue.Digest()
		return result
	}
	targetDigest, outputDigest, outputExists := bundleTargetDigests(input.UnauthorizedBundle, "artifact.json")
	consumedReceipt, consumeErr := ConsumeBundle(input.UnauthorizedBundle, unauthorized, "artifact.json")
	if consumeErr == nil {
		result := observedFailure(artifact, "INVARIANT_ONLY", "UNAUTHORIZED_CONSUMER_ACCEPTED", "CONSUME_BUNDLE", "consumer",
			map[string]string{"source-bytes-bound": "DISCHARGED", "operation-receipt-bound": "DISCHARGED", "no-byte-authority": "DISCHARGED", "recipe-match": "DISCHARGED", "consumer-authority": "REFUTED"},
			base.ArtifactDigest, base.SourceDigest, base.SemanticDigest, base.OperationDigest)
		result.ConsumerTargetDigest, result.ConsumerOutputDigest, result.ConsumerOutputExists = consumedReceipt.TargetDigest, consumedReceipt.OutputDigest, consumedReceipt.OutputExists
		return result
	}
	consumerFailure := consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
	var typed *ConsumerError
	if errors.As(consumeErr, &typed) {
		consumerFailure = typed
	}
	resultReason, resultResolution, resultStage, resultStep := "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", "INVARIANT_ONLY", "CONSUME_BUNDLE", "attestation"
	states := map[string]string{"source-bytes-bound": "DISCHARGED", "operation-receipt-bound": "DISCHARGED", "no-byte-authority": "DISCHARGED", "recipe-match": "DISCHARGED", "consumer-authority": "REFUTED"}
	if consumerFailure.Class != ConsumerErrorAttestationMismatch {
		resultReason, resultResolution, resultStage, resultStep = "BUNDLE_CONSUMPTION_NOT_OBSERVED", "LOWER_RESOLUTION", "CONSUME_BUNDLE", "consumer-recheck"
		states["consumer-authority"] = "OPEN"
	}
	result := observedFailure(artifact, resultResolution, resultReason, resultStage, resultStep,
		states,
		base.ArtifactDigest, base.SourceDigest, base.SemanticDigest, base.OperationDigest)
	result.ConsumerTargetDigest, result.ConsumerOutputDigest, result.ConsumerOutputExists = targetDigest, outputDigest, outputExists
	result.ConsumerErrorClass, result.ConsumerErrorDigest = string(consumerFailure.Class), consumerFailure.Digest()
	return result
}

func fixedCounterexamples(input Input, report Report) []Counterexample {
	bundleMissing := consumerError(ConsumerErrorBundleInvalid, "bundle was not provided")
	bundleCorrupt := consumerError(ConsumerErrorBundleInvalid, "bundle digest or file content is invalid")
	targetMissing := consumerError(ConsumerErrorTargetMissing, "consumer target is absent from bundle: missing-target.json")
	outputAbsent := consumerError(ConsumerErrorReceiptMismatch, "consumer receipt declares output_exists=false")
	proofFalse := digestValue(struct {
		Stage  string `json:"stage"`
		Step   string `json:"step"`
		Reason string `json:"reason"`
	}{"VERIFY_PROOF", "final-gate", "PROOF_NOT_SATISFIED"})
	indicatorShortfall := digestValue(struct {
		Value  int `json:"value"`
		Target int `json:"target"`
	}{38, 40})
	unauthorizedClass := report.UnauthorizedConsumerErrorClass
	unauthorizedDigest := report.UnauthorizedConsumerErrorDigest
	unauthorizedTarget := report.UnauthorizedConsumerTargetDigest
	if unauthorizedClass == "" {
		unauthorizedClass = string(ConsumerErrorAttestationMismatch)
		unauthorizedDigest = consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified").Digest()
	}
	return []Counterexample{
		{ID: "bundle-not-provided", Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Coordinate: Coordinate{"CONSUME_BUNDLE", "read-bundle", "BUNDLE_CONSUMPTION_NOT_OBSERVED"}, ClaimID: "consumer-authority", From: "CARRIED", To: "OPEN", ErrorClass: string(bundleMissing.Class), ErrorDigest: bundleMissing.Digest(), OutputExists: false},
		{ID: "bundle-corrupt", Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Coordinate: Coordinate{"CONSUME_BUNDLE", "validate-bundle", "BUNDLE_INVALID"}, ClaimID: "consumer-authority", From: "CARRIED", To: "OPEN", ErrorClass: string(bundleCorrupt.Class), ErrorDigest: bundleCorrupt.Digest(), OutputExists: false},
		{ID: "unauthorized-attestation-mismatch", Decision: "FAIL_CLOSED", Resolution: "INVARIANT_ONLY", Coordinate: Coordinate{"CONSUME_BUNDLE", "attestation", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED"}, ClaimID: "consumer-authority", From: "CARRIED", To: "REFUTED", ErrorClass: unauthorizedClass, ErrorDigest: unauthorizedDigest, TargetDigest: unauthorizedTarget, OutputExists: false},
		{ID: "consumer-target-missing", Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Coordinate: Coordinate{"CONSUME_BUNDLE", "target-missing", "TARGET_MISSING"}, ClaimID: "consumer-authority", From: "CARRIED", To: "OPEN", ErrorClass: string(targetMissing.Class), ErrorDigest: targetMissing.Digest(), OutputExists: false},
		{ID: "consumer-output-absent", Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Coordinate: Coordinate{"CONSUME_BUNDLE", "receipt", "CONSUMER_OUTPUT_ABSENT"}, ClaimID: "consumer-authority", From: "CARRIED", To: "OPEN", ErrorClass: string(outputAbsent.Class), ErrorDigest: outputAbsent.Digest(), OutputExists: false},
		{ID: "proof-false", Decision: "FAIL_CLOSED", Resolution: "INVARIANT_ONLY", Coordinate: Coordinate{"VERIFY_PROOF", "final-gate", "PROOF_NOT_SATISFIED"}, ClaimID: "proof-gate", From: "CARRIED", To: "OPEN", ErrorClass: "PROOF_GATE", ErrorDigest: proofFalse, OutputExists: false},
		{ID: "main-indicator-38-of-40", Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Coordinate: Coordinate{"EVALUATE", "final-conformance-gate", "INDICATOR_GATE_NOT_SATISFIED"}, ClaimID: "conformance", From: "CARRIED", To: "OPEN", ErrorClass: "INDICATOR_GATE", ErrorDigest: indicatorShortfall, OutputExists: false},
	}
}

func caseInput(input Input, kind string) ([]byte, []byte, []byte, []byte) {
	switch kind {
	case "VALID":
		return input.ValidArtifact, input.Source, input.Operation, input.Recipe
	case "TAMPERED":
		return input.TamperedArtifact, input.Source, input.Operation, input.Recipe
	case "COHERENT_TAMPER":
		return input.CoherentTamperedArtifact, input.Source, input.CoherentOperation, input.Recipe
	case "MISSING":
		return input.MissingArtifact, input.Source, input.Operation, input.Recipe
	case "BYTE_ONLY":
		return input.ByteOnlyArtifact, nil, nil, nil
	case "WRONG_RECIPE":
		return input.ValidArtifact, input.Source, input.Operation, input.WrongRecipe
	case "RECIPE_ONLY":
		artifact := input.RecipeOnlyArtifact
		if len(artifact) == 0 {
			artifact = input.ValidArtifact
		}
		return artifact, input.Source, input.Operation, input.WrongRecipe
	case "MISSING_ATTACHMENT":
		return input.ValidArtifact, input.Source, nil, input.Recipe
	case "WRONG_ATTACHMENT_DIGEST":
		return input.ValidArtifact, input.Source, input.WrongAttachmentDigest, input.Recipe
	case "UNRELATED_TAMPER":
		return input.UnrelatedTamperedArtifact, input.Source, input.Operation, input.Recipe
	case "STALE_HEAD":
		return input.StaleHeadArtifact, input.Source, input.Operation, input.Recipe
	case "CLAIM_PROPOSITION_TAMPER":
		return input.ClaimPropositionArtifact, input.Source, input.Operation, input.Recipe
	case "CLAIM_DEPENDENCY_TAMPER":
		return input.ClaimDependencyArtifact, input.Source, input.Operation, input.Recipe
	case "CLAIM_PROOF_CHOICE_TAMPER":
		return input.ClaimProofChoiceArtifact, input.Source, input.Operation, input.Recipe
	case "CLAIM_TARGET_TAMPER":
		return input.ClaimTargetArtifact, input.Source, input.Operation, input.Recipe
	default:
		return nil, nil, nil, nil
	}
}

func validCheckout(input Input, contract Contract) bool {
	checkout := input.Checkout
	if checkout.HeadSHA != input.HeadSHA || checkout.ActualHeadSHA != input.HeadSHA || !validHead(checkout.HeadSHA) || !validDigest(checkout.TreeDigest) ||
		len(input.ContractBytes) == 0 || checkout.SourceDigest != digestBytes(input.Source) || checkout.OperationDigest != digestBytes(input.Operation) ||
		checkout.RecipeDigest != digestBytes(input.Recipe) || checkout.ContractDigest != digestBytes(input.ContractBytes) || digestValue(contract) != digestValue(CanonicalContract()) {
		return false
	}
	return input.BundleDigest == "" || validDigest(input.BundleDigest)
}

func failedReport(input Input, reason string) Report {
	canonicalRecipe := CanonicalRecipe()
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Producer: ProducerID, Consumer: ConsumerID,
		ConformanceDecision: "FAIL_CLOSED", ConformanceResolution: "LOWER_RESOLUTION", ConformanceReason: reason,
		ContractDigest: digestValue(input.Contract), RecipeDigest: digestValue(canonicalRecipe), RecipeVersion: canonicalRecipe.Version,
		IndependenceDigest: digestValue(input.Independence), Cases: []CaseResult{}, Summary: Summary{CasesTotal: CaseTotal},
		Indicators: []Indicator{}, Proofs: []Proof{}, Transitions: []ClaimTransition{}, WriteSet: input.WriteSet, Checkout: input.Checkout, BundleDigest: input.BundleDigest,
		NetChangedPaths: input.WriteSet.NetChangedPaths, CapabilityMutationGranted: input.WriteSet.CapabilityMutationGranted, PromotionAuthority: false, SemanticAuthority: false,
		Interventions: []InterventionResult{}, AuthorityObservation: "UNKNOWN_GLOBAL_TRANSIENT_SCOPE",
		NotClaimed: []string{"authority without the canonical contract and independent verifier"}}
	if input.BundleDigest == "" {
		report.CheckoutBindingScope = "CURRENT_CHECKOUT_OBSERVATION"
	} else {
		report.CheckoutBindingScope = "BUNDLE_HISTORICAL_SUBJECT_BINDING"
	}
	report.Digest = reportDigest(report)
	return report
}
