package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"

func newInput(value config, program, source, verification []byte) closure.Input {
	return closure.Input{
		Repository: value.repository, SubjectSHA: value.subjectSHA,
		RunID: value.runID, RunAttempt: value.runAttempt,
		Artifact: closure.ArtifactIdentity{
			ID: value.artifactID, Name: value.artifactName,
			Digest: value.artifactDigest, URL: value.artifactURL,
		},
		ProgramJSON: program, Source: source, VerificationJSON: verification,
	}
}
