package languageproofartifactverifier

import "reflect"

func Evaluate(input Input) Report {
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Producer: ProducerID, Consumer: ConsumerID,
		ContractDigest: digestValue(input.Contract), RecipeDigest: digestValue(CanonicalRecipe()), RecipeVersion: CanonicalRecipe().Version,
		IndependenceDigest: digestValue(input.Independence), NotClaimed: []string{
			"full compiler semantic correctness", "cryptographic signer authenticity", "external side effects", "authority from artifact bytes alone"}}
	if !reflect.DeepEqual(input.Contract, CanonicalContract()) || !validHead(input.HeadSHA) ||
		input.Independence.Schema != "gooo/language-proof-carrying-artifact-independence/v1" ||
		input.Independence.ProducerDependencies != input.Independence.ProducerImportNumerator ||
		input.Independence.ProducerImportNumerator != 0 || input.Independence.ProducerImportDenominator <= 0 ||
		input.Independence.CoreParserDependencies <= 0 {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_CONTRACT_OR_INDEPENDENCE_MISMATCH")
	}
	validArtifact, err := decodeStrict[Artifact](input.ValidArtifact)
	if err != nil || validateWriteSet(validArtifact.WriteSet) != nil || !reflect.DeepEqual(validArtifact.WriteSet, input.WriteSet) {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_WRITE_SET_NOT_BOUND")
	}
	// Contract cases are expectations used to measure conformance. They never
	// supply the subject artifact decision or authorize consumption.
	definitions := input.Contract.Cases
	results := make([]CaseResult, 0, len(definitions))
	for _, definition := range definitions {
		artifact, source, operation, recipe := caseInput(input, definition.InputKind)
		observed := verifyArtifact(artifact, source, operation, recipe, input.HeadSHA)
		status := "NOT_SATISFIED"
		if observed.Decision == definition.ExpectedDecision && observed.Resolution == definition.ExpectedResolution && observed.Reason == definition.ExpectedReason {
			status = "SATISFIED"
		}
		results = append(results, CaseResult{ID: definition.ID, Status: status,
			ExpectedDecision: definition.ExpectedDecision, ExpectedResolution: definition.ExpectedResolution,
			ExpectedReason: definition.ExpectedReason, ObservedDecision: observed.Decision,
			ObservedResolution: observed.Resolution, ObservedReason: observed.Reason,
			ProofChoice: definition.ProofChoice, MetaOperation: definition.MetaOperation, Coordinate: observed.Coordinate,
			Claims: observed.Claims, ArtifactDigest: observed.ArtifactDigest, SourceDigest: observed.SourceDigest,
			SemanticDigest: observed.SemanticDigest, OperationDigest: observed.OperationDigest})
	}
	interventions := evaluateInterventions(input.Interventions, input.HeadSHA)
	summary := summarize(results, input.Independence, input.WriteSet, interventions)
	report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "FAIL_CLOSED", "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_CONTRACT_VIOLATED"
	report.WriteSet = input.WriteSet
	report.RepositoryWrites = input.WriteSet.RepositoryWrites
	report.MutationAuthority = input.WriteSet.MutationAuthority
	report.PromotionAuthority = false
	report.SemanticAuthority = false
	report.Interventions = interventions
	if artifact, err := decodeStrict[Artifact](input.ValidArtifact); err == nil {
		report.SubjectArtifactDecision, report.SubjectArtifactResolution, report.SubjectArtifactReason = artifact.Decision, artifact.Resolution, artifact.Reason
		report.PriorLedger = artifact.PriorLedger
		report.Ledger = dischargeLedger(artifact.PriorLedger, results)
	}
	if summary.CasesSatisfied == CaseTotal && summary.ValidArtifacts == 1 && summary.PreservedTransitions == EvidenceTotal &&
		summary.TamperedRejections == 1 && summary.CoherentTamperRejections == 1 && summary.MissingEvidenceRejections == 1 &&
		summary.ByteOnlyDenials == 1 && summary.RecipeRejections == 1 && summary.ProducerDependencies == 0 &&
		summary.SemanticInterventions == 1 && summary.NonsemanticInterventions == 1 && summary.ReadOnlyAuthorities == 1 &&
		summary.LedgerDischargedClaims == 3 && summary.LedgerOpenClaims == 6 && summary.LedgerRefutedClaims == 9 &&
		summary.ProducerImportNumerator == 0 && summary.ProducerImportDenominator > 0 &&
		summary.GeneratedAuthority == 0 && summary.RepositoryWrites == 0 && summary.MutationAuthorities == 0 &&
		summary.PromotionAuthorities == 0 && summary.SemanticAuthorities == 0 {
		report.ConformanceDecision, report.ConformanceResolution, report.ConformanceReason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		report.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
	}
	report.Cases = results
	report.Summary = summary
	report.Indicators = indicators(summary)
	report.Proofs = proofs(summary)
	report.Transitions = transitions(results)
	report.Digest = reportDigest(report)
	return report
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
	default:
		return nil, nil, nil, nil
	}
}

func failedReport(input Input, reason string) Report {
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, ConformanceDecision: "FAIL_CLOSED", ConformanceResolution: "LOWER_RESOLUTION", ConformanceReason: reason,
		ContractDigest: digestValue(input.Contract), RecipeDigest: digestValue(CanonicalRecipe()), RecipeVersion: CanonicalRecipe().Version,
		IndependenceDigest: digestValue(input.Independence), Cases: []CaseResult{}, Summary: Summary{CasesTotal: CaseTotal},
		Indicators: []Indicator{}, Proofs: []Proof{}, Transitions: []ClaimTransition{}, WriteSet: input.WriteSet,
		RepositoryWrites: input.WriteSet.RepositoryWrites, MutationAuthority: input.WriteSet.MutationAuthority,
		PromotionAuthority: false, SemanticAuthority: false, Interventions: []InterventionResult{},
		NotClaimed: []string{"authority without the canonical contract and independent verifier"}}
	report.Digest = reportDigest(report)
	return report
}
