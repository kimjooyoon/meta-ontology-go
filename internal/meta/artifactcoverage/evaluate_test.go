package artifactcoverage

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func TestEvaluateSelectsCanonicalGapDeterministically(t *testing.T) {
	action, observations := coverageFixture()
	report, err := Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil { t.Fatal(err) }
	if report.Decision != "FIXED_POINT" || report.Summary.CanonicalOperations != 5 ||
		report.Summary.CanonicalCoverageBasisPoints != 10000 { t.Fatalf("fixed point report = %#v", report) }
	first, _ := Marshal(report)
	replay, err := Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil { t.Fatal(err) }
	second, _ := Marshal(replay)
	if string(first) != string(second) { t.Fatal("coverage report replay differs") }
	filtered := observations.Artifacts[:0]
	for _, artifact := range observations.Artifacts {
		if !strings.HasPrefix(artifact.Name, "directory-kind-separation-") { filtered = append(filtered, artifact) }
	}
	observations.Artifacts = filtered
	report, err = Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil { t.Fatal(err) }
	if report.Decision != "IMPROVE" || report.SelectedOperation != "separate-directory-kinds" {
		t.Fatalf("gap decision = %s selected = %s", report.Decision, report.SelectedOperation)
	}
}

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
		artifact.Digest, artifact.ReplayDigest = "sha256:"+strings.Repeat("1", 64), "sha256:"+strings.Repeat("2", 64)
		artifact.EvidenceKeys = append(artifact.EvidenceKeys, binding.EvidenceKey)
		artifacts[name] = artifact
	}
	observations := ObservationDocument{Schema: ObservationSchema, CommitSHA: head,
		Repository: action.Repository, RunID: 1, RunAttempt: 1}
	for _, artifact := range artifacts { observations.Artifacts = append(observations.Artifacts, artifact) }
	sort.Slice(observations.Artifacts, func(i, j int) bool { return observations.Artifacts[i].Name < observations.Artifacts[j].Name })
	return action, observations
}
