package operationprovenance

import "fmt"

func validateArtifact(metric metricSpec, relationKind, declared string, artifact artifactEnvelope) error {
	if artifact.MetricID != metric.ID || artifact.Endpoint != declared {
		return fmt.Errorf("ARTIFACT_ENDPOINT_MISMATCH")
	}
	switch relationKind {
	case "PRODUCES":
		if artifact.Kind != "producer_receipt" || artifact.Output != "metric:"+metric.ID || artifact.Status != "PASS" {
			return fmt.Errorf("PRODUCER_OUTPUT_NOT_OBSERVED")
		}
	case "CONSUMES":
		if artifact.Kind != "consumer_reconstruction_receipt" || artifact.Reads != "metric:"+metric.ID || artifact.Source == "" || artifact.Status != "PASS" {
			return fmt.Errorf("CONSUMER_READ_NOT_OBSERVED")
		}
	case "OPERATES":
		if artifact.Kind != "executed_meta_operation_receipt" || !artifact.Executed || artifact.Input != "metric:"+metric.ID || artifact.Output != "metric:"+metric.ID {
			return fmt.Errorf("META_OPERATION_EXECUTION_NOT_OBSERVED")
		}
	case "EVIDENCED_BY":
		if artifact.Kind != "evidence_artifact" || artifact.Path != declared || artifact.Payload == "" {
			return fmt.Errorf("EVIDENCE_ARTIFACT_NOT_OBSERVED")
		}
	default:
		return fmt.Errorf("UNKNOWN_RELATION_KIND")
	}
	return nil
}
