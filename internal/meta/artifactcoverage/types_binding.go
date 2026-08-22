package artifactcoverage

type ArtifactBinding struct {
	Operation       string      `json:"operation"`
	Activity        string      `json:"activity"`
	ProofChoice     ProofChoice `json:"proof_choice"`
	Registry        string      `json:"registry"`
	Executor        string      `json:"executor"`
	Evaluator       string      `json:"evaluator"`
	ArtifactPattern string      `json:"artifact_pattern"`
	EvidenceKey     string      `json:"evidence_key"`
	ExactHead       bool        `json:"exact_head"`
	DigestBound     bool        `json:"digest_bound"`
	ReplayRequired  bool        `json:"replay_required"`
}
