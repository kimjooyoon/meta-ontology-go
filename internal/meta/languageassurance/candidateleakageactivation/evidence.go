package candidateleakageactivation

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
		artifact("language-assurance", 9509154504,
			"sha256:d6006f0d8db1bd12c1b527afa7bbf8a1bc2e60dbd7352c2665216fdd39f91a56",
			AssuranceCapsuleHash, input.Assurance),
		artifact("candidate-leakage-eligibility", 9509166817,
			"sha256:db7636e6c831afc1a68148c0cecc9f1d057532bda020f0562bf26328c8c39089",
			EligibilityCapsuleHash, input.Eligibility),
	}
}

func artifact(name string, id int64, archive, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, ArtifactID: id, ArtifactDigest: archive,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
