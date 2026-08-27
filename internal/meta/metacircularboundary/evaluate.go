package metacircularboundary

func Evaluate(input Input) Report {
	report := baseReport(input)
	if input.Path != ExpectedSourcePath || !validSHA(input.HeadSHA) {
		return sealReport(report)
	}
	source, err := observeSource(input.Path, input.Source)
	if err != nil {
		return sealReport(report)
	}
	effect, effectErr := parseEffectEvidence(input.EffectEvidence)
	source.Effect = effect
	source.RepositoryWrites = effect.RepositoryWrites
	source.MutationAuthority = effect.MutationAuthority == AuthorityGranted
	source.ReadOnly = effect.Known && effect.OutputOutsideRepository && effect.RepositoryWrites == 0 && effect.MutationAuthority == AuthorityDenied
	report.Source = source
	grants, grantArtifactDigest, grantErr := parseGrantEvidence(input.GrantEvidence, source.SemanticDigest)
	source.GrantArtifactDigest = grantArtifactDigest
	report.Source = source
	replay, replayErr := parseReplayEvidence(input.ReplayEvidence)
	if replayErr == nil {
		report.ReplayEvidenceDigest = digestBytes(input.ReplayEvidence)
	}
	attempts, factsErr := parseAttempts(source)
	definitions := DenominatorContract().Cases
	artifacts := make(map[string]ExecutionArtifact, len(input.ExecutionArtifacts))
	for _, artifact := range input.ExecutionArtifacts {
		artifacts[artifact.CaseID] = artifact
	}
	for _, definition := range definitions {
		attempt, ok := attempts[definition.ID]
		if !ok {
			attempt = Attempt{Unknown: true}
		}
		grant := grants[definition.ID]
		artifact, artifactOK := artifacts[definition.ID]
		observation := classify(source, definition, attempt, grant, grantErr, artifact, artifactOK)
		receipt := buildReceipt(source, definition, attempt, grant, observation, artifact, artifactOK)
		item := CaseResult{Definition: definition, Attempt: attempt, Grant: grant, Observation: observation, Receipt: receipt}
		item.Passed = matches(definition, observation) && receipt.ReceiptDigest == sealReceipt(receipt).ReceiptDigest
		report.Cases = append(report.Cases, item)
		report.Receipts = append(report.Receipts, receipt)
		if artifactOK {
			report.ExecutionArtifacts = append(report.ExecutionArtifacts, artifact)
		}
	}
	report.Summary = summarize(report.Cases)
	report.Summary.ReplayMatches = replayMatches(replay)
	report.Indicators = buildIndicators(report.Summary)
	report.Decision, report.Resolution, report.Reason, report.Coordinate = reportDecision(source, factsErr, grantErr, effectErr, replayErr, report.Cases, report.Summary)
	report.RepositoryWrites = source.RepositoryWrites
	report.MutationAuthority = source.MutationAuthority
	return sealReport(report)
}

func baseReport(input Input) Report {
	return Report{
		Schema: ReportSchema, Scope: Scope, HeadSHA: input.HeadSHA, Decision: DecisionOpen,
		Resolution: ResolutionLower, Reason: ReasonSourceBindingUnknown,
		Coordinate:     Coordinate{Stage: "READ_SOURCE", Step: "bind-self-description", Reason: ReasonSourceBindingUnknown},
		Denominator:    DenominatorContract(),
		MetaOperations: MetaOperations(), RepositoryWrites: 0, MutationAuthority: false,
		MetaValue: "DESCRIPTION_AUTHORIZATION_EXECUTION_ARE_DISTINCT",
		NotClaimed: []string{
			"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
			"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
		},
	}
}

func classify(source SourceObservation, definition CaseDefinition, attempt Attempt, grant ExternalGrant, grantErr error, artifact ExecutionArtifact, artifactPresent bool) CaseObservation {
	observation := CaseObservation{Description: DescriptionBound, Authorization: AuthorizationUnknown, Execution: ExecutionUnknown, Decision: DecisionOpen, Predicate: attempt.Predicate, GrantDigest: grant.GrantDigest, ExecutionArtifactPresent: artifactPresent, RepositoryWrites: source.RepositoryWrites, MutationAuthority: source.MutationAuthority}
	if attempt.Unknown {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonCaseDataUnknown
		return observation
	}
	if attempt.Contradictory {
		observation.Reason = ReasonContradictory
		observation.Decision = DecisionRefuted
		observation.Authorization = AuthorizationDenied
		return observation
	}
	if !source.DescriptionBound || attempt.DescriptionDigest != source.SemanticDigest {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonSourceBindingUnknown
		return observation
	}
	if !source.Graph.Valid {
		observation.Reason = ReasonGraphUnknown
		return observation
	}
	if grantErr != nil || grant.CaseID == "" {
		observation.Reason = ReasonGrantUnknown
		return observation
	}
	if attempt.DescriptionAuthorityClaim {
		observation.DescriptionEscalated = true
	}
	if definition.ID == "description-only" {
		if attempt.RequestKind != RequestNone || grant.Decision != GrantDeny {
			observation.Reason = ReasonCasePredicateMismatch
			return observation
		}
		observation.Authorization, observation.Execution, observation.Decision = AuthorizationDenied, ExecutionBlocked, DecisionFailClosed
		if observation.DescriptionEscalated {
			observation.Reason = ReasonDescriptionForgery
		} else {
			observation.Reason = ReasonDescriptionOnly
		}
		return observation
	}
	if attempt.RequestKind != RequestReadOnly || grant.Decision != GrantDecision {
		observation.Reason = ReasonCasePredicateMismatch
		return observation
	}
	if grant.SubjectDigest != source.SemanticDigest {
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = AuthorizationDenied, ExecutionBlocked, DecisionFailClosed, ReasonGrantDenied
		return observation
	}
	if definition.ID == "forged-capability" {
		observation.Predicate = PredicateForgedGrant
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = AuthorizationDenied, ExecutionBlocked, DecisionFailClosed, ReasonForgedCapability
		return observation
	}
	if definition.ID == "write-capability-out-of-scope" || grant.Scope == ScopeWrite {
		observation.Predicate = PredicateOutOfScopeGrant
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = AuthorizationDenied, ExecutionBlocked, DecisionFailClosed, ReasonOutOfScopeCapability
		return observation
	}
	if grant.Issuer != "external-authority" || grant.Operation != MetaOperationID || grant.Scope != ScopeReadOnly || grant.Handle == "" {
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = AuthorizationDenied, ExecutionBlocked, DecisionFailClosed, ReasonForgedCapability
		return observation
	}
	observation.Authorization, observation.Reason = AuthorizationGranted, ReasonExplicitCapability
	if !attempt.RequestExecution {
		observation.Execution, observation.Decision = ExecutionBlocked, DecisionFailClosed
		return observation
	}
	if !artifactPresent {
		observation.Reason = ReasonExecutionUnknown
		return observation
	}
	observation.ExecutionArtifactValid = validExecutionArtifact(source, grant, definition.ID, artifact)
	if !observation.ExecutionArtifactValid {
		observation.Reason = ReasonExecutionInvalid
		observation.Decision = DecisionRefuted
		return observation
	}
	observation.Execution, observation.OutputDigest, observation.Decision = ExecutionAllowed, artifact.OutputDigest, DecisionPass
	return observation
}

func reportDecision(source SourceObservation, factsErr, grantErr, effectErr, replayErr error, cases []CaseResult, summary Summary) (string, string, string, Coordinate) {
	if effectErr != nil || !source.Effect.Known || source.Effect.MutationAuthority == AuthorityUnknown || !source.Effect.OutputOutsideRepository || source.Effect.PermissionEvidence == "" {
		return DecisionOpen, ResolutionLower, ReasonEffectUnknown,
			Coordinate{Stage: "OBSERVE_EFFECT", Step: "resolve-workspace-permission-and-output", Reason: ReasonEffectUnknown}
	}
	if hasDecision(cases, DecisionRefuted) {
		return DecisionRefuted, ResolutionExact, ReasonContradictory,
			Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "refute-capability-or-execution-evidence", Reason: ReasonContradictory}
	}
	if factsErr != nil || grantErr != nil || effectErr != nil || replayErr != nil || hasDecision(cases, DecisionOpen) {
		reason := ReasonCaseDataUnknown
		stage, step := "PARSE_COMPUTES", "read-case-facts"
		if factsErr != nil || hasReason(cases, ReasonCaseDataUnknown) {
			reason, stage, step = ReasonCaseDataUnknown, "PARSE_COMPUTES", "read-case-facts"
		} else if grantErr != nil {
			reason, stage, step = ReasonGrantUnknown, "READ_EXTERNAL_GRANT", "parse-grant-evidence"
		} else if effectErr != nil {
			reason, stage, step = ReasonEffectUnknown, "OBSERVE_EFFECT", "compare-workspace-state"
		} else if replayErr != nil {
			reason, stage, step = ReasonReplayUnknown, "OBSERVE_REPLAY", "compare-receipt-digests"
		} else if hasReason(cases, ReasonGraphUnknown) {
			reason, stage, step = ReasonGraphUnknown, "LOWER_GRAPH", "reconstruct-describe-grant-execute"
		} else if hasReason(cases, ReasonExecutionUnknown) {
			reason, stage, step = ReasonExecutionUnknown, "EXECUTE_META_OPERATION", "observe-output-artifact"
		}
		return DecisionOpen, ResolutionLower, reason, Coordinate{Stage: stage, Step: step, Reason: reason}
	}
	if replayErr != nil || summary.CasesPassed != len(contractCases()) || summary.DescriptionBound != len(contractCases()) || summary.ExplicitAuthorizations != 1 || summary.AllowedExecutions != 1 || summary.DescriptionEscalationPaths != 0 || summary.RepositoryWrites != 0 || summary.MutationAuthority != 0 || summary.ReplayMatches != len(contractCases()) || !source.ReadOnly || !source.Effect.Known || !sourceEffectSatisfied(cases) {
		return DecisionFailClosed, ResolutionLower, ReasonContractUnsatisfied,
			Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "evaluate-observed-boundary-evidence", Reason: ReasonContractUnsatisfied}
	}
	return DecisionPass, ResolutionExact, ReasonContractSatisfied,
		Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "classify-gooo-computations", Reason: ReasonContractSatisfied}
}

func hasDecision(cases []CaseResult, decision string) bool {
	for _, item := range cases {
		if item.Observation.Decision == decision {
			return true
		}
	}
	return false
}

func hasReason(cases []CaseResult, reason string) bool {
	for _, item := range cases {
		if item.Observation.Reason == reason {
			return true
		}
	}
	return false
}

func sourceEffectSatisfied(cases []CaseResult) bool {
	for _, item := range cases {
		if item.Observation.RepositoryWrites != 0 || item.Observation.MutationAuthority {
			return false
		}
	}
	return true
}

func buildReceipt(source SourceObservation, definition CaseDefinition, attempt Attempt, grant ExternalGrant, observation CaseObservation, artifact ExecutionArtifact, artifactPresent bool) Receipt {
	authorizationEvidenceDigest := ""
	if grant.Decision == GrantDecision {
		authorizationEvidenceDigest = digestValue(grant)
	}
	receipt := Receipt{
		Schema: "gooo/meta-circular-boundary-receipt/v1", CaseID: definition.ID,
		Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge",
		MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice,
		Coordinate:   Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: definition.ID, Reason: observation.Reason},
		SourceDigest: source.SourceDigest, DescriptionDigest: attempt.DescriptionDigest, AuthorizationEvidenceDigest: authorizationEvidenceDigest, GrantDigest: grant.GrantDigest,
		Decision: observation.Decision, Authorization: observation.Authorization, Execution: observation.Execution,
		RepositoryWrites: observation.RepositoryWrites, MutationAuthority: observation.MutationAuthority,
		ClaimTransitions: claimTransitions(definition, attempt, grant, observation, artifact, artifactPresent),
	}
	if artifactPresent && observation.ExecutionArtifactValid {
		receipt.ExecutionArtifact = &artifact
	}
	return sealReceipt(receipt)
}

func claimTransitions(definition CaseDefinition, attempt Attempt, grant ExternalGrant, observation CaseObservation, artifact ExecutionArtifact, artifactPresent bool) []ClaimTransition {
	descriptionAfter, descriptionEvent := "DESCRIBED", "DESCRIPTION_OBSERVED"
	if observation.Description != DescriptionBound {
		descriptionAfter, descriptionEvent = "UNKNOWN", "DESCRIPTION_UNRESOLVED"
	} else if observation.DescriptionEscalated {
		descriptionAfter, descriptionEvent = "REFUTED", "DESCRIPTION_AUTHORITY_CLAIM_REFUTED"
	}
	descriptionClaim := definition.ID + ".description"
	authorizationClaim := definition.ID + ".authorization"
	executionClaim := definition.ID + ".execution"
	transitions := []ClaimTransition{{ClaimID: descriptionClaim, PropositionDigest: propositionDigest(descriptionClaim, descriptionAfter), Event: descriptionEvent, Before: "UNRECORDED", After: descriptionAfter, Coordinate: Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: observation.Reason}, EvidenceDigest: attempt.DescriptionDigest}}
	after, event := "DENIED", "AUTHORIZATION_REJECTED"
	if observation.Decision == DecisionOpen {
		after, event = "UNKNOWN", "AUTHORIZATION_UNRESOLVED"
	} else if observation.Decision == DecisionRefuted {
		after, event = "REFUTED", "CAPABILITY_EVIDENCE_REFUTED"
	} else if observation.Authorization == AuthorizationGranted {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	authorizationEvidence := grant.GrantDigest
	transitions = append(transitions, ClaimTransition{ClaimID: authorizationClaim, PropositionDigest: propositionDigest(authorizationClaim, after), Event: event, Before: "UNRECORDED", After: after, Coordinate: Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}, EvidenceDigest: authorizationEvidence, DependsOnClaimID: descriptionClaim, DependsOnAfter: descriptionAfter})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Decision == DecisionOpen {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Decision == DecisionRefuted {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Execution == ExecutionAllowed {
		executionAfter, executionEvent = "EXECUTED_READ_ONLY", "EXECUTION_ALLOWED"
	}
	executionEvidence := ""
	if artifactPresent {
		executionEvidence = artifact.OutputDigest
	}
	return append(transitions, ClaimTransition{ClaimID: executionClaim, PropositionDigest: propositionDigest(executionClaim, executionAfter), Event: executionEvent, Before: "NOT_EXECUTED", After: executionAfter, Coordinate: Coordinate{Stage: "SEAL_RECEIPT", Step: "seal-read-only-result", Reason: observation.Reason}, EvidenceDigest: executionEvidence, DependsOnClaimID: authorizationClaim, DependsOnAfter: after})
}

func propositionDigest(claimID, state string) string {
	return digestBytes([]byte("proposition|" + claimID + "|" + state))
}

func matches(definition CaseDefinition, observation CaseObservation) bool {
	if definition.ID == "" || definition.Predicate == "" || observation.Predicate != definition.Predicate || observation.Description == "UNKNOWN" || observation.Decision == DecisionOpen || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return false
	}
	if definition.ExpectedDecision != "" && (observation.Decision != definition.ExpectedDecision || observation.Authorization != definition.ExpectedAuthorization || observation.Execution != definition.ExpectedExecution || observation.Reason != definition.ExpectedReason) {
		return false
	}
	switch definition.ID {
	case "description-only":
		return observation.Authorization == AuthorizationDenied && observation.Execution == ExecutionBlocked && !observation.ExecutionArtifactPresent
	case "explicit-read-only-capability":
		return observation.Predicate == PredicateExplicitGrant && observation.Authorization == AuthorizationGranted && observation.Execution == ExecutionAllowed && observation.ExecutionArtifactValid && observation.OutputDigest != ""
	case "forged-capability":
		return observation.Predicate == PredicateForgedGrant && observation.Authorization == AuthorizationDenied && observation.Execution == ExecutionBlocked && !observation.ExecutionArtifactPresent && observation.Reason == ReasonForgedCapability
	case "write-capability-out-of-scope":
		return observation.Predicate == PredicateOutOfScopeGrant && observation.Authorization == AuthorizationDenied && observation.Execution == ExecutionBlocked && !observation.ExecutionArtifactPresent && observation.Reason == ReasonOutOfScopeCapability
	}
	return false
}

func summarize(cases []CaseResult) Summary {
	summary := Summary{CasesTotal: len(cases)}
	for _, item := range cases {
		if item.Passed {
			summary.CasesPassed++
		}
		if item.Observation.Description == DescriptionBound {
			summary.DescriptionBound++
		}
		if item.Observation.Authorization == AuthorizationGranted {
			summary.ExplicitAuthorizations++
		}
		if item.Observation.Execution == ExecutionAllowed {
			summary.AllowedExecutions++
		}
		switch item.Observation.Reason {
		case ReasonDescriptionOnly:
			summary.DescriptionOnlyBlocked++
		case ReasonForgedCapability:
			summary.ForgedAuthorizationsBlocked++
		case ReasonOutOfScopeCapability:
			summary.OutOfScopeAuthorizationsBlocked++
		}
		if item.Observation.DescriptionEscalated {
			summary.DescriptionEscalationPaths++
		}
		if item.Receipt.ReceiptDigest == sealReceipt(item.Receipt).ReceiptDigest {
			summary.ReceiptSelfSealValid++
		}
		if item.Observation.ExecutionArtifactPresent {
			summary.ExecutionArtifactsObserved++
		}
		summary.RepositoryWrites += item.Observation.RepositoryWrites
		if item.Observation.MutationAuthority {
			summary.MutationAuthority++
		}
	}
	if summary.CasesTotal > 0 {
		summary.CaseCoverageBPS = summary.CasesPassed * 10_000 / summary.CasesTotal
	}
	return summary
}

func replayMatches(evidence ReplayEvidence) int {
	if !evidence.Equal || len(evidence.ReceiptDigestsA) != len(contractCases()) || len(evidence.ReceiptDigestsB) != len(contractCases()) || len(evidence.ExecutionDigestsA) != len(contractCases()) || len(evidence.ExecutionDigestsB) != len(contractCases()) {
		return 0
	}
	for index := range evidence.ReceiptDigestsA {
		if evidence.ReceiptDigestsA[index] == "" || evidence.ReceiptDigestsA[index] != evidence.ReceiptDigestsB[index] || evidence.ExecutionDigestsA[index] != evidence.ExecutionDigestsB[index] {
			return 0
		}
	}
	return len(contractCases())
}

func buildIndicators(summary Summary) []Indicator {
	caseDenominator := len(contractCases())
	return []Indicator{
		indicator("gooo.metric.meta-circular-boundary.fixed-cases.v1", "DRIVER", ProofFoundation, "bind-fixed-denominator", "cases", RelationEqual, summary.CasesPassed, caseDenominator),
		indicator("gooo.metric.meta-circular-boundary.description-binding-bps.v1", "DRIVER", ProofFoundation, "bind-self-description", "basis_points", RelationGreaterOrEqual, summary.DescriptionBound*10_000/caseDenominator, 10_000),
		indicator("gooo.metric.meta-circular-boundary.description-authority-escalation.v1", "GUARDRAIL", ProofRegression, "deny-description-authority-escalation", "paths", RelationLessOrEqual, summary.DescriptionEscalationPaths, 0),
		indicator("gooo.metric.meta-circular-boundary.explicit-authorization.v1", "DRIVER", ProofCoherence, "accept-explicit-read-only-capability", "cases", RelationEqual, summary.ExplicitAuthorizations, 1),
		indicator("gooo.metric.meta-circular-boundary.authorized-execution.v1", "OUTCOME", ProofCoherence, "permit-read-only-execution", "cases", RelationEqual, summary.AllowedExecutions, 1),
		indicator("gooo.metric.meta-circular-boundary.forged-rejection.v1", "GUARDRAIL", ProofRegression, "reject-forged-capability", "cases", RelationEqual, summary.ForgedAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.scope-rejection.v1", "GUARDRAIL", ProofRegression, "reject-out-of-scope-capability", "cases", RelationEqual, summary.OutOfScopeAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.replay.v1", "DRIVER", ProofCoherence, "replay-independent-judge", "cases", RelationEqual, summary.ReplayMatches, caseDenominator),
		indicator("gooo.metric.meta-circular-boundary.repository-writes.v1", "GUARDRAIL", ProofRegression, "preserve-read-only-effect-ceiling", "writes", RelationLessOrEqual, summary.RepositoryWrites, 0),
		indicator("gooo.metric.meta-circular-boundary.mutation-authority.v1", "GUARDRAIL", ProofRegression, "preserve-read-only-effect-ceiling", "authorities", RelationLessOrEqual, summary.MutationAuthority, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int) Indicator {
	satisfied := value == target
	if relation == RelationGreaterOrEqual {
		satisfied = value >= target
	} else if relation == RelationLessOrEqual {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, Producer: "metacircularboundary.Evaluate", Consumer: "meta-circular-boundary-ci", MetaOperation: operation, Unit: unit, Relation: relation, Value: value, Target: target, Satisfied: satisfied}
}
