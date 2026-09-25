package rollbackintegrityeligibility

import (
	"fmt"
	"reflect"
)

func bindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{binding(input.Assurance), binding(input.ShadowReportA), binding(input.ShadowReportB)}
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
		input.ShadowReportA.ArtifactID == ShadowArtifactID && input.ShadowReportA.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowReportA.CapsuleDigest == ShadowReportACapsuleDigest &&
		digestBytes(input.ShadowReportA.Payload) == ShadowReportACapsuleDigest &&
		input.ShadowReportB.ArtifactID == ShadowArtifactID && input.ShadowReportB.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowReportB.CapsuleDigest == ShadowReportBCapsuleDigest &&
		digestBytes(input.ShadowReportB.Payload) == ShadowReportBCapsuleDigest
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID || !reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("rollback integrity eligibility report does not replay")
	}
	return nil
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("rollback integrity eligibility suite does not replay")
	}
	return nil
}
