package languageconcept

import (
	"io/fs"
	"reflect"
)

func BuildArtifact(repository fs.FS) Artifact {
	return buildArtifact(repository, Catalog())
}

func buildArtifact(repository fs.FS, concepts []Concept) Artifact {
	report := Evaluate(repository, concepts)
	replay := Evaluate(repository, concepts)
	bindings := observeBindings(repository, concepts)
	artifact := Artifact{Schema: ArtifactSchema}
	artifact.Producer = "languageconcept.BuildArtifact"
	artifact.Consumer = "self-improvement-cycle"
	artifact.MetaOperation = "bind-language-concept-artifact"
	artifact.CatalogSource = CatalogSourcePath
	artifact.CatalogDigest = digest(concepts)
	artifact.Report = report
	artifact.ReplayReportDigest = replay.ReportDigest
	artifact.ReplayEqual = reflect.DeepEqual(report, replay)
	artifact.Bindings = bindings
	artifact.RepositoryWrites = 0
	artifact.Decision, artifact.Reason = artifactDecision(artifact)
	artifact.ArtifactDigest = artifactDigest(artifact)
	return artifact
}

func artifactDecision(artifact Artifact) (string, string) {
	switch {
	case artifact.Report.Decision != "PASS":
		return "FAIL_CLOSED", "LANGUAGE_CONCEPT_ARTIFACT_REPORT_FAILED"
	case !artifact.ReplayEqual:
		return "FAIL_CLOSED", "LANGUAGE_CONCEPT_ARTIFACT_REPLAY_DIVERGED"
	case artifact.Bindings.Missing != 0 || artifact.Bindings.Unsupported != 0:
		return "FAIL_CLOSED", "LANGUAGE_CONCEPT_ARTIFACT_BINDING_UNAVAILABLE"
	case artifact.Bindings.Files == 0 || artifact.RepositoryWrites != 0:
		return "FAIL_CLOSED", "LANGUAGE_CONCEPT_ARTIFACT_EFFECT_UNBOUND"
	default:
		return "PASS", "LANGUAGE_CONCEPT_ARTIFACT_READY"
	}
}

func artifactDigest(artifact Artifact) string {
	artifact.ArtifactDigest = ""
	return digest(artifact)
}
