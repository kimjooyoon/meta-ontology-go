package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
	closureverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure/verify"
)

func verifierInput(input closure.Input) closureverify.Input {
	return closureverify.Input{
		Repository: input.Repository, SubjectSHA: input.SubjectSHA,
		RunID: input.RunID, RunAttempt: input.RunAttempt,
		Artifact: closureverify.ArtifactIdentity{
			ID: input.Artifact.ID, Name: input.Artifact.Name,
			Digest: input.Artifact.Digest, URL: input.Artifact.URL,
		},
		ProgramJSON: input.ProgramJSON, Source: input.Source,
		VerificationJSON: input.VerificationJSON,
	}
}
