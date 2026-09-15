package toolchainusecases

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

const invalidDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func Evaluate(repository fs.FS, headSHA string, registryRaw []byte,
	artifact languageconcept.Artifact) Report {
	registry, registryErr := decodeRegistry(registryRaw)
	registryEvidence := digestBytes(registryRaw)
	if registryErr == nil {
		registryEvidence = registryDigest()
	}
	report := Report{
		Schema: ReportSchema, Producer: "toolchainusecases.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: "execute-versioned-use-cases",
		HeadSHA: headSHA, RepositoryWrites: 0, MutationAuthorized: false,
		Source: Source{ExpectedHeadSHA: headSHA, ConceptArtifactDigest: artifact.ArtifactDigest,
			CatalogDigest: artifact.CatalogDigest, RegistryDigest: registryEvidence,
			ConceptRepositoryWrites: artifact.RepositoryWrites},
	}
	if registryErr != nil {
		report.Cases = unresolvedCases(artifact.ArtifactDigest)
		return finish(report)
	}
	for _, definition := range registry.Cases {
		observed := execute(repository, artifact, definition.Mutation)
		status := "NOT_SATISFIED"
		if observed == definition.ExpectedDecision {
			status = "SATISFIED"
		}
		item := CaseResult{Definition: definition, ObservedDecision: observed, Status: status}
		item.EvidenceDigest = caseDigest(item, artifact.ArtifactDigest)
		report.Cases = append(report.Cases, item)
	}
	return finish(report)
}

func execute(repository fs.FS, artifact languageconcept.Artifact, mutation string) string {
	candidate := artifact
	switch mutation {
	case "NONE":
	case "REPLAY_DIGEST":
		candidate.ReplayReportDigest = invalidDigest
	case "REPOSITORY_WRITE":
		candidate.RepositoryWrites = 1
	default:
		return "UNKNOWN"
	}
	if languageconcept.ValidateArtifact(repository, candidate) == nil {
		return DecisionPass
	}
	return DecisionClosed
}
