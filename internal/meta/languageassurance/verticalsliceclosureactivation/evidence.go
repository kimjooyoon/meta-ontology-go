package verticalsliceclosureactivation

import _ "embed"

//go:embed evidence/assurance.json
var assuranceCapsule []byte

//go:embed evidence/eligibility.json
var eligibilityCapsule []byte

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA, Assurance: append([]byte(nil), assuranceCapsule...), Eligibility: append([]byte(nil), eligibilityCapsule...)}
}

func artifactBindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{
		artifact("language-assurance", 9525256269,
			"sha256:e640b072baafdfe59e504edaabca377588176c6ff76bcbeb4efb95f496d1f831", AssuranceCapsuleHash, input.Assurance),
		artifact("vertical-slice-closure-eligibility", 9525476954,
			"sha256:b0a9b2262eef5562b6d2c3cdebfc98c1c69928772d8a319eb4ea9947a0ec0c65", EligibilityCapsuleHash, input.Eligibility),
	}
}

func artifact(name string, id int64, archive, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, ArtifactID: id, ArtifactDigest: archive,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
