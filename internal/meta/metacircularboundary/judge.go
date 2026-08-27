package metacircularboundary

import (
	"fmt"
	"reflect"
)

// Judge re-reads the source and derives the expected four outcomes without
// trusting the producer's top-level decision. It is the consumer-side gate.
func Judge(report Report, input Input) error {
	if report.Schema != ReportSchema || report.Scope != Scope || !validSHA(input.HeadSHA) || report.HeadSHA != input.HeadSHA || input.Path != ExpectedSourcePath || report.Source.Path != input.Path || report.ReportDigest != sealReport(report).ReportDigest {
		return fmt.Errorf("%s", ReasonReportMismatch)
	}
	observed, err := observeSource(input.Path, input.Source)
	if err != nil || !reflect.DeepEqual(observed, report.Source) {
		return fmt.Errorf("%s", ReasonSourceBindingUnknown)
	}
	denominator := DenominatorContract()
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact || report.Reason != ReasonContractSatisfied || report.Summary.CasesTotal != CaseTotal || len(report.Cases) != CaseTotal || len(report.Receipts) != CaseTotal || len(report.Indicators) != IndicatorTotal {
		return fmt.Errorf("%s", ReasonReportMismatch)
	}
	for index, definition := range denominator.Cases {
		attempt, ok := CaseInput(definition.ID, observed.SourceDigest)
		if !ok || !reflect.DeepEqual(report.Cases[index].Definition, definition) || !reflect.DeepEqual(report.Cases[index].Attempt, attempt) {
			return fmt.Errorf("%s", ReasonReportMismatch)
		}
		want := independentlyJudgeAttempt(observed, attempt)
		if !reflect.DeepEqual(report.Cases[index].Observation, want) || !report.Cases[index].Passed || !matches(definition, want) {
			return fmt.Errorf("%s", ReasonReportMismatch)
		}
		if !reflect.DeepEqual(report.Cases[index].Receipt, report.Receipts[index]) || report.Receipts[index].ReceiptDigest != sealReceipt(report.Receipts[index]).ReceiptDigest {
			return fmt.Errorf("%s", ReasonReportMismatch)
		}
	}
	if !reflect.DeepEqual(report.Summary, summarize(report.Cases)) || !reflect.DeepEqual(report.Indicators, buildIndicators(report.Summary)) || report.IndependentJudge.Mismatches != 0 || report.IndependentJudge.Decision != DecisionPass {
		return fmt.Errorf("%s", ReasonReportMismatch)
	}
	return nil
}

func independentlyJudgeAttempt(source SourceObservation, attempt Attempt) CaseObservation {
	observation := CaseObservation{
		Description: DescriptionBound, Authorization: AuthorizationDenied,
		Execution: ExecutionBlocked, RepositoryWrites: 0, MutationAuthority: false,
	}
	if !source.DescriptionBound || attempt.DescriptionDigest != source.SourceDigest {
		observation.Description = "UNKNOWN"
		observation.Reason = ReasonSourceBindingUnknown
		return observation
	}
	grant := attempt.Capability
	if grant == nil {
		observation.Reason = ReasonDescriptionOnly
		return observation
	}
	if grant.Scope == ScopeWrite {
		observation.Reason = ReasonOutOfScopeCapability
		return observation
	}
	bound := grant.Issuer == "external-authority" &&
		grant.SubjectDigest == source.SourceDigest &&
		grant.Operation == MetaOperationID &&
		grant.Scope == ScopeReadOnly &&
		grant.Handle == capabilityHandle(source.SourceDigest)
	if !bound {
		observation.Reason = ReasonForgedCapability
		return observation
	}
	observation.Authorization, observation.Reason = AuthorizationGranted, ReasonExplicitCapability
	if attempt.RequestExecution {
		observation.Execution = ExecutionAllowed
	}
	return observation
}
