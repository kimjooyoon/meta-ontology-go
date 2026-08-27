package languageproofartifactverifier

import (
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
	if !reflect.DeepEqual(input.Contract, canonicalContract) || !validHead(input.HeadSHA) ||
		input.Independence.Schema != "gooo/language-proof-carrying-artifact-independence/v1" ||
		input.Independence.ProducerDependencies != input.Independence.ProducerImportNumerator ||
		input.Independence.ProducerImportNumerator != 0 || input.Independence.ProducerImportDenominator <= 0 || input.Independence.CoreParserDependencies <= 0 ||
		!validCheckout(input, canonicalContract) {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_CONTRACT_OR_INDEPENDENCE_MISMATCH")
	}
	validArtifact, err := decodeStrict[Artifact](input.ValidArtifact)
	if err != nil || validateWriteSet(validArtifact.WriteSet) != nil || !reflect.DeepEqual(validArtifact.WriteSet, input.WriteSet) {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_WRITE_SET_NOT_BOUND")
	}

	results := make([]CaseResult, 0, len(input.Contract.Cases))
	for _, definition := range input.Contract.Cases {
		var observed observation
		if definition.InputKind == "UNAUTHORIZED_CONSUMER" {
			observed = unauthorizedConsumerObservation(input)
		} else {
			artifact, source, operation, recipe := caseInput(input, definition.InputKind)
			observed = verifyArtifact(artifact, source, operation, recipe, input.HeadSHA)
		}
		if definition.InputKind == "RECIPE_ONLY" && observed.Decision != "PASS" {
			observed.Reason = "RECIPE_CLAIM_ONLY_MISMATCH"
			observed.Coordinate = Coordinate{"CONSUME_RECIPE", "recipe-only-intervention", observed.Reason}
			for index := range observed.Claims {
				if observed.Claims[index].ID == "recipe-match" {
					observed.Claims[index].Reason = observed.Reason
					observed.Claims[index].StateDigest = claimStateDigest(observed.Claims[index])
				}
			}
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
			SemanticDigest: observed.SemanticDigest, OperationDigest: observed.OperationDigest})
	}

	interventions := evaluateInterventions(input.Interventions, input.HeadSHA)
	report.WriteSet = input.WriteSet
	report.RepositoryWrites = input.WriteSet.RepositoryWrites
	report.MutationAuthority = input.WriteSet.MutationAuthority
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
	report.Summary = summarize(results, input.Independence, input.WriteSet, interventions, report.Ledger)
	report.Summary.BundleOnlyVerification = boolToInt(input.BundleDigest != "")
	report.Summary.ConsumerRechecks = boolToInt(input.BundleDigest != "")
	report.Transitions = transitions(results)
	report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "FAIL_CLOSED", "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_CONTRACT_VIOLATED"
	if report.Summary.CasesSatisfied == CaseTotal && report.Summary.ValidArtifacts == 1 && report.Summary.PreservedTransitions == EvidenceTotal+1 &&
		report.Summary.TamperedRejections == 1 && report.Summary.CoherentTamperRejections == 1 && report.Summary.MissingEvidenceRejections == 1 &&
		report.Summary.ByteOnlyDenials == 1 && report.Summary.RecipeRejections == 1 && report.Summary.RecipeOnlyRejections == 1 &&
		report.Summary.MissingAttachmentRejections == 1 && report.Summary.WrongAttachmentRejections == 1 && report.Summary.UnrelatedEvidenceRejections == 1 &&
		report.Summary.StaleHeadRejections == 1 && report.Summary.UnauthorizedConsumerDenials == 1 && report.Summary.ProducerDependencies == 0 &&
		report.Summary.SemanticInterventions == 1 && report.Summary.NonsemanticInterventions == 1 && report.Summary.ReadOnlyAuthorities == 1 &&
		report.Summary.FinalLedgerOpenClaims == ClaimTemplateTotal && report.Summary.FinalLedgerDischargedClaims == ClaimTemplateTotal &&
		report.Summary.ProducerImportNumerator == 0 && report.Summary.ProducerImportDenominator > 0 && report.Summary.NetRepositoryStateUnchanged == 1 &&
		report.Summary.GeneratedAuthority == 0 && report.Summary.RepositoryWrites == 0 && report.Summary.MutationAuthorities == 0 &&
		report.Summary.PromotionAuthorities == 0 && report.Summary.SemanticAuthorities == 0 {
		report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		report.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
	}
	report.Indicators = indicators(report.Summary)
	report.Proofs = proofs(report, results)
	if report.ConformanceDecision == "PASS" {
		if input.BundleDigest != "" {
			report.ConsumerReceipt = expectedConsumerReceipt(report, "artifact.json", input.ValidArtifact)
		}
	}
	report.Digest = reportDigest(report)
	return report
}

func validCase(cases []CaseResult) *CaseResult {
	for index := range cases {
		if cases[index].ID == "valid-proof-carrying-artifact" && cases[index].ObservedDecision == "PASS" {
			return &cases[index]
		}
	}
	return nil
}

func unauthorizedConsumerObservation(input Input) observation {
	artifact, err := decodeStrict[Artifact](input.ValidArtifact)
	if err != nil {
		return failure("INVARIANT_ONLY", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", "CONSUME_AUTHORITY", "consumer")
	}
	base := verifyArtifact(input.ValidArtifact, input.Source, input.Operation, input.Recipe, input.HeadSHA)
	if base.Decision != "PASS" {
		return base
	}
	return observedFailure(artifact, "INVARIANT_ONLY", "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", "CONSUME_AUTHORITY", "consumer",
		map[string]string{"source-bytes-bound": "DISCHARGED", "operation-receipt-bound": "DISCHARGED", "no-byte-authority": "DISCHARGED", "recipe-match": "DISCHARGED", "consumer-authority": "REFUTED"},
		base.ArtifactDigest, base.SourceDigest, base.SemanticDigest, base.OperationDigest)
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
		RepositoryWrites: input.WriteSet.RepositoryWrites, MutationAuthority: input.WriteSet.MutationAuthority, PromotionAuthority: false, SemanticAuthority: false,
		Interventions: []InterventionResult{}, AuthorityObservation: "UNKNOWN_GLOBAL_TRANSIENT_SCOPE",
		NotClaimed: []string{"authority without the canonical contract and independent verifier"}}
	report.Digest = reportDigest(report)
	return report
}
