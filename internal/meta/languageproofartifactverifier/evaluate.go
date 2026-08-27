package languageproofartifactverifier

import "reflect"

func Evaluate(input Input) Report {
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Producer: ProducerID, Consumer: ConsumerID,
		ContractDigest: digestValue(input.Contract), RecipeDigest: digestValue(CanonicalRecipe()),
		IndependenceDigest: digestValue(input.Independence), NotClaimed: []string{
			"full compiler semantic correctness", "cryptographic signer authenticity", "external side effects", "authority from artifact bytes alone"}}
	if !reflect.DeepEqual(input.Contract, CanonicalContract()) || !validHead(input.HeadSHA) ||
		input.Independence.Schema != "gooo/language-proof-carrying-artifact-independence/v1" ||
		input.Independence.ProducerDependencies != 0 {
		return failedReport(input, "PROOF_CARRYING_ARTIFACT_CONTRACT_OR_INDEPENDENCE_MISMATCH")
	}
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
			OperationDigest: observed.OperationDigest})
	}
	summary := summarize(results, input.Independence)
	report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_CONTRACT_VIOLATED"
	if summary.CasesSatisfied == CaseTotal && summary.ValidArtifacts == 1 && summary.PreservedTransitions == TransitionTotal &&
		summary.TamperedRejections == 1 && summary.MissingEvidenceRejections == 1 && summary.ByteOnlyDenials == 1 &&
		summary.RecipeRejections == 1 && summary.ProducerDependencies == 0 && summary.GeneratedAuthority == 0 {
		report.Decision, report.Resolution, report.Reason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		report.AuthorityGranted = true
	}
	report.Cases = results
	report.Summary = summary
	report.Indicators = indicators(summary)
	report.Proofs = proofs(summary)
	report.Transitions = transitions(results)
	report.RepositoryWrites = 0
	report.MutationAuthority = false
	report.Digest = reportDigest(report)
	return report
}

func caseInput(input Input, kind string) ([]byte, []byte, []byte, []byte) {
	switch kind {
	case "VALID":
		return input.ValidArtifact, input.Source, input.Operation, input.Recipe
	case "TAMPERED":
		return input.TamperedArtifact, input.Source, input.Operation, input.Recipe
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
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Reason: reason,
		ContractDigest: digestValue(input.Contract), RecipeDigest: digestValue(CanonicalRecipe()),
		IndependenceDigest: digestValue(input.Independence), Cases: []CaseResult{}, Summary: Summary{CasesTotal: CaseTotal},
		Indicators: []Indicator{}, Proofs: []Proof{}, Transitions: []ClaimTransition{}, RepositoryWrites: 0, MutationAuthority: false,
		NotClaimed: []string{"authority without the canonical contract and independent verifier"}}
	report.Digest = reportDigest(report)
	return report
}
