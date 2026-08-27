package operationprovenance

func evidenceKind(relationKind string) string {
	return map[string]string{"PRODUCES": "producer_receipt", "CONSUMES": "consumer_reconstruction_receipt", "OPERATES": "executed_meta_operation_receipt", "EVIDENCED_BY": "evidence_artifact"}[relationKind]
}
