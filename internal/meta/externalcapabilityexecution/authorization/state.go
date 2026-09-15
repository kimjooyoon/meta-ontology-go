package authorization

import capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"

type metricState struct {
	known     bool
	satisfied bool
	reason    string
	evidence  string
}

func reportState(input Input) metricState {
	if !input.EnvelopeAvailable || !input.ReportAvailable {
		return metricState{reason: "EXECUTION_REPORT_UNAVAILABLE"}
	}
	report := input.Report
	decisionKnown := report.Decision == capability.DecisionExecutable ||
		report.Decision == capability.DecisionFailClosed
	resolutionKnown := report.Resolution == capability.ResolutionExact ||
		report.Resolution == capability.ResolutionUnknown ||
		report.Resolution == capability.ResolutionInvariant
	if !decisionKnown || !resolutionKnown {
		return metricState{reason: "EXECUTION_REPORT_DECISION_UNKNOWN"}
	}
	digest := sealedReportDigest(report)
	exact := report.ReportDigest == digest && input.Envelope.SourceReportDigest == digest
	exact = exact && report.Decision == capability.DecisionExecutable &&
		report.Resolution == capability.ResolutionExact && report.Completed == 10 &&
		report.Total == 10 && report.BasisPoints == 10000 && report.UnknownIndicators == 0
	exact = exact && report.RepositoryWrites == 0 && report.ExternalRepositoryWrites == 0 &&
		report.OfficialMutationCount == 0 && report.PromotionCount == 0
	return metricState{known: true, satisfied: exact, evidence: report.ReportDigest}
}

func policyState(input Input) metricState {
	if !input.Policy.SourceAvailable {
		return metricState{reason: "POLICY_SOURCE_UNAVAILABLE"}
	}
	if !input.Policy.GeneratedAvailable {
		return metricState{reason: "POLICY_GENERATED_ARTIFACT_UNAVAILABLE"}
	}
	if !input.Foundation.Available {
		return metricState{reason: "POLICY_FOUNDATION_UNAVAILABLE"}
	}
	foundation := input.Foundation
	known := foundation.Schema != "" && foundation.SubjectSHA != "" &&
		foundation.ProducerRunID != "" && foundation.ArtifactID > 0 &&
		validDigest(foundation.ArchiveDigest) && validDigest(foundation.PolicySourceDigest) &&
		validDigest(foundation.PolicyGeneratedDigest)
	if !known {
		return metricState{reason: "POLICY_FOUNDATION_COORDINATES_UNKNOWN"}
	}
	exact := foundation.Schema == FoundationSchema && foundation.SubjectSHA == input.Invocation.SubjectSHA
	exact = exact && foundation.PolicySourceDigest == input.Policy.SourceDigest &&
		foundation.PolicyGeneratedDigest == input.Policy.GeneratedDigest
	exact = exact && input.Envelope.PolicySourceDigest == input.Policy.SourceDigest &&
		input.Envelope.PolicyGeneratedDigest == input.Policy.GeneratedDigest
	return metricState{known: true, satisfied: exact, evidence: foundation.ArchiveDigest}
}
