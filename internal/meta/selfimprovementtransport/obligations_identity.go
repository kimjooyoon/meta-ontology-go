package selfimprovementtransport

func sourceRepositoryObligation(input evaluationInput) Obligation {
	expectedURI := "https://github.com/" + input.expectedRepository
	return knownObligation("source-repository-commit", "FOUNDATION", "SOURCE", "validate-repository-commit",
		input.sourceErr == nil && input.producerErr == nil && input.source.Schema == ObservationSchema &&
			validSHA(input.source.SubjectSHA) && input.producer.SubjectSHA == input.source.SubjectSHA &&
			input.producer.RepositoryURI == expectedURI,
		"SOURCE_REPOSITORY_COMMIT_VERIFIED", "SOURCE_REPOSITORY_COMMIT_MISMATCH",
		struct{ Repository, Subject string }{input.producer.RepositoryURI, input.source.SubjectSHA})
}

func checkoutBindingObligation(input evaluationInput) Obligation {
	return knownObligation("producer-checkout-binding", "COHERENCE", "PRODUCE", "verify-checkout-head",
		input.producerErr == nil && validSHA(input.producer.CheckoutSHA) &&
			input.producer.CheckoutSHA == input.source.SubjectSHA,
		"PRODUCER_CHECKOUT_BOUND", "PRODUCER_CHECKOUT_MISMATCH",
		struct{ Checkout, Subject string }{input.producer.CheckoutSHA, input.source.SubjectSHA})
}

func producerIdentityObligation(input evaluationInput) Obligation {
	passed := input.producerErr == nil && input.metadataErr == nil && input.expectedRunID > 0 &&
		input.producer.RunID == input.expectedRunID && input.metadata.ProducerRunID == input.expectedRunID &&
		input.producer.RunAttempt == input.metadata.ProducerRunAttempt && input.producer.WorkflowRef != "" &&
		validSHA(input.producer.WorkflowSHA) && input.producer.Job != "" &&
		input.metadata.WorkflowPath != "" && validSHA(input.metadata.OrchestrationHeadSHA)
	evidence := struct {
		RunID    int64
		Attempt  int
		Workflow string
	}{input.metadata.ProducerRunID, input.metadata.ProducerRunAttempt, input.metadata.WorkflowPath}
	return knownObligation("producer-run-identity", "FOUNDATION", "PRODUCE", "bind-run-identity",
		passed, "PRODUCER_RUN_IDENTITY_BOUND", "PRODUCER_RUN_IDENTITY_MISMATCH", evidence)
}
