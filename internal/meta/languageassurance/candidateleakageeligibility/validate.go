package candidateleakageeligibility

import (
	"fmt"
	"reflect"
)

func bindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{binding(input.Assurance), binding(input.Shadow)}
}

func binding(capsule Capsule) ArtifactBinding {
	observed := digestBytes(capsule.Payload)
	return ArtifactBinding{Name: capsule.Name, ArtifactID: capsule.ArtifactID,
		ArtifactDigest: capsule.ArtifactDigest, CapsuleDigest: capsule.CapsuleDigest,
		ObservedDigest: observed, Exact: observed == capsule.CapsuleDigest}
}

func exactDigests(input Input) bool {
	return input.Assurance.ArtifactID == AssuranceArtifactID &&
		input.Assurance.ArtifactDigest == AssuranceArtifactDigest &&
		input.Assurance.CapsuleDigest == AssuranceCapsuleDigest &&
		input.Shadow.ArtifactID == ShadowArtifactID && input.Shadow.ArtifactDigest == ShadowArtifactDigest &&
		input.Shadow.CapsuleDigest == ShadowCapsuleDigest &&
		digestBytes(input.Assurance.Payload) == AssuranceCapsuleDigest &&
		digestBytes(input.Shadow.Payload) == ShadowCapsuleDigest
}

func validSemantics(assurance assuranceCapsule, shadow shadowCapsule) bool {
	return assurance.Schema == "gooo/language-assurance-report/v1" &&
		assurance.SubjectSHA == EvidenceSubjectSHA &&
		assurance.DenominatorDigest == "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" &&
		assurance.Summary.DenominatorTotal == 12 && assurance.Summary.Operating == 7 &&
		assurance.Summary.NotImplemented == 5 && assurance.Summary.ImplementationCoverageBPS == 5833 &&
		candidateBaseline(assurance) && shadow.Schema == "gooo/candidate-leakage-conformance/v1" &&
		shadow.SubjectSHA == EvidenceSubjectSHA && shadow.DenominatorDigest ==
		"sha256:7053f28b9b0619d98694dbc04fef30d2a29589fcd7c959b8c2ec4887f43307da" &&
		shadow.Decision == "PASS" && shadow.Resolution == "EXACT" && shadow.Summary.CasesTotal == 6 &&
		shadow.Summary.CasesPassed == 6 && shadow.Summary.CoverageBPS == 10_000 &&
		shadow.RepositoryWrites == 0 && shadow.PromotionCreditBPS == 0
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID ||
		!reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("candidate leakage eligibility report does not replay")
	}
	return nil
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("candidate leakage eligibility suite does not replay")
	}
	return nil
}
