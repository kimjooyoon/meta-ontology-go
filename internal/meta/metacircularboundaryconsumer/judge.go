package metacircularboundaryconsumer

import (
	"fmt"
	"reflect"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

// Judge is the consumer-side gate. It independently derives cases from the
// source computations and recomputes observations, receipts, summaries, and
// indicators without importing producer.
func Judge(report contract.Report, input contract.Input) error {
	if report.Schema != reportSchema || report.Scope != scope || !validSHA(input.HeadSHA) || report.HeadSHA != input.HeadSHA || input.Path != reportSourcePath() || (report.Source.Path != "" && report.Source.Path != input.Path) || report.ReportDigest != sealReport(report).ReportDigest {
		return fmt.Errorf("%s", reasonMismatch)
	}
	observed, err := observeSource(input.Path, input.Source)
	if err != nil {
		if !reflect.DeepEqual(report.Source, contract.SourceObservation{}) || report.Decision != decisionOpen || report.Resolution != resolutionLower || report.Reason != reasonSource || report.Coordinate != (contract.Coordinate{Stage: "READ_SOURCE", Step: "bind-self-description", Reason: reasonSource}) || len(report.Cases) != 0 || len(report.Receipts) != 0 || len(report.Indicators) != 0 {
			return fmt.Errorf("%s", reasonMismatch)
		}
		return nil
	}
	effect, effectErr := parseEffectEvidence(input.EffectEvidence)
	observed.Effect = effect
	observed.RepositoryWrites = effect.RepositoryWrites
	observed.MutationAuthority = effect.MutationAuthority == authorityGranted
	observed.ReadOnly = effect.Known && effect.OutputOutsideRepository && effect.RepositoryWrites == 0 && effect.MutationAuthority == authorityDenied
	grants, grantArtifactDigest, grantErr := parseGrantEvidence(input.GrantEvidence, observed.SemanticDigest)
	observed.GrantArtifactDigest = grantArtifactDigest
	replay, replayErr := parseReplayEvidence(input.ReplayEvidence)
	if !reflect.DeepEqual(observed, report.Source) || !validDigest(observed.SourceDigest) || !validDigest(observed.SemanticDigest) {
		return fmt.Errorf("%s", reasonSource)
	}
	if !reflect.DeepEqual(report.Denominator, expectedDenominator()) || len(report.Cases) != len(expectedCases()) || len(report.Receipts) != len(expectedCases()) || len(report.Indicators) != indicatorTotal || !reflect.DeepEqual(report.MetaOperations, expectedMetaOperations()) || !reflect.DeepEqual(report.NotClaimed, notClaimed()) || report.MetaValue != metaValue {
		return fmt.Errorf("%s", reasonMismatch)
	}

	attempts, factsErr := parseAttempts(observed)
	expected := expectedCases()
	artifacts := make(map[string]contract.ExecutionArtifact, len(input.ExecutionArtifacts))
	for _, artifact := range input.ExecutionArtifacts {
		artifacts[artifact.CaseID] = artifact
	}
	for index, definition := range expected {
		attempt, ok := attempts[definition.ID]
		if !ok {
			attempt = contract.Attempt{Predicate: definition.Predicate, Unknown: true}
		}
		grant := grants[definition.ID]
		artifact, artifactPresent := artifacts[definition.ID]
		if !reflect.DeepEqual(report.Cases[index].Definition, definition) || !reflect.DeepEqual(report.Cases[index].Attempt, attempt) || !reflect.DeepEqual(report.Cases[index].Grant, grant) {
			return fmt.Errorf("%s", reasonMismatch)
		}
		want := classify(observed, definition, attempt, grant, grantErr, artifact, artifactPresent)
		wantPassed := matches(definition, want)
		if !reflect.DeepEqual(report.Cases[index].Observation, want) || report.Cases[index].Passed != wantPassed {
			return fmt.Errorf("%s", reasonMismatch)
		}
		wantReceipt := receiptFor(observed, definition, attempt, grant, want, artifact, artifactPresent)
		if !reflect.DeepEqual(report.Cases[index].Receipt, wantReceipt) || !reflect.DeepEqual(report.Cases[index].Receipt, report.Receipts[index]) {
			return fmt.Errorf("%s", reasonMismatch)
		}
	}

	summary := summarize(report.Cases)
	summary.ReplayMatches = replayMatches(replay)
	if replayErr == nil && report.ReplayEvidenceDigest != digestBytes(input.ReplayEvidence) {
		return fmt.Errorf("%s", reasonMismatch)
	}
	if !reflect.DeepEqual(report.ExecutionArtifacts, input.ExecutionArtifacts) || !reflect.DeepEqual(report.Summary, summary) || !reflect.DeepEqual(report.Indicators, buildIndicators(summary)) || report.RepositoryWrites != observed.RepositoryWrites || report.MutationAuthority != observed.MutationAuthority {
		return fmt.Errorf("%s", reasonMismatch)
	}
	wantDecision, wantResolution, wantReason, wantCoordinate := reportDecision(observed, factsErr, grantErr, effectErr, replayErr, report.Cases, summary)
	if report.Decision != wantDecision || report.Resolution != wantResolution || report.Reason != wantReason || report.Coordinate != wantCoordinate {
		return fmt.Errorf("%s", reasonMismatch)
	}
	return nil
}

func JudgeWithReceipt(report contract.Report, input contract.Input) (contract.JudgeReceipt, error) {
	if err := Judge(report, input); err != nil {
		return contract.JudgeReceipt{}, err
	}
	receipt := contract.JudgeReceipt{Schema: judgeReceiptSchema, Producer: "metacircularboundaryconsumer.Judge", Consumer: "meta-circular-boundary-ci", InputReportDigest: report.ReportDigest, ComparedCases: len(report.Cases), Mismatches: 0, Decision: report.Decision, Reason: report.Reason}
	for _, item := range report.Receipts {
		receipt.ReceiptDigests = append(receipt.ReceiptDigests, item.ReceiptDigest)
	}
	receipt.Digest = digestValue(struct {
		Schema            string
		Producer          string
		Consumer          string
		InputReportDigest string
		ComparedCases     int
		Mismatches        int
		Decision          string
		Reason            string
		ReceiptDigests    []string
	}{receipt.Schema, receipt.Producer, receipt.Consumer, receipt.InputReportDigest, receipt.ComparedCases, receipt.Mismatches, receipt.Decision, receipt.Reason, receipt.ReceiptDigests})
	return receipt, nil
}

func reportSourcePath() string { return "examples/meta-circular-boundary/main.gooo" }

func classify(source contract.SourceObservation, definition contract.CaseDefinition, attempt contract.Attempt, grant contract.ExternalGrant, grantErr error, artifact contract.ExecutionArtifact, artifactPresent bool) contract.CaseObservation {
	observation := contract.CaseObservation{Description: descriptionBound, Authorization: authorizationUnknown, Execution: executionUnknown, Decision: decisionOpen, Predicate: attempt.Predicate, GrantDigest: grant.GrantDigest, ExecutionArtifactPresent: artifactPresent, RepositoryWrites: source.RepositoryWrites, MutationAuthority: source.MutationAuthority}
	if attempt.Unknown {
		observation.Description = "UNKNOWN"
		observation.Reason = reasonCaseData
		return observation
	}
	if attempt.Contradictory {
		observation.Reason = reasonContradictory
		observation.Decision = decisionRefuted
		observation.Authorization = authorizationDeny
		return observation
	}
	if !source.DescriptionBound || attempt.DescriptionDigest != source.SemanticDigest {
		observation.Description = "UNKNOWN"
		observation.Reason = reasonSource
		return observation
	}
	if !source.Graph.Valid {
		observation.Reason = reasonGraphUnknown
		return observation
	}
	if grantErr != nil || grant.CaseID == "" {
		observation.Reason = reasonGrantUnknown
		return observation
	}
	if attempt.DescriptionAuthorityClaim {
		observation.DescriptionEscalated = true
	}
	if definition.ID == "description-only" {
		if attempt.RequestKind != requestNone || grant.Decision != grantDeny {
			observation.Reason = reasonCasePredicate
			return observation
		}
		observation.Authorization, observation.Execution, observation.Decision = authorizationDeny, executionBlocked, decisionFailClosed
		if observation.DescriptionEscalated {
			observation.Reason = reasonDescriptionForgery
		} else {
			observation.Reason = reasonDescription
		}
		return observation
	}
	if attempt.RequestKind != requestReadOnly || grant.Decision != grantDecision {
		observation.Reason = reasonCasePredicate
		return observation
	}
	if grant.SubjectDigest != source.SemanticDigest {
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = authorizationDeny, executionBlocked, decisionFailClosed, reasonGrantDenied
		return observation
	}
	if definition.ID == "forged-capability" {
		observation.Predicate = predicateForgedGrant
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = authorizationDeny, executionBlocked, decisionFailClosed, reasonForged
		return observation
	}
	if definition.ID == "write-capability-out-of-scope" || grant.Scope == scopeWrite {
		observation.Predicate = predicateOutOfScopeGrant
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = authorizationDeny, executionBlocked, decisionFailClosed, reasonOutOfScope
		return observation
	}
	if grant.Issuer != "external-authority" || grant.Operation != metaOperationID || grant.Scope != scopeReadOnly || grant.Handle == "" {
		observation.Authorization, observation.Execution, observation.Decision, observation.Reason = authorizationDeny, executionBlocked, decisionFailClosed, reasonForged
		return observation
	}
	observation.Authorization, observation.Reason = authorizationGrant, reasonExplicit
	if !attempt.RequestExecution {
		observation.Execution, observation.Decision = executionBlocked, decisionFailClosed
		return observation
	}
	if !artifactPresent {
		observation.Reason = reasonExecutionUnknown
		return observation
	}
	observation.ExecutionArtifactValid = validExecutionArtifact(source, grant, definition.ID, artifact)
	if !observation.ExecutionArtifactValid {
		observation.Reason = reasonExecutionInvalid
		observation.Decision = decisionRefuted
		return observation
	}
	observation.Execution, observation.OutputDigest, observation.Decision = executionAllowed, artifact.OutputDigest, decisionPass
	return observation
}

func reportDecision(source contract.SourceObservation, factsErr, grantErr, effectErr, replayErr error, cases []contract.CaseResult, summary contract.Summary) (string, string, string, contract.Coordinate) {
	if effectErr != nil || !source.Effect.Known || source.Effect.MutationAuthority == authorityUnknown || !source.Effect.OutputOutsideRepository || source.Effect.PermissionEvidence == "" {
		return decisionOpen, resolutionLower, reasonEffectUnknown, contract.Coordinate{Stage: "OBSERVE_EFFECT", Step: "resolve-workspace-permission-and-output", Reason: reasonEffectUnknown}
	}
	if hasDecision(cases, decisionRefuted) {
		return decisionRefuted, resolutionExact, reasonContradictory, contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "refute-capability-or-execution-evidence", Reason: reasonContradictory}
	}
	if factsErr != nil || grantErr != nil || effectErr != nil || replayErr != nil || hasDecision(cases, decisionOpen) {
		reason, stage, step := reasonCaseData, "PARSE_COMPUTES", "read-case-facts"
		if grantErr != nil {
			reason, stage, step = reasonGrantUnknown, "READ_EXTERNAL_GRANT", "parse-grant-evidence"
		} else if effectErr != nil {
			reason, stage, step = reasonEffectUnknown, "OBSERVE_EFFECT", "compare-workspace-state"
		} else if replayErr != nil {
			reason, stage, step = reasonReplayUnknown, "OBSERVE_REPLAY", "compare-receipt-digests"
		} else if hasReason(cases, reasonGraphUnknown) {
			reason, stage, step = reasonGraphUnknown, "LOWER_GRAPH", "reconstruct-describe-grant-execute"
		} else if hasReason(cases, reasonExecutionUnknown) {
			reason, stage, step = reasonExecutionUnknown, "EXECUTE_META_OPERATION", "observe-output-artifact"
		}
		return decisionOpen, resolutionLower, reason, contract.Coordinate{Stage: stage, Step: step, Reason: reason}
	}
	if summary.CasesPassed != len(expectedCases()) || summary.DescriptionBound != len(expectedCases()) || summary.ExplicitAuthorizations != 1 || summary.AllowedExecutions != 1 || summary.DescriptionEscalationPaths != 0 || summary.RepositoryWrites != 0 || summary.MutationAuthority != 0 || summary.ReplayMatches != len(expectedCases()) || !source.ReadOnly || !source.Effect.Known || source.Effect.MutationAuthority != authorityDenied {
		return decisionFailClosed, resolutionLower, reasonContractUnsatisfied, contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "evaluate-observed-boundary-evidence", Reason: reasonContractUnsatisfied}
	}
	return decisionPass, resolutionExact, reasonContract, contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "classify-gooo-computations", Reason: reasonContract}
}

func hasDecision(cases []contract.CaseResult, decision string) bool {
	for _, item := range cases {
		if item.Observation.Decision == decision {
			return true
		}
	}
	return false
}

func hasReason(cases []contract.CaseResult, reason string) bool {
	for _, item := range cases {
		if item.Observation.Reason == reason {
			return true
		}
	}
	return false
}

func matches(definition contract.CaseDefinition, observation contract.CaseObservation) bool {
	if definition.ID == "" || definition.Predicate == "" || observation.Predicate != definition.Predicate || observation.Description == "UNKNOWN" || observation.Decision == decisionOpen || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return false
	}
	if definition.ExpectedDecision != "" && (observation.Decision != definition.ExpectedDecision || observation.Authorization != definition.ExpectedAuthorization || observation.Execution != definition.ExpectedExecution || observation.Reason != definition.ExpectedReason) {
		return false
	}
	switch definition.ID {
	case "description-only":
		return observation.Authorization == authorizationDeny && observation.Execution == executionBlocked && !observation.ExecutionArtifactPresent
	case "explicit-read-only-capability":
		return observation.Predicate == predicateExplicitGrant && observation.Authorization == authorizationGrant && observation.Execution == executionAllowed && observation.ExecutionArtifactValid && observation.OutputDigest != ""
	case "forged-capability":
		return observation.Predicate == predicateForgedGrant && observation.Authorization == authorizationDeny && observation.Execution == executionBlocked && !observation.ExecutionArtifactPresent && observation.Reason == reasonForged
	case "write-capability-out-of-scope":
		return observation.Predicate == predicateOutOfScopeGrant && observation.Authorization == authorizationDeny && observation.Execution == executionBlocked && !observation.ExecutionArtifactPresent && observation.Reason == reasonOutOfScope
	}
	return false
}

func receiptFor(source contract.SourceObservation, definition contract.CaseDefinition, attempt contract.Attempt, grant contract.ExternalGrant, observation contract.CaseObservation, artifact contract.ExecutionArtifact, artifactPresent bool) contract.Receipt {
	authorizationEvidenceDigest := ""
	if grant.Decision == grantDecision {
		authorizationEvidenceDigest = digestValue(grant)
	}
	receipt := contract.Receipt{
		Schema: receiptSchema, CaseID: definition.ID,
		Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge",
		MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice,
		Coordinate:   contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: definition.ID, Reason: observation.Reason},
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

func claimTransitions(definition contract.CaseDefinition, attempt contract.Attempt, grant contract.ExternalGrant, observation contract.CaseObservation, artifact contract.ExecutionArtifact, artifactPresent bool) []contract.ClaimTransition {
	descriptionAfter, descriptionEvent := "DESCRIBED", "DESCRIPTION_OBSERVED"
	if observation.Description != descriptionBound {
		descriptionAfter, descriptionEvent = "UNKNOWN", "DESCRIPTION_UNRESOLVED"
	} else if observation.DescriptionEscalated {
		descriptionAfter, descriptionEvent = "REFUTED", "DESCRIPTION_AUTHORITY_CLAIM_REFUTED"
	}
	descriptionClaim := definition.ID + ".description"
	authorizationClaim := definition.ID + ".authorization"
	executionClaim := definition.ID + ".execution"
	transitions := []contract.ClaimTransition{{ClaimID: descriptionClaim, PropositionDigest: propositionDigest(descriptionClaim, descriptionAfter), Event: descriptionEvent, Before: "UNRECORDED", After: descriptionAfter, Coordinate: contract.Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: observation.Reason}, EvidenceDigest: attempt.DescriptionDigest}}
	after, event := "DENIED", "AUTHORIZATION_REJECTED"
	if observation.Decision == decisionOpen {
		after, event = "UNKNOWN", "AUTHORIZATION_UNRESOLVED"
	} else if observation.Decision == decisionRefuted {
		after, event = "REFUTED", "CAPABILITY_EVIDENCE_REFUTED"
	} else if observation.Authorization == authorizationGrant {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	transitions = append(transitions, contract.ClaimTransition{ClaimID: authorizationClaim, PropositionDigest: propositionDigest(authorizationClaim, after), Event: event, Before: "UNRECORDED", After: after, Coordinate: contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}, EvidenceDigest: grant.GrantDigest, DependsOnClaimID: descriptionClaim, DependsOnAfter: descriptionAfter})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Decision == decisionOpen {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Decision == decisionRefuted {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Execution == executionAllowed {
		executionAfter, executionEvent = "EXECUTED_READ_ONLY", "EXECUTION_ALLOWED"
	}
	executionEvidence := ""
	if artifactPresent {
		executionEvidence = artifact.OutputDigest
	}
	return append(transitions, contract.ClaimTransition{ClaimID: executionClaim, PropositionDigest: propositionDigest(executionClaim, executionAfter), Event: executionEvent, Before: "NOT_EXECUTED", After: executionAfter, Coordinate: contract.Coordinate{Stage: "SEAL_RECEIPT", Step: "seal-read-only-result", Reason: observation.Reason}, EvidenceDigest: executionEvidence, DependsOnClaimID: authorizationClaim, DependsOnAfter: after})
}

func propositionDigest(claimID, state string) string {
	return digestBytes([]byte("proposition|" + claimID + "|" + state))
}

func summarize(cases []contract.CaseResult) contract.Summary {
	summary := contract.Summary{CasesTotal: len(cases)}
	for _, item := range cases {
		if item.Passed {
			summary.CasesPassed++
		}
		if item.Observation.Description == descriptionBound {
			summary.DescriptionBound++
		}
		if item.Observation.Authorization == authorizationGrant {
			summary.ExplicitAuthorizations++
		}
		if item.Observation.Execution == executionAllowed {
			summary.AllowedExecutions++
		}
		switch item.Observation.Reason {
		case reasonDescription:
			summary.DescriptionOnlyBlocked++
		case reasonForged:
			summary.ForgedAuthorizationsBlocked++
		case reasonOutOfScope:
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

func replayMatches(evidence contract.ReplayEvidence) int {
	if !evidence.Equal || len(evidence.ReceiptDigestsA) != len(expectedCases()) || len(evidence.ReceiptDigestsB) != len(expectedCases()) || len(evidence.ExecutionDigestsA) != len(expectedCases()) || len(evidence.ExecutionDigestsB) != len(expectedCases()) {
		return 0
	}
	for index := range evidence.ReceiptDigestsA {
		if evidence.ReceiptDigestsA[index] == "" || evidence.ReceiptDigestsA[index] != evidence.ReceiptDigestsB[index] || evidence.ExecutionDigestsA[index] != evidence.ExecutionDigestsB[index] {
			return 0
		}
	}
	return len(expectedCases())
}

func buildIndicators(summary contract.Summary) []contract.Indicator {
	return []contract.Indicator{
		indicator("gooo.metric.meta-circular-boundary.fixed-cases.v1", "DRIVER", proofFoundation, "bind-fixed-denominator", "cases", relationEqual, summary.CasesPassed, len(expectedCases())),
		indicator("gooo.metric.meta-circular-boundary.description-binding-bps.v1", "DRIVER", proofFoundation, "bind-self-description", "basis_points", relationGreaterOrEqual, summary.DescriptionBound*10_000/len(expectedCases()), 10_000),
		indicator("gooo.metric.meta-circular-boundary.description-authority-escalation.v1", "GUARDRAIL", proofRegression, "deny-description-authority-escalation", "paths", relationLessOrEqual, summary.DescriptionEscalationPaths, 0),
		indicator("gooo.metric.meta-circular-boundary.explicit-authorization.v1", "DRIVER", proofCoherence, "accept-explicit-read-only-capability", "cases", relationEqual, summary.ExplicitAuthorizations, 1),
		indicator("gooo.metric.meta-circular-boundary.authorized-execution.v1", "OUTCOME", proofCoherence, "permit-read-only-execution", "cases", relationEqual, summary.AllowedExecutions, 1),
		indicator("gooo.metric.meta-circular-boundary.forged-rejection.v1", "GUARDRAIL", proofRegression, "reject-forged-capability", "cases", relationEqual, summary.ForgedAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.scope-rejection.v1", "GUARDRAIL", proofRegression, "reject-out-of-scope-capability", "cases", relationEqual, summary.OutOfScopeAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.replay.v1", "DRIVER", proofCoherence, "replay-independent-judge", "cases", relationEqual, summary.ReplayMatches, len(expectedCases())),
		indicator("gooo.metric.meta-circular-boundary.repository-writes.v1", "GUARDRAIL", proofRegression, "preserve-read-only-effect-ceiling", "writes", relationLessOrEqual, summary.RepositoryWrites, 0),
		indicator("gooo.metric.meta-circular-boundary.mutation-authority.v1", "GUARDRAIL", proofRegression, "preserve-read-only-effect-ceiling", "authorities", relationLessOrEqual, summary.MutationAuthority, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int) contract.Indicator {
	satisfied := value == target
	if relation == relationGreaterOrEqual {
		satisfied = value >= target
	} else if relation == relationLessOrEqual {
		satisfied = value <= target
	}
	return contract.Indicator{MetricID: id, Class: class, ProofChoice: proof, Producer: "metacircularboundary.Evaluate", Consumer: "meta-circular-boundary-ci", MetaOperation: operation, Unit: unit, Relation: relation, Value: value, Target: target, Satisfied: satisfied}
}
