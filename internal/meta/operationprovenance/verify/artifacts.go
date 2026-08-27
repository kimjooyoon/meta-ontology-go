package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var relationKinds = []string{"PRODUCES", "CONSUMES", "OPERATES", "EVIDENCED_BY"}

func collectArtifacts(root string, metrics []cMetric) map[string][]relationObservation {
	all := map[string][]relationObservation{}
	for _, metric := range metrics {
		for _, relation := range relationKinds {
			all[metric.id] = append(all[metric.id], observeArtifact(root, metric, relation))
		}
	}
	return all
}

func observeArtifact(root string, metric cMetric, relation string) relationObservation {
	relative := relativePath(metric, relation)
	declared := endpoint(metric, relation)
	observation := relationObservation{Relation: relation, DeclaredEndpoint: declared, ObservedArtifact: filepath.ToSlash(relative), EvidenceKind: evidenceKind(relation), RelationStatus: "UNKNOWN", Stage: "ARTIFACT", Step: "read-artifact", Cause: "DIRECT_CAUSE"}
	if !safeRelative(relative) {
		observation.Step, observation.Reason = "resolve-artifact-path", "INVALID_ARTIFACT_PATH"
		return observation
	}
	payload, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		observation.Reason = "REQUIRED_RAW_ARTIFACT_MISSING"
		return observation
	}
	observation.ObservedDigest = digest(payload)
	var artifact artifactEnvelope
	if err := json.Unmarshal(payload, &artifact); err != nil {
		observation.Step, observation.Reason = "parse-artifact", "MALFORMED_RAW_ARTIFACT"
		return observation
	}
	if !validArtifact(metric, relation, declared, artifact) {
		observation.Step, observation.Reason = "validate-artifact", "RAW_ARTIFACT_DOES_NOT_PROVE_RELATION"
		return observation
	}
	observation.ObservedEndpoint = artifact.Endpoint
	observation.RelationStatus, observation.Step, observation.Reason, observation.Cause = "PASS", "reconstruct-edge", "OBSERVED_ARTIFACT_MATCHES_DECLARATION", "OBSERVED_EVIDENCE"
	return observation
}
