package artifactcoverage

func CanonicalBindings() []ArtifactBinding {
	return []ArtifactBinding{
		binding("bind-indicator-meta-program", "BindIndicatorMetaProgram", ProofCoherence,
			"meta-binding", "bootstrap/meta-binding-witness", "meta-binding-coverage", "meta-binding.report"),
		binding("measure-integration-progress", "MeasureIntegrationProgress", ProofFoundation,
			"integration-progress", "cmd/integration-progress-witness", "integration-progress-evidence", "integration-progress.report"),
		binding("measure-language-utility", "MeasureLanguageUtility", ProofFoundation,
			"language-utility", "cmd/language-utility-witness", "language-utility-evidence", "language-utility.report"),
		binding("partition-directory", "PartitionDirectory", ProofFoundation,
			"source-policy", "cmd/directory-partition-witness", "directory-partition", "directory-partition.report"),
		binding("separate-directory-kinds", "SeparateDirectoryKinds", ProofFoundation,
			"source-policy", "cmd/directory-kind-witness", "directory-kind-separation", "directory-kind.report"),
		binding("split-go-declarations", "SplitGoDeclarations", ProofFoundation,
			"generation", "scripts/source-splitter", "self-improvement-generation", "operation.split-go-declarations"),
		binding("split-gooo-sections", "SplitGoooSections", ProofFoundation,
			"generation", "bootstrap/source-repacker", "self-improvement-generation", "operation.split-gooo-sections"),
	}
}

func binding(operation, activity string, proof ProofChoice, registry, executor, artifact, evidence string) ArtifactBinding {
	return ArtifactBinding{
		Operation: operation, Activity: activity, ProofChoice: proof, Registry: registry,
		Executor: executor, Evaluator: executor + ":check",
		ArtifactPattern: artifact + "-{head_sha}", EvidenceKey: evidence,
		ExactHead: true, DigestBound: true, ReplayRequired: true,
	}
}
