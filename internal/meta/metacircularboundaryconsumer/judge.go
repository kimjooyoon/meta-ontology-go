package metacircularboundaryconsumer

import (
	"fmt"
	"reflect"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

// Judge is the consumer-side gate. It independently derives cases, expected
// outcomes, receipts, summaries, and indicators without importing producer.
func Judge(report contract.Report, input contract.Input) error {
	if report.Schema != reportSchema || report.Scope != scope || !validSHA(input.HeadSHA) || report.HeadSHA != input.HeadSHA || input.Path != reportSourcePath() || report.Source.Path != input.Path || report.ReportDigest != sealReport(report).ReportDigest {
		return fmt.Errorf("%s", reasonMismatch)
	}
	observed, err := observeSource(input.Path, input.Source)
	if err != nil || !reflect.DeepEqual(observed, report.Source) {
		return fmt.Errorf("%s", reasonSource)
	}
	if report.Decision != decisionPass || report.Resolution != resolutionExact || report.Reason != reasonContract || len(report.Cases) != caseTotal || len(report.Receipts) != caseTotal || len(report.Indicators) != indicatorTotal || !reflect.DeepEqual(report.MetaOperations, expectedMetaOperations()) || !reflect.DeepEqual(report.NotClaimed, notClaimed()) || report.MetaValue != metaValue {
		return fmt.Errorf("%s", reasonMismatch)
	}

	expected := expectedCases(observed.SourceDigest)
	for index, definition := range expected {
		attempt, ok := expectedAttempt(definition.ID, observed.SourceDigest)
		if !ok || !reflect.DeepEqual(report.Cases[index].Definition, definition) || !reflect.DeepEqual(report.Cases[index].Attempt, attempt) {
			return fmt.Errorf("%s", reasonMismatch)
		}
		want := classify(observed, attempt)
		if !reflect.DeepEqual(report.Cases[index].Observation, want) || !report.Cases[index].Passed || !matches(definition, want) {
			return fmt.Errorf("%s", reasonMismatch)
		}
		wantReceipt := receiptFor(observed, definition, attempt, want)
		if !reflect.DeepEqual(report.Cases[index].Receipt, wantReceipt) || !reflect.DeepEqual(report.Cases[index].Receipt, report.Receipts[index]) {
			return fmt.Errorf("%s", reasonMismatch)
		}
	}

	summary := summarize(report.Cases)
	if !reflect.DeepEqual(report.Summary, summary) || !reflect.DeepEqual(report.Indicators, buildIndicators(summary)) || report.RepositoryWrites != 0 || report.MutationAuthority || report.IndependentJudge != (contract.JudgeEvidence{Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ComparedCases: caseTotal, Mismatches: 0, Decision: decisionPass, Reason: reasonContract}) {
		return fmt.Errorf("%s", reasonMismatch)
	}
	return nil
}

func reportSourcePath() string { return "examples/meta-circular-boundary/main.gooo" }

func classify(source contract.SourceObservation, attempt contract.Attempt) contract.CaseObservation {
	bound := source.DescriptionBound && attempt.DescriptionDigest == source.SourceDigest
	observation := contract.CaseObservation{Description: descriptionBound, Authorization: authorizationDeny, Execution: executionBlocked, RepositoryWrites: 0, MutationAuthority: false}
	if !bound {
		observation.Description = "UNKNOWN"
		observation.Reason = reasonSource
		return observation
	}
	if attempt.Capability == nil {
		observation.Reason = reasonDescription
		return observation
	}
	if attempt.Capability.Scope == scopeWrite {
		observation.Reason = reasonOutOfScope
		return observation
	}
	grant := attempt.Capability
	if grant.Issuer != "external-authority" || grant.SubjectDigest != source.SourceDigest || grant.Operation != metaOperationID || grant.Scope != scopeReadOnly || grant.Handle != capabilityHandle(source.SourceDigest) {
		observation.Reason = reasonForged
		return observation
	}
	observation.Authorization, observation.Reason = authorizationGrant, reasonExplicit
	if attempt.RequestExecution {
		observation.Execution = executionAllowed
	}
	return observation
}

func matches(definition contract.CaseDefinition, observation contract.CaseObservation) bool {
	return observation.Description == descriptionBound && observation.Authorization == definition.ExpectedAuthorization && observation.Execution == definition.ExpectedExecution && observation.Reason == definition.ExpectedReason
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
		Decision: definition.ExpectedDecision, Authorization: observation.Authorization, Execution: observation.Execution,
		RepositoryWrites: observation.RepositoryWrites, MutationAuthority: observation.MutationAuthority,
		ClaimTransitions: claimTransitions(definition, attempt, observation),
	}
	return sealReceipt(receipt)
}

func claimTransitions(definition contract.CaseDefinition, attempt contract.Attempt, observation contract.CaseObservation) []contract.ClaimTransition {
	transitions := []contract.ClaimTransition{{ClaimID: definition.ID + ".description", Event: "DESCRIPTION_OBSERVED", Before: "UNRECORDED", After: "DESCRIBED", Coordinate: contract.Coordinate{Stage: "PARSE_AST", Step: "observe-description", Reason: "SELF_DESCRIPTION_OBSERVED"}, EvidenceDigest: attempt.DescriptionDigest}}
	after, event := "DENIED", "AUTHORIZATION_REJECTED"
	if observation.Authorization == authorizationGrant {
		after, event = "AUTHORIZED", "EXPLICIT_AUTHORIZATION_ACCEPTED"
	}
	transitions = append(transitions, contract.ClaimTransition{ClaimID: definition.ID + ".authorization", Event: event, Before: "UNRECORDED", After: after, Coordinate: contract.Coordinate{Stage: "REPLAY_AUTHORIZATION", Step: "judge-capability", Reason: observation.Reason}})
	executionAfter, executionEvent := "BLOCKED", "EXECUTION_BLOCKED"
	if observation.Execution == executionAllowed {
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
		switch item.Definition.ID {
		case "description-only":
			if item.Observation.Execution == executionBlocked {
				summary.DescriptionOnlyBlocked++
			}
		case "forged-capability":
			if item.Observation.Authorization == authorizationDeny {
				summary.ForgedAuthorizationsBlocked++
			}
		case "write-capability-out-of-scope":
			if item.Observation.Authorization == authorizationDeny {
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
