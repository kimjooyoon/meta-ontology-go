package metacircularboundary

func Evaluate(input Input) Report {
	report := Report{
		Schema: ReportSchema, Scope: Scope, HeadSHA: input.HeadSHA, Decision: DecisionFailClosed,
		Resolution: ResolutionLower, Reason: ReasonSourceBindingUnknown,
		Coordinate:     Coordinate{Stage: "READ_SOURCE", Step: "bind-self-description", Reason: ReasonSourceBindingUnknown},
		MetaOperations: MetaOperations(), RepositoryWrites: 0, MutationAuthority: false,
		MetaValue: "DESCRIPTION_AUTHORIZATION_EXECUTION_ARE_DISTINCT",
		NotClaimed: []string{
			"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
			"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
		},
	}
	if input.Path != ExpectedSourcePath || !validSHA(input.HeadSHA) {
		return sealReport(report)
	}
	source, err := observeSource(input.Path, input.Source)
	if err != nil {
		return sealReport(report)
	}
	report.Source = source
	report.Coordinate = Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "classify-fixed-cases", Reason: ReasonContractSatisfied}
	denominator := DenominatorContract()
	for _, definition := range denominator.Cases {
		attempt, ok := CaseInput(definition.ID, source.SourceDigest)
		if !ok {
			continue
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
	if report.Summary.CasesPassed != CaseTotal || report.Summary.DescriptionEscalationPaths != 0 || report.Summary.RepositoryWrites != 0 || report.Summary.MutationAuthority != 0 {
		report.Reason = ReasonReplayMismatch
	} else {
		report.Decision, report.Resolution, report.Reason = DecisionPass, ResolutionExact, ReasonContractSatisfied
	}
	report.IndependentJudge = JudgeEvidence{Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ComparedCases: len(report.Cases), Decision: report.Decision, Reason: report.Reason}
	return sealReport(report)
}

type Input struct {
	Path    string
	HeadSHA string
	Source  []byte
}

func classify(source SourceObservation, attempt Attempt) CaseObservation {
	descriptionBound := source.DescriptionBound && attempt.DescriptionDigest == source.SourceDigest
	observation := CaseObservation{Description: DescriptionBound, Authorization: AuthorizationDenied, Execution: ExecutionBlocked, RepositoryWrites: 0, MutationAuthority: false}
	if !descriptionBound {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonSourceBindingUnknown
		return observation
	}
	if attempt.Capability == nil {
		observation.Reason = ReasonDescriptionOnly
		return observation
	}
	if attempt.Capability.Scope == ScopeWrite {
		observation.Reason = ReasonOutOfScopeCapability
		return observation
	}
	if attempt.Capability.Issuer != "external-authority" || attempt.Capability.SubjectDigest != source.SourceDigest || attempt.Capability.Operation != MetaOperationID || attempt.Capability.Scope != ScopeReadOnly || attempt.Capability.Handle != capabilityHandle(source.SourceDigest) {
		observation.Reason = ReasonForgedCapability
		return observation
	}
	observation.Authorization = AuthorizationGranted
	observation.Reason = ReasonExplicitCapability
	if attempt.RequestExecution {
		observation.Execution = ExecutionAllowed
	}
	return observation
}

func buildReceipt(source SourceObservation, definition CaseDefinition, attempt Attempt, observation CaseObservation) Receipt {
	capabilityDigest := ""
	if attempt.Capability != nil {
		capabilityDigest = digestValue(attempt.Capability)
	}
	receipt := Receipt{
		Schema: "gooo/meta-circular-boundary-receipt/v1", CaseID: definition.ID,
		Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge",
		MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice,
		Coordinate:   Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: definition.ID, Reason: observation.Reason},
		SourceDigest: source.SourceDigest, DescriptionDigest: attempt.DescriptionDigest, CapabilityDigest: capabilityDigest,
		Decision: definition.ExpectedDecision, Authorization: observation.Authorization, Execution: observation.Execution,
		RepositoryWrites: observation.RepositoryWrites, MutationAuthority: observation.MutationAuthority,
		ClaimTransitions: claimTransitions(definition, attempt, observation),
	}
	return sealReceipt(receipt)
}

func claimTransitions(definition CaseDefinition, attempt Attempt, observation CaseObservation) []ClaimTransition {
	transitions := []ClaimTransition{{ClaimID: definition.ID + ".description", Event: "DESCRIPTION_OBSERVED", Before: "UNRECORDED", After: "DESCRIBED", Coordinate: Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: "SELF_DESCRIPTION_OBSERVED"}, EvidenceDigest: attempt.DescriptionDigest}}
	after := "DENIED"
	event := "AUTHORIZATION_REJECTED"
	if observation.Authorization == AuthorizationGranted {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	transitions = append(transitions, ClaimTransition{ClaimID: definition.ID + ".authorization", Event: event, Before: "UNRECORDED", After: after, Coordinate: Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Execution == ExecutionAllowed {
		executionAfter, executionEvent = "EXECUTED_READ_ONLY", "EXECUTION_ALLOWED"
	}
	transitions = append(transitions, ClaimTransition{ClaimID: definition.ID + ".execution", Event: executionEvent, Before: "NOT_EXECUTED", After: executionAfter, Coordinate: Coordinate{Stage: "SEAL_RECEIPT", Step: "seal-read-only-result", Reason: observation.Reason}})
	return transitions
}

func matches(definition CaseDefinition, observation CaseObservation) bool {
	return observation.Description == DescriptionBound && observation.Authorization == definition.ExpectedAuthorization && observation.Execution == definition.ExpectedExecution && observation.Reason == definition.ExpectedReason
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
		switch item.Definition.ID {
		case "description-only":
			if item.Observation.Execution == ExecutionBlocked {
				summary.DescriptionOnlyBlocked++
			}
		case "forged-capability":
			if item.Observation.Authorization == AuthorizationDenied {
				summary.ForgedAuthorizationsBlocked++
			}
		case "write-capability-out-of-scope":
			if item.Observation.Authorization == AuthorizationDenied {
				summary.OutOfScopeAuthorizationsBlocked++
			}
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
