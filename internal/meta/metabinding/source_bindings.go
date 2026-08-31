package metabinding

func sourceBindings() []Binding {
	return []Binding{
		{Operation: "bind-indicator-meta-program", Activity: "BindIndicatorMetaProgram", ProofChoice: "coherence", Registry: "meta-binding"},
		{Operation: "exempt-project-root-readme", Activity: "BindRootREADMEExemption", ProofChoice: "foundation", Registry: "source-policy"},
		{Operation: "extract-function", Activity: "ExtractFunction", ProofChoice: "foundation", Registry: "repository-projection", Executor: "bootstrap/function-extractor", Evaluator: ".github/workflows/repository-projection.yml"},
		{Operation: "inspect-wrapper", Activity: "InspectWrapper", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "measure-integration-progress", Activity: "MeasureIntegrationProgress", ProofChoice: "foundation", Registry: "integration-progress", Executor: "cmd/integration-progress-witness", Evaluator: ".github/workflows/integration-progress-evidence.yml"},
		{Operation: "measure-language-utility", Activity: "MeasureLanguageUtility", ProofChoice: "foundation", Registry: "language-utility", Executor: "cmd/language-utility-witness", Evaluator: ".github/workflows/language-utility-evidence.yml"},
		{Operation: "observe", Activity: "ObserveMetric", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "partition-directory", Activity: "PartitionDirectory", ProofChoice: "foundation", Registry: "source-policy", Executor: "cmd/directory-partition-witness", Evaluator: "cmd/directory-partition-witness:check"},
		{Operation: "preserve-workflow-discovery", Activity: "PreserveWorkflowDiscovery", ProofChoice: "foundation", Registry: "source-policy", Executor: "scripts/line-metrics", Evaluator: ".github/workflows/repository-projection.yml"},
		{Operation: "separate-directory-kinds", Activity: "SeparateDirectoryKinds", ProofChoice: "foundation", Registry: "source-policy", Executor: "cmd/directory-kind-witness", Evaluator: "cmd/directory-kind-witness:check"},
	}
}
