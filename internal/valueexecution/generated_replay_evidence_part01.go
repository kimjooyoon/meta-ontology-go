package valueexecution

// GeneratedReplayEvidencePart01 binds a generated runtime-plan replay to the
// source, evaluator, artifact, and reverse execution observation that produced
// it. The fields are evidence only and grant no execution or adoption
// authority.
type GeneratedReplayEvidencePart01 struct {
	SourceDigest             string `json:"source_digest"`
	SemanticDigest           string `json:"semantic_digest"`
	TypedPlanDigest          string `json:"typed_plan_digest"`
	RuntimePlanDigest        string `json:"runtime_plan_digest"`
	GeneratedArtifactDigest  string `json:"generated_artifact_digest"`
	ToolchainDigest          string `json:"toolchain_digest"`
	EvaluatorDigest          string `json:"evaluator_digest"`
	ReverseObservationDigest string `json:"reverse_observation_digest"`
}

// StageDigests returns the stable source-to-replay evidence order. Missing or
// invalid values are interpreted by the consuming provenance boundary.
func (evidence GeneratedReplayEvidencePart01) StageDigests() []string {
	return evidence.RawStageDigests()
}

// RawStageDigests returns a defensive copy in the stable evidence order.
func (evidence GeneratedReplayEvidencePart01) RawStageDigests() []string {
	return []string{
		evidence.SourceDigest,
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
		evidence.RuntimePlanDigest,
		evidence.GeneratedArtifactDigest,
		evidence.ToolchainDigest,
		evidence.EvaluatorDigest,
		evidence.ReverseObservationDigest,
	}
}
