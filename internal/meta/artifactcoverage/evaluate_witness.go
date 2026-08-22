package artifactcoverage

import (
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func evaluateOperations(action actionability.Report, observations ObservationDocument,
	bindings []ArtifactBinding,
) (Summary, []OperationWitness, string) {
	index := make(map[string]ArtifactBinding, len(bindings))
	for _, binding := range bindings {
		index[binding.Operation] = binding
	}
	operations := append([]actionability.OperationWitness(nil), action.Operations...)
	sort.Slice(operations, func(i, j int) bool { return operations[i].Operation < operations[j].Operation })
	summary := Summary{RepositoryWrites: observations.RepositoryWrites}
	witnesses, selected, selectedCount := make([]OperationWitness, 0, len(operations)), "", -1
	for _, operation := range operations {
		if !operation.Executable {
			continue
		}
		witness := observeOperation(operation, index[operation.Operation], observations)
		summary.RequiredOperations++
		if witness.ExactHead { summary.ExactHeadOperations++ }
		if witness.DigestBound { summary.DigestBoundOperations++ }
		if witness.ReplayBound { summary.ReplayBoundOperations++ }
		if witness.Canonical { summary.CanonicalOperations++ } else {
			summary.UncoveredOperations++
			if operation.IndicatorCount > selectedCount || operation.IndicatorCount == selectedCount && operation.Operation < selected {
				selected, selectedCount = operation.Operation, operation.IndicatorCount
			}
		}
		if witness.ObservedArtifacts > 1 { summary.AmbiguousOperations++ }
		witnesses = append(witnesses, witness)
	}
	summary.CanonicalCoverageBasisPoints = coverage(summary.CanonicalOperations, summary.RequiredOperations)
	summary.ExactHeadCoverageBasisPoints = coverage(summary.ExactHeadOperations, summary.RequiredOperations)
	summary.DigestBoundCoverageBasisPoints = coverage(summary.DigestBoundOperations, summary.RequiredOperations)
	summary.ReplayBoundCoverageBasisPoints = coverage(summary.ReplayBoundOperations, summary.RequiredOperations)
	return summary, witnesses, selected
}

func observeOperation(operation actionability.OperationWitness, binding ArtifactBinding,
	document ObservationDocument,
) OperationWitness {
	witness := OperationWitness{Operation: operation.Operation, IndicatorCount: operation.IndicatorCount}
	if binding.Operation == "" {
		witness.Status = "ARTIFACT_BINDING_MISSING"
		return witness
	}
	witness.ArtifactPattern, witness.EvidenceKey = binding.ArtifactPattern, binding.EvidenceKey
	witness.ExpectedArtifact = strings.ReplaceAll(binding.ArtifactPattern, "{head_sha}", document.CommitSHA)
	matches := make([]ArtifactObservation, 0, 1)
	for _, artifact := range document.Artifacts {
		if slices.Contains(artifact.EvidenceKeys, binding.EvidenceKey) { matches = append(matches, artifact) }
	}
	witness.ObservedArtifacts = len(matches)
	if len(matches) != 1 {
		if len(matches) == 0 { witness.Status = "ARTIFACT_MISSING" } else { witness.Status = "ARTIFACT_AMBIGUOUS" }
		return witness
	}
	artifact := matches[0]
	witness.ExactHead = artifact.Name == witness.ExpectedArtifact && artifact.HeadSHA == document.CommitSHA
	witness.DigestBound, witness.ReplayBound = validDigest(artifact.Digest), validDigest(artifact.ReplayDigest)
	witness.Canonical = witness.ExactHead && witness.DigestBound && witness.ReplayBound
	switch { case !witness.ExactHead: witness.Status = "EXACT_HEAD_MISMATCH"; case !witness.DigestBound:
		witness.Status = "DIGEST_UNBOUND"; case !witness.ReplayBound: witness.Status = "REPLAY_UNBOUND"; default:
		witness.Status = "CANONICAL" }
	return witness
}

func coverage(covered, total int) int { if total == 0 { return 10000 }; return covered * 10000 / total }
