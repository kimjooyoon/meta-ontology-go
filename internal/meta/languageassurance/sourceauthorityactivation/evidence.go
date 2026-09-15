package sourceauthorityactivation

import _ "embed"

//go:embed evidence/assurance.json
var assuranceCapsule []byte

//go:embed evidence/upstream.json
var upstreamCapsule []byte

//go:embed evidence/eligibility.json
var eligibilityCapsule []byte

func EmbeddedInput(subjectSHA string) Input {
	return Input{
		SubjectSHA:  subjectSHA,
		Assurance:   append([]byte(nil), assuranceCapsule...),
		Upstream:    append([]byte(nil), upstreamCapsule...),
		Eligibility: append([]byte(nil), eligibilityCapsule...),
	}
}

func artifactBindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{
		artifact("language-assurance", 9506929857, "sha256:3cb628f79e51d034f923c182dbf821acd8b8e705cd8ee573e877a92077071579", AssuranceCapsuleHash, input.Assurance),
		artifact("source-authority-upstream", 9506932581, "sha256:fdf911ebaee84c3d3e5c27f12e72e48d1f359fcdd78f6d34cce61cb96078501d", UpstreamCapsuleHash, input.Upstream),
		artifact("source-authority-promotion-eligibility", 9506943519, "sha256:ad65e902ff07ddd3e30c48e0155769806b59bf37c12e66882b9a1ea3358d0068", EligibilityCapsuleHash, input.Eligibility),
	}
}

func artifact(name string, id int64, artifactDigest, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, ArtifactID: id, ArtifactDigest: artifactDigest,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
