package replay

type LawObservation struct {
	AnchorPath                    string `json:"anchor_path"`
	PresentationChanged           bool   `json:"presentation_changed"`
	PresentationInvariant         bool   `json:"presentation_invariant"`
	CandidateRecorded             bool   `json:"candidate_recorded"`
	CandidateNonAuthoritative     bool   `json:"candidate_non_authoritative"`
	DeterministicRecorded         bool   `json:"deterministic_recorded"`
	DeterministicAuthoritative    bool   `json:"deterministic_authoritative"`
	StructureSemanticHash         string `json:"structure_semantic_hash"`
	PresentationSemanticHash      string `json:"presentation_semantic_hash"`
	CandidateSemanticHash         string `json:"candidate_semantic_hash"`
	DeterministicSemanticHash     string `json:"deterministic_semantic_hash"`
	CandidateCanonicalChanged     bool   `json:"candidate_canonical_changed"`
	DeterministicCanonicalChanged bool   `json:"deterministic_canonical_changed"`
}
