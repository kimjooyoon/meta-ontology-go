package artifactcoverage

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func CanonicalBindings() []ArtifactBinding {
	bindings := []ArtifactBinding{
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
	}
	for _, consumer := range generation.DefaultRegistry() {
		registry := "generation"
		switch consumer.Operation {
		case sourcepolicy.OperationSplitGo, sourcepolicy.OperationSplitGooo:
		case sourcepolicy.OperationExtractFunction:
			registry = "repository-projection"
		default:
			continue
		}
		if consumer.ExecutionRoute == "" || consumer.Executor == "" || consumer.Evaluator == "" {
			continue
		}
		projected := binding(string(consumer.Operation), consumer.Activity,
			ProofChoice(strings.ToLower(string(consumer.ProofChoice))), registry, consumer.Executor,
			"self-improvement-generation", consumer.Output)
		projected.Evaluator = consumer.Evaluator
		bindings = append(bindings, projected)
	}
	return bindings
}

func binding(operation, activity string, proof ProofChoice, registry, executor, artifact, evidence string) ArtifactBinding {
	return ArtifactBinding{
		Operation: operation, Activity: activity, ProofChoice: proof, Registry: registry,
		Executor: executor, Evaluator: executor + ":check",
		ArtifactPattern: artifact + "-{head_sha}", EvidenceKey: evidence,
		ExactHead: true, DigestBound: true, ReplayRequired: true,
	}
}
