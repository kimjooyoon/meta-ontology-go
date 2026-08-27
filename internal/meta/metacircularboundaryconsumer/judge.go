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
	if !reflect.DeepEqual(observed, report.Source) || !validDigest(observed.SourceDigest) || !validDigest(observed.SemanticDigest) {
		return fmt.Errorf("%s", reasonSource)
	}
	if len(report.Cases) != caseTotal || len(report.Receipts) != caseTotal || len(report.Indicators) != indicatorTotal || !reflect.DeepEqual(report.MetaOperations, expectedMetaOperations()) || !reflect.DeepEqual(report.NotClaimed, notClaimed()) || report.MetaValue != metaValue {
		return fmt.Errorf("%s", reasonMismatch)
	}

	attempts, factsErr := parseAttempts(observed)
	expected := expectedCases()
	for index, definition := range expected {
		attempt, ok := attempts[definition.ID]
		if !ok {
			attempt = contract.Attempt{Unknown: true}
		}
		if !reflect.DeepEqual(report.Cases[index].Definition, definition) || !reflect.DeepEqual(report.Cases[index].Attempt, attempt) {
			return fmt.Errorf("%s", reasonMismatch)
		}
		want := classify(observed, attempt)
		wantPassed := matches(definition, want)
		if !reflect.DeepEqual(report.Cases[index].Observation, want) || report.Cases[index].Passed != wantPassed {
			return fmt.Errorf("%s", reasonMismatch)
		}
		wantReceipt := receiptFor(observed, definition, attempt, want)
		if !reflect.DeepEqual(report.Cases[index].Receipt, wantReceipt) || !reflect.DeepEqual(report.Cases[index].Receipt, report.Receipts[index]) {
			return fmt.Errorf("%s", reasonMismatch)
		}
	}

	summary := summarize(report.Cases)
	if !reflect.DeepEqual(report.Summary, summary) || !reflect.DeepEqual(report.Indicators, buildIndicators(summary)) || report.RepositoryWrites != 0 || report.MutationAuthority {
		return fmt.Errorf("%s", reasonMismatch)
	}
	wantDecision, wantResolution, wantReason, wantCoordinate := reportDecision(factsErr, report.Cases, summary)
	if report.Decision != wantDecision || report.Resolution != wantResolution || report.Reason != wantReason || report.Coordinate != wantCoordinate {
		return fmt.Errorf("%s", reasonMismatch)
	}
	if report.IndependentJudge != (contract.JudgeEvidence{Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ComparedCases: caseTotal, Mismatches: 0, Decision: wantDecision, Reason: wantReason}) {
		return fmt.Errorf("%s", reasonMismatch)
	}
	return nil
}

func reportSourcePath() string { return "examples/meta-circular-boundary/main.gooo" }

func classify(source contract.SourceObservation, attempt contract.Attempt) contract.CaseObservation {
	observation := contract.CaseObservation{Description: descriptionBound, Authorization: authorizationDeny, Execution: executionBlocked, Decision: decisionOpen, RepositoryWrites: 0, MutationAuthority: false}
	if attempt.Unknown {
		observation.Description = "UNKNOWN"
		observation.Reason = reasonCaseData
		return observation
	}
	if attempt.Contradictory {
		observation.Reason = reasonContradictory
		observation.Decision = decisionRefuted
		return observation
	}
	if !(source.DescriptionBound && attempt.DescriptionDigest == source.SourceDigest) {
		observation.Description = "UNKNOWN"
		observation.Reason = reasonSource
		return observation
	}
	if attempt.Capability == nil {
		observation.Reason = reasonDescription
		observation.Decision = decisionFailClosed
		return observation
	}
	if attempt.Capability.Scope == scopeWrite {
		observation.Reason = reasonOutOfScope
		observation.Decision = decisionFailClosed
		return observation
	}
	grant := attempt.Capability
	if grant.Issuer != "external-authority" || grant.SubjectDigest != source.SourceDigest || grant.Operation != metaOperationID || grant.Scope != scopeReadOnly || grant.Handle != capabilityHandle(source.SourceDigest) {
		observation.Reason = reasonForged
		observation.Decision = decisionFailClosed
		return observation
	}
	observation.Authorization, observation.Reason, observation.Decision = authorizationGrant, reasonExplicit, decisionPass
	if attempt.RequestExecution {
		observation.Execution = executionAllowed
	}
	return observation
}

func reportDecision(factsErr error, cases []contract.CaseResult, summary contract.Summary) (string, string, string, contract.Coordinate) {
	if factsErr != nil || hasDecision(cases, decisionOpen) {
		return decisionOpen, resolutionLower, reasonCaseData, contract.Coordinate{Stage: "PARSE_COMPUTES", Step: "read-case-facts", Reason: reasonCaseData}
	}
	if hasDecision(cases, decisionRefuted) {
		return decisionRefuted, resolutionExact, reasonContradictory, contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "refute-capability-evidence", Reason: reasonContradictory}
	}
	if summary.CasesPassed != caseTotal || summary.DescriptionBound != caseTotal || summary.ExplicitAuthorizations != 1 || summary.AllowedExecutions != 1 || summary.DescriptionEscalationPaths != 0 || summary.RepositoryWrites != 0 || summary.MutationAuthority != 0 {
		return decisionFailClosed, resolutionLower, reasonContractUnsatisfied, contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "evaluate-fixed-boundary-indicators", Reason: reasonContractUnsatisfied}
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

func matches(definition contract.CaseDefinition, observation contract.CaseObservation) bool {
	if definition.ID == "" || observation.Description == "UNKNOWN" || observation.Decision == decisionOpen || observation.Decision == decisionRefuted || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return false
	}
	if observation.Authorization == authorizationGrant {
		return observation.Execution == executionAllowed || observation.Execution == executionBlocked
	}
	return observation.Authorization == authorizationDeny && observation.Execution == executionBlocked
}

func receiptFor(source contract.SourceObservation, definition contract.CaseDefinition, attempt contract.Attempt, observation contract.CaseObservation) contract.Receipt {
	capabilityDigest := ""
	if attempt.Capability != nil {
		capabilityDigest = digestValue(attempt.Capability)
	}
	receipt := contract.Receipt{
		Schema: receiptSchema, CaseID: definition.ID,
		Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge",
		MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice,
		Coordinate:   contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: definition.ID, Reason: observation.Reason},
		SourceDigest: source.SourceDigest, DescriptionDigest: attempt.DescriptionDigest, CapabilityDigest: capabilityDigest,
		Decision: observation.Decision, Authorization: observation.Authorization, Execution: observation.Execution,
		RepositoryWrites: observation.RepositoryWrites, MutationAuthority: observation.MutationAuthority,
		ClaimTransitions: claimTransitions(definition, attempt, observation),
	}
	return sealReceipt(receipt)
}

func claimTransitions(definition contract.CaseDefinition, attempt contract.Attempt, observation contract.CaseObservation) []contract.ClaimTransition {
	descriptionAfter, descriptionEvent := "DESCRIBED", "DESCRIPTION_OBSERVED"
	if observation.Description != descriptionBound {
		descriptionAfter, descriptionEvent = "UNKNOWN", "DESCRIPTION_UNRESOLVED"
	}
	transitions := []contract.ClaimTransition{{ClaimID: definition.ID + ".description", Event: descriptionEvent, Before: "UNRECORDED", After: descriptionAfter, Coordinate: contract.Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: observation.Reason}, EvidenceDigest: attempt.DescriptionDigest}}
	after, event := "DENIED", "AUTHORIZATION_REJECTED"
	if observation.Decision == decisionOpen {
		after, event = "UNKNOWN", "AUTHORIZATION_UNRESOLVED"
	} else if observation.Decision == decisionRefuted {
		after, event = "REFUTED", "CAPABILITY_EVIDENCE_REFUTED"
	} else if observation.Authorization == authorizationGrant {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	transitions = append(transitions, contract.ClaimTransition{ClaimID: definition.ID + ".authorization", Event: event, Before: "UNRECORDED", After: after, Coordinate: contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Decision == decisionOpen {
		executionAfter, executionEvent = "UNKNOWN", "EXECUTION_UNRESOLVED"
	} else if observation.Decision == decisionRefuted {
		executionAfter, executionEvent = "REFUTED", "EXECUTION_REFUTED"
	} else if observation.Execution == executionAllowed {
		executionAfter, executionEvent = "EXECUTED_READ_ONLY", "EXECUTION_ALLOWED"
	}
	return append(transitions, contract.ClaimTransition{ClaimID: definition.ID + ".execution", Event: executionEvent, Before: "NOT_EXECUTED", After: executionAfter, Coordinate: contract.Coordinate{Stage: "SEAL_RECEIPT", Step: "seal-read-only-result", Reason: observation.Reason}})
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

func buildIndicators(summary contract.Summary) []contract.Indicator {
	return []contract.Indicator{
		indicator("gooo.metric.meta-circular-boundary.fixed-cases.v1", "DRIVER", proofFoundation, "bind-fixed-denominator", "cases", relationEqual, summary.CasesPassed, caseTotal),
		indicator("gooo.metric.meta-circular-boundary.description-binding-bps.v1", "DRIVER", proofFoundation, "bind-self-description", "basis_points", relationGreaterOrEqual, summary.DescriptionBound*10_000/caseTotal, 10_000),
		indicator("gooo.metric.meta-circular-boundary.description-authority-escalation.v1", "GUARDRAIL", proofRegression, "deny-description-authority-escalation", "paths", relationLessOrEqual, summary.DescriptionEscalationPaths, 0),
		indicator("gooo.metric.meta-circular-boundary.explicit-authorization.v1", "DRIVER", proofCoherence, "accept-explicit-read-only-capability", "cases", relationEqual, summary.ExplicitAuthorizations, 1),
		indicator("gooo.metric.meta-circular-boundary.authorized-execution.v1", "OUTCOME", proofCoherence, "permit-read-only-execution", "cases", relationEqual, summary.AllowedExecutions, 1),
		indicator("gooo.metric.meta-circular-boundary.forged-rejection.v1", "GUARDRAIL", proofRegression, "reject-forged-capability", "cases", relationEqual, summary.ForgedAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.scope-rejection.v1", "GUARDRAIL", proofRegression, "reject-out-of-scope-capability", "cases", relationEqual, summary.OutOfScopeAuthorizationsBlocked, 1),
		indicator("gooo.metric.meta-circular-boundary.replay.v1", "DRIVER", proofCoherence, "replay-independent-judge", "cases", relationEqual, summary.ReplayMatches, caseTotal),
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
