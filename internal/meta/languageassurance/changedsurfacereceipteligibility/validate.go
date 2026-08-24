package changedsurfacereceipteligibility

import (
	"fmt"
	"reflect"
)

func bindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{binding(input.Assurance), binding(input.ShadowReport), binding(input.ShadowSuite)}
}

func binding(capsule Capsule) ArtifactBinding {
	observed := digestBytes(capsule.Payload)
	return ArtifactBinding{Name: capsule.Name, ArtifactID: capsule.ArtifactID,
		ArtifactDigest: capsule.ArtifactDigest, CapsuleDigest: capsule.CapsuleDigest,
		ObservedDigest: observed, Exact: observed == capsule.CapsuleDigest}
}

func exactDigests(input Input) bool {
	return input.Assurance.ArtifactID == AssuranceArtifactID && input.Assurance.ArtifactDigest == AssuranceArtifactDigest &&
		input.Assurance.CapsuleDigest == AssuranceCapsuleDigest && digestBytes(input.Assurance.Payload) == AssuranceCapsuleDigest &&
		input.ShadowReport.ArtifactID == ShadowArtifactID && input.ShadowReport.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowReport.CapsuleDigest == ShadowReportCapsuleDigest && digestBytes(input.ShadowReport.Payload) == ShadowReportCapsuleDigest &&
		input.ShadowSuite.ArtifactID == ShadowArtifactID && input.ShadowSuite.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowSuite.CapsuleDigest == ShadowSuiteCapsuleDigest && digestBytes(input.ShadowSuite.Payload) == ShadowSuiteCapsuleDigest
}

func validSemantics(assurance assuranceCapsule, report shadowReportCapsule, suite shadowSuiteCapsule) bool {
	return assurance.Schema == "gooo/language-assurance-report/v1" && assurance.SubjectSHA == EvidenceSubjectSHA &&
		assurance.DenominatorDigest == "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" &&
		assurance.Summary.DenominatorTotal == 12 && assurance.Summary.Operating == 8 && assurance.Summary.NotImplemented == 4 &&
		assurance.Summary.ImplementationCoverageBPS == 6666 && assurance.Summary.UnresolvedIndicators == 0 &&
		assurance.Summary.ViolatedGuardrails == 0 && assurance.Summary.RepositoryWrites == 0 && assuranceBaseline(assurance) &&
		report.Schema == "gooo/changed-surface-receipt-report/v1" && report.SubjectSHA == EvidenceSubjectSHA &&
		report.DenominatorID == "gooo/changed-surface-receipt-denominator/v1" && report.DenominatorDigest ==
		"sha256:61982bc0b87ba06031219143a0a67e69ea0b297bed75ee626e16077495d28499" &&
		report.Decision == "FIXED_POINT" && report.Resolution == "EXACT" && report.EnforcementEffect == "OBSERVE_ONLY" &&
		report.Summary.ChangedSurfaces == 2 && report.Summary.ReceiptsObserved == 2 && report.Summary.BoundReceipts == 2 &&
		report.Summary.TotalityBPS == 10_000 && report.Summary.UnknownPaths == 0 && report.Summary.BlockedPaths == 0 &&
		report.RepositoryWrites == 0 && hasMetaOperation(report) && suite.Schema == "gooo/changed-surface-receipt-conformance/v1" &&
		suite.SubjectSHA == EvidenceSubjectSHA && suite.DenominatorID == report.DenominatorID &&
		suite.DenominatorDigest == report.DenominatorDigest && suite.Decision == "FIXED_POINT" && suite.Resolution == "EXACT" &&
		suite.CasesTotal == 6 && suite.CasesPassed == 6 && suite.CoverageBPS == 10_000
}

func assuranceBaseline(capsule assuranceCapsule) bool {
	for _, obligation := range capsule.Obligations {
		if obligation.MetricID == MetricID {
			return obligation.Status == "NOT_IMPLEMENTED" && obligation.Resolution == "NONE"
		}
	}
	return false
}

func hasMetaOperation(report shadowReportCapsule) bool {
	for _, operation := range report.MetaOperations {
		if operation.ID == MetaOperation && operation.ProofChoice == "COHERENCE" {
			return true
		}
	}
	return false
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID || !reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("changed surface receipt eligibility report does not replay")
	}
	return nil
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("changed surface receipt eligibility suite does not replay")
	}
	return nil
}
