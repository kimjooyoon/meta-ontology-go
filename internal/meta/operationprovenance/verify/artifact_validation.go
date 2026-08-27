package verify

import (
	"path/filepath"
	"strings"
)

func relativePath(metric cMetric, relation string) string {
	if relation == "EVIDENCED_BY" {
		return filepath.Clean(metric.evidence)
	}
	return filepath.Join(map[string]string{"PRODUCES": "producer", "CONSUMES": "consumer", "OPERATES": "meta-operation"}[relation], metric.id+".json")
}

func endpoint(metric cMetric, relation string) string {
	switch relation {
	case "PRODUCES":
		return metric.producer
	case "CONSUMES":
		return metric.consumer
	case "OPERATES":
		return metric.operation
	default:
		return metric.evidence
	}
}

func validArtifact(metric cMetric, relation, declared string, artifact artifactEnvelope) bool {
	if artifact.MetricID != metric.id || artifact.Endpoint != declared {
		return false
	}
	switch relation {
	case "PRODUCES":
		return artifact.Kind == "producer_receipt" && artifact.Output == "metric:"+metric.id && artifact.Status == "PASS"
	case "CONSUMES":
		return artifact.Kind == "consumer_reconstruction_receipt" && artifact.Reads == "metric:"+metric.id && artifact.Source != "" && artifact.Status == "PASS"
	case "OPERATES":
		return artifact.Kind == "executed_meta_operation_receipt" && artifact.Executed && artifact.Input == "metric:"+metric.id && artifact.Output == "metric:"+metric.id
	case "EVIDENCED_BY":
		return artifact.Kind == "evidence_artifact" && artifact.Path == declared && artifact.Payload != ""
	}
	return false
}

func safeRelative(value string) bool {
	return value != "" && !filepath.IsAbs(value) && !strings.HasPrefix(filepath.Clean(value), ".."+string(filepath.Separator))
}

func evidenceKind(relation string) string {
	return map[string]string{"PRODUCES": "producer_receipt", "CONSUMES": "consumer_reconstruction_receipt", "OPERATES": "executed_meta_operation_receipt", "EVIDENCED_BY": "evidence_artifact"}[relation]
}
