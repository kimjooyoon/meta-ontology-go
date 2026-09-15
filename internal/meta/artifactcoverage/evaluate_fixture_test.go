package artifactcoverage

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func coverageFixture() (actionability.Report, ObservationDocument) {
	head := strings.Repeat("a", 40)
	action := actionability.Report{Schema: actionability.Schema, CommitSHA: head,
		Repository: "kimjooyoon/meta-ontology-go", Decision: "FIXED_POINT"}
	artifacts := make(map[string]ArtifactObservation)
	for index, binding := range CanonicalBindings() {
		action.Operations = append(action.Operations, actionability.OperationWitness{
			Operation: binding.Operation, IndicatorCount: index + 1, Executable: true})
		name := strings.ReplaceAll(binding.ArtifactPattern, "{head_sha}", head)
		artifact := artifacts[name]
		artifact.Name, artifact.HeadSHA = name, head
		artifact.Digest = "sha256:" + strings.Repeat("1", 64)
		artifact.ReplayDigest = "sha256:" + strings.Repeat("2", 64)
		artifact.EvidenceKeys = append(artifact.EvidenceKeys, binding.EvidenceKey)
		artifacts[name] = artifact
	}
	observations := ObservationDocument{Schema: ObservationSchema, CommitSHA: head,
		Repository: action.Repository, RunID: 1, RunAttempt: 1}
	for _, artifact := range artifacts {
		observations.Artifacts = append(observations.Artifacts, artifact)
	}
	sort.Slice(observations.Artifacts, func(i, j int) bool {
		return observations.Artifacts[i].Name < observations.Artifacts[j].Name
	})
	return action, observations
}
