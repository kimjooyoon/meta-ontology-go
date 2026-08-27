package operationprovenance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var relationKinds = []string{"PRODUCES", "CONSUMES", "OPERATES", "EVIDENCED_BY"}

func collectArtifacts(root string, metrics []metricSpec) (map[string][]RelationObservation, error) {
	all := make(map[string][]RelationObservation, len(metrics))
	for _, metric := range metrics {
		observations := make([]RelationObservation, 0, len(relationKinds))
		for _, kind := range relationKinds {
			observations = append(observations, observeArtifact(root, metric, kind))
		}
		all[metric.ID] = observations
	}
	return all, nil
}

func observeArtifact(root string, metric metricSpec, relationKind string) RelationObservation {
	relative, err := artifactRelativePath(metric, relationKind)
	path := filepath.Join(root, relative)
	observation := RelationObservation{Relation: relationKind, DeclaredEndpoint: declaredEndpoint(metric, relationKind), ObservedArtifact: filepath.ToSlash(relative), EvidenceKind: evidenceKind(relationKind), RelationStatus: "UNKNOWN", Stage: "ARTIFACT", Step: "read-artifact", Cause: "DIRECT_CAUSE"}
	if err != nil {
		observation.Step, observation.Reason = "resolve-artifact-path", "INVALID_ARTIFACT_PATH"
		return observation
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		observation.Reason = "REQUIRED_RAW_ARTIFACT_MISSING"
		return observation
	}
	observation.ObservedDigest = digestBytes(payload)
	var artifact artifactEnvelope
	if err := json.Unmarshal(payload, &artifact); err != nil {
		observation.Step, observation.Reason = "parse-artifact", "MALFORMED_RAW_ARTIFACT"
		return observation
	}
	if err := validateArtifact(metric, relationKind, observation.DeclaredEndpoint, artifact); err != nil {
		observation.Step, observation.Reason = "validate-artifact", err.Error()
		return observation
	}
	observation.ObservedEndpoint = artifact.Endpoint
	observation.RelationStatus, observation.Step, observation.Reason, observation.Cause = "PASS", "reconstruct-edge", "OBSERVED_ARTIFACT_MATCHES_DECLARATION", "OBSERVED_EVIDENCE"
	return observation
}

func artifactRelativePath(metric metricSpec, relationKind string) (string, error) {
	relative := map[string]string{"PRODUCES": filepath.Join("producer", metric.ID+".json"), "CONSUMES": filepath.Join("consumer", metric.ID+".json"), "OPERATES": filepath.Join("meta-operation", metric.ID+".json"), "EVIDENCED_BY": metric.EvidencePath}[relationKind]
	if relative == "" || filepath.IsAbs(relative) || strings.HasPrefix(filepath.Clean(relative), ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid artifact path")
	}
	return relative, nil
}

func declaredEndpoint(metric metricSpec, relationKind string) string {
	switch relationKind {
	case "PRODUCES":
		return metric.Producer
	case "CONSUMES":
		return metric.Consumer
	case "OPERATES":
		return metric.MetaOperation
	default:
		return metric.EvidencePath
	}
}
