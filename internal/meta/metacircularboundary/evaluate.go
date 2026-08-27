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
	report.Source = source
	attempts, factsErr := parseAttempts(source)
	definitions := DenominatorContract().Cases
	for _, definition := range definitions {
		attempt, ok := attempts[definition.ID]
		if !ok {
			attempt = Attempt{Unknown: true}
		}
		observation := classify(source, attempt)
		receipt := buildReceipt(source, definition, attempt, observation)
		item := CaseResult{Definition: definition, Attempt: attempt, Observation: observation, Receipt: receipt}
		item.Passed = matches(definition, observation) && receipt.ReceiptDigest == sealReceipt(receipt).ReceiptDigest
		report.Cases = append(report.Cases, item)
		report.Receipts = append(report.Receipts, receipt)
	}
	report.Summary = summarize(report.Cases)
	report.Indicators = buildIndicators(report.Summary)
	report.Decision, report.Resolution, report.Reason, report.Coordinate = reportDecision(factsErr, report.Cases, report.Summary)
	report.IndependentJudge = JudgeEvidence{Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ComparedCases: len(report.Cases), Decision: report.Decision, Reason: report.Reason}
	return sealReport(report)
}

func baseReport(input Input) Report {
	return Report{
		Schema: ReportSchema, Scope: Scope, HeadSHA: input.HeadSHA, Decision: DecisionOpen,
		Resolution: ResolutionLower, Reason: ReasonSourceBindingUnknown,
		Coordinate:     Coordinate{Stage: "READ_SOURCE", Step: "bind-self-description", Reason: ReasonSourceBindingUnknown},
		MetaOperations: MetaOperations(), RepositoryWrites: 0, MutationAuthority: false,
		MetaValue: "DESCRIPTION_AUTHORIZATION_EXECUTION_ARE_DISTINCT",
		NotClaimed: []string{
			"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
			"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
		},
	}
}

func classify(source SourceObservation, attempt Attempt) CaseObservation {
	observation := CaseObservation{Description: DescriptionBound, Authorization: AuthorizationDenied, Execution: ExecutionBlocked, Decision: DecisionOpen, RepositoryWrites: 0, MutationAuthority: false}
	if attempt.Unknown {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonCaseDataUnknown
		return observation
	}
	if attempt.Contradictory {
		observation.Reason = ReasonContradictory
		observation.Decision = DecisionRefuted
		return observation
	}
	if !(source.DescriptionBound && attempt.DescriptionDigest == source.SourceDigest) {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonSourceBindingUnknown
		return observation
	}
	if attempt.Capability == nil {
		observation.Reason = ReasonDescriptionOnly
		observation.Decision = DecisionFailClosed
		return observation
	}
	if attempt.Capability.Scope == ScopeWrite {
		observation.Reason = ReasonOutOfScopeCapability
		observation.Decision = DecisionFailClosed
		return observation
	}
	grant := attempt.Capability
	if grant.Issuer != "external-authority" || grant.SubjectDigest != source.SourceDigest || grant.Operation != MetaOperationID || grant.Scope != ScopeReadOnly || grant.Handle != capabilityHandle(source.SourceDigest) {
		observation.Reason = ReasonForgedCapability
		observation.Decision = DecisionFailClosed
		return observation
	}
	observation.Authorization, observation.Reason, observation.Decision = AuthorizationGranted, ReasonExplicitCapability, DecisionPass
	if attempt.RequestExecution {
		observation.Execution = ExecutionAllowed
	}
	return observation
}

func reportDecision(factsErr error, cases []CaseResult, summary Summary) (string, string, string, Coordinate) {
	if factsErr != nil || hasDecision(cases, DecisionOpen) {
		return DecisionOpen, ResolutionLower, ReasonCaseDataUnknown,
			Coordinate{Stage: "PARSE_COMPUTES", Step: "read-case-facts", Reason: ReasonCaseDataUnknown}
	}
	if hasDecision(cases, DecisionRefuted) {
		return DecisionRefuted, ResolutionExact, ReasonContradictory,
			Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "refute-capability-evidence", Reason: ReasonContradictory}
	}
	if summary.CasesPassed != CaseTotal || summary.DescriptionBound != CaseTotal || summary.ExplicitAuthorizations != 1 || summary.AllowedExecutions != 1 || summary.DescriptionEscalationPaths != 0 || summary.RepositoryWrites != 0 || summary.MutationAuthority != 0 {
		return DecisionFailClosed, ResolutionLower, ReasonContractUnsatisfied,
			Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "evaluate-fixed-boundary-indicators", Reason: ReasonContractUnsatisfied}
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

func buildReceipt(source SourceObservation, definition CaseDefinition, attempt Attempt, observation CaseObservation) Receipt {
	capabilityDigest := ""
	if attempt.Capability != nil {
		capabilityDigest = digestValue(attempt.Capability)
	}
	receipt := Receipt{
		Schema: "gooo/meta-circular-boundary-receipt/v1", CaseID: definition.ID,
		Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge",
		MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice,
		Coordinate:   Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: definition.ID, Reason: observation.Reason},
		SourceDigest: source.SourceDigest, DescriptionDigest: attempt.DescriptionDigest, CapabilityDigest: capabilityDigest,
		Decision: observation.Decision, Authorization: observation.Authorization, Execution: observation.Execution,
		RepositoryWrites: observation.RepositoryWrites, MutationAuthority: observation.MutationAuthority,
		ClaimTransitions: claimTransitions(definition, attempt, observation),
	}
	return sealReceipt(receipt)
}

func claimTransitions(definition CaseDefinition, attempt Attempt, observation CaseObservation) []ClaimTransition {
	descriptionAfter, descriptionEvent := "DESCRIBED", "DESCRIPTION_OBSERVED"
	if observation.Description != DescriptionBound {
		descriptionAfter, descriptionEvent = "UNKNOWN", "DESCRIPTION_UNRESOLVED"
	}
	transitions := []ClaimTransition{{ClaimID: definition.ID + ".description", Event: descriptionEvent, Before: "UNRECORDED", After: descriptionAfter, Coordinate: Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: observation.Reason}, EvidenceDigest: attempt.DescriptionDigest}}
	after, event := "DENIED", "AUTHORIZATION_REJECTED"
	if observation.Decision == DecisionOpen {
		after, event = "UNKNOWN", "AUTHORIZATION_UNRESOLVED"
	} else if observation.Decision == DecisionRefuted {
		after, event = "REFUTED", "CAPABILITY_EVIDENCE_REFUTED"
	} else if observation.Authorization == AuthorizationGranted {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	transitions = append(transitions, ClaimTransition{ClaimID: definition.ID + ".authorization", Event: event, Before: "UNRECORDED", After: after, Coordinate: Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Decision == DecisionOpen {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Decision == DecisionRefuted {
		executionAfter, executionEvent = "REFUTED", "EXECUTION_REFUTED"
	} else if observation.Execution == ExecutionAllowed {
		executionAfter, executionEvent = "EXECUTED_READ_ONLY", "EXECUTION_ALLOWED"
	}
	return append(transitions, ClaimTransition{ClaimID: definition.ID + ".execution", Event: executionEvent, Before: "NOT_EXECUTED", After: executionAfter, Coordinate: Coordinate{Stage: "SEAL_RECEIPT", Step: "seal-read-only-result", Reason: observation.Reason}})
}

func matches(definition CaseDefinition, observation CaseObservation) bool {
	if definition.ID == "" || observation.Description == "UNKNOWN" || observation.Decision == DecisionOpen || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return false
	}
	if observation.Decision == DecisionRefuted {
		return false
	}
	if observation.Authorization == AuthorizationGranted {
		return observation.Execution == ExecutionAllowed || observation.Execution == ExecutionBlocked
	}
	return observation.Authorization == AuthorizationDenied && observation.Execution == ExecutionBlocked
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
			summary.ReplayMatches++
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

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		indicator("gooo.metric.meta-circular-boundary.fixed-cases.v1", "DRIVER", ProofFoundation, "bind-fixed-denominator", "cases", RelationEqual, summary.CasesPassed, CaseTotal),
		indicator("gooo.metric.meta-circular-boundary.description-binding-bps.v1", "DRIVER", ProofFoundation, "bind-self-description", "basis_points", RelationGreaterOrEqual, summary.DescriptionBound*10_000/CaseTotal, 10_000),
		indicator("gooo.metric.meta-circular-boundary.description-authority-escalation.v1", "GUARDRAIL", ProofRegression, "deny-description-authority-escalation", "paths", RelationLessOrEqual, summary.DescriptionEscalationPaths, 0),
		indicator("gooo.metric.meta-circular-boundary.explicit-authorization.v1", "DRIVER", ProofCoherence, "accept-explicit-read-only-capability", "cases", RelationEqual, summary.ExplicitAuthorizations, 1),
		indicator("gooo.metric.meta-circular-boundary.authorized-execution.v1", "OUTCOME", ProofCoherence, "permit-read-only-execution", "cases", RelationEqual, summary.AllowedExecutions, 1),
		indicator("gooo.metric.meta-circular-boundary.forged-rejection.v1", "GUARDRAIL", ProofRegression, "reject-forged-capability", "cases", RelationEqual, summary.ForgedAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.scope-rejection.v1", "GUARDRAIL", ProofRegression, "reject-out-of-scope-capability", "cases", RelationEqual, summary.OutOfScopeAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.replay.v1", "DRIVER", ProofCoherence, "replay-independent-judge", "cases", RelationEqual, summary.ReplayMatches, CaseTotal),
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
