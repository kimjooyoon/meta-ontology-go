package languageartifactoracle

import (
	"reflect"
	"strings"
)

func compareArtifact(want projection, artifact sourceArtifact, filename, sourceDigest string) []CheckResult {
	wantEvents := []artifactEvent{
		{1, "SOURCE_PARSED", sourceDigest},
		{2, "SEMANTIC_LOWERED", artifact.SemanticDigest},
		{3, "ACTIVITY_INVOKED", want.Activity},
		{4, "ENTITY_PRODUCED", want.Output.ID},
	}
	values := []bool{
		artifact.Digest == artifactDigest(artifact),
		artifact.SourceDigest == sourceDigest,
		artifact.Schema == SourceArtifactSchema && artifact.Decision == "PASS" &&
			artifact.Reason == "SOURCE_ACTIVITY_EXECUTED" && artifact.Resolution == "EXACT" && artifact.Filename == filename,
		artifact.Entry.Package == want.Package && artifact.Entry.Namespace == want.Namespace && artifact.Entry.Activity == want.Activity,
		reflect.DeepEqual(artifact.Entry.Inputs, want.Inputs),
		artifact.Entry.Output == want.Output,
		reflect.DeepEqual(artifact.Events, wantEvents),
		semanticCoherent(artifact),
		artifact.Effects.RepositoryWrites == 0 && !artifact.Effects.MutationAuthority,
	}
	result := make([]CheckResult, len(fixedChecks))
	for index, spec := range fixedChecks {
		status, observed := "FAIL", "false"
		if values[index] {
			status, observed = "PASS", "true"
		}
		result[index] = CheckResult{ID: spec.id, Status: status, ProofChoice: spec.proof,
			MetaOperation: spec.operation, Expected: "true", Observed: observed}
	}
	return result
}

func semanticCoherent(artifact sourceArtifact) bool {
	return strings.HasPrefix(artifact.SemanticDigest, "sha256:") && len(artifact.Events) == 4 &&
		artifact.Events[1].Subject == artifact.SemanticDigest
}

func firstFailedCheck(checks []CheckResult) string {
	for _, check := range checks {
		if check.Status == "FAIL" {
			return check.ID
		}
	}
	return ""
}
