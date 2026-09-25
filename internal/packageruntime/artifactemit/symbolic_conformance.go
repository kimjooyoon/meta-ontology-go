package artifactemit

import "encoding/json"

type artifactJSONFields Artifact

type artifactEnvelope struct {
	artifactJSONFields
	Conformance *SymbolicInvocationConformance `json:"conformance,omitempty"`
}

func (artifact Artifact) MarshalJSON() ([]byte, error) {
	return json.Marshal(artifactEnvelope{
		artifactJSONFields: artifactJSONFields(artifact),
		Conformance:        projectSymbolicConformance(artifact),
	})
}

func projectSymbolicConformance(artifact Artifact) *SymbolicInvocationConformance {
	if artifact.Decision != "PASS" || artifact.Kind != SymbolicInvocationSchemaKind {
		return nil
	}
	inputs := make([]string, len(artifact.Operation.Inputs))
	for index, input := range artifact.Operation.Inputs {
		inputs[index] = input.ID
	}
	vectors := []ConformanceVector{
		{
			ID: "accept-exact", Expected: "ACCEPT", ProofChoice: "FOUNDATION",
			MetaOperation: "project-exact-symbolic-invocation",
			Instance:      ConformanceInstance{Activity: artifact.Operation.Activity, Inputs: inputs},
		},
		{
			ID: "reject-missing-activity", Expected: "REJECT", ProofChoice: "REGRESSION",
			MetaOperation: "remove-required-activity",
			Instance:      ConformanceInstance{Inputs: inputs},
		},
	}
	return &SymbolicInvocationConformance{
		Schema: SymbolicInvocationConformanceSchema, Decision: "PASS",
		Resolution: "STRUCTURAL_ONLY", Reason: "SYMBOLIC_CONFORMANCE_VECTORS_PROJECTED",
		GeneratedVectors: len(vectors), EmbeddedHandwrittenVectors: 0, Vectors: vectors,
		Effects:    artifact.Effects,
		NotClaimed: []string{"external validation", "external fixture replacement", "value-level execution", "domain correctness", "production readiness"},
	}
}
