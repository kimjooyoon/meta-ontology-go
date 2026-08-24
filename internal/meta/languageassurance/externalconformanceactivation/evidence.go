package externalconformanceactivation

import _ "embed"

//go:embed evidence/assurance.json
var assuranceCapsule []byte

//go:embed evidence/eligibility.json
var eligibilityCapsule []byte

//go:embed evidence/merge.json
var mergeCapsule []byte

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA, Assurance: append([]byte(nil), assuranceCapsule...),
		Eligibility: append([]byte(nil), eligibilityCapsule...), Merge: append([]byte(nil), mergeCapsule...)}
}

func artifactBindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{
		artifact("language-assurance", "github-actions:32763887778", 9533711537,
			"sha256:b901dcc9827ee20c28fa268155bf0766347db08437c7d75bb7ee23c3b3e14eed", AssuranceCapsuleHash, input.Assurance),
		artifact("external-conformance-eligibility", "github-actions:32762969070", 9533411143,
			"sha256:a147f4f4b124a927cb3e0679155af8b618e68088797dc9cd30bb250631456e5c", EligibilityCapsuleHash, input.Eligibility),
		artifact("pull-request-merge", "github-pull-request:474", 474,
			MergeCapsuleHash, MergeCapsuleHash, input.Merge),
	}
}

func artifact(name, source string, id int64, archive, expected string, raw []byte) ArtifactBinding {
	observed := digestBytes(raw)
	return ArtifactBinding{Name: name, SourceURI: source, ArtifactID: id, ArtifactDigest: archive,
		CapsuleDigest: expected, ObservedDigest: observed, Bytes: len(raw), Exact: observed == expected}
}
