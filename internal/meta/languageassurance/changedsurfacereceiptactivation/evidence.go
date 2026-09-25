package changedsurfacereceiptactivation

import _ "embed"

//go:embed evidence/assurance.json
var assuranceCapsule []byte

//go:embed evidence/eligibility.json
var eligibilityCapsule []byte

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA,
		Assurance:   append([]byte(nil), assuranceCapsule...),
		Eligibility: append([]byte(nil), eligibilityCapsule...)}
}

func artifactBindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{
		artifact("language-assurance", 9511928780,
			"sha256:99cc771dbcc9fdea6a3df534e63f5de5357efab5ed521590a3c0fec1ff4b229a",
			AssuranceCapsuleHash, input.Assurance),
		artifact("changed-surface-receipt-eligibility", 9511958307,
			"sha256:7bc661464f478d6d37d9fa231c03f5308a2c7e448086da592d21066777b6376c",
			EligibilityCapsuleHash, input.Eligibility),
	}
}

func artifact(name string, id int64, archive, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, ArtifactID: id, ArtifactDigest: archive,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
