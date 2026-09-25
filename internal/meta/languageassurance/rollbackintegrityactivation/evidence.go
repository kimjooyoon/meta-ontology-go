package rollbackintegrityactivation

import _ "embed"

//go:embed evidence/assurance.json
var assuranceCapsule []byte

//go:embed evidence/eligibility.json
var eligibilityCapsule []byte

func EmbeddedInput(subjectSHA string) Input {
	return Input{
		SubjectSHA:  subjectSHA,
		Assurance:   append([]byte(nil), assuranceCapsule...),
		Eligibility: append([]byte(nil), eligibilityCapsule...),
	}
}

func artifactBindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{
		artifact("language-assurance", 9515629846,
			"sha256:547576f410897dc8d662c990e2827333405e305158bd75e269c1361aed82b9e9",
			AssuranceCapsuleHash, input.Assurance),
		artifact("rollback-integrity-eligibility", 9515667591,
			"sha256:ede353397fcabb56328181c0b78a293aba0d00e0cafcb9dc443565d6943e4fef",
			EligibilityCapsuleHash, input.Eligibility),
	}
}

func artifact(name string, id int64, archive, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, ArtifactID: id, ArtifactDigest: archive,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
