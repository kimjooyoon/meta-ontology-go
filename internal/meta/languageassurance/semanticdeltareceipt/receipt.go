package semanticdeltareceipt

type StructuralDelta struct {
	Status       string `json:"status"`
	AddedNodes   []Node `json:"added_nodes,omitempty"`
	RemovedNodes []Node `json:"removed_nodes,omitempty"`
	AddedFacts   []Fact `json:"added_facts,omitempty"`
	RemovedFacts []Fact `json:"removed_facts,omitempty"`
}

type ClaimChange struct {
	ID     string `json:"id"`
	Before Claim  `json:"before"`
	After  Claim  `json:"after"`
}

type ClaimDelta struct {
	Status    string        `json:"status"`
	Reason    string        `json:"reason,omitempty"`
	Added     []Claim       `json:"added,omitempty"`
	Removed   []Claim       `json:"removed,omitempty"`
	Changed   []ClaimChange `json:"changed,omitempty"`
	Ambiguous []ClaimMatch  `json:"ambiguous,omitempty"`
}

type ClaimTransition struct {
	ClaimID             string `json:"claim_id"`
	Kind                string `json:"kind"`
	FromStatus          string `json:"from_status"`
	ToStatus            string `json:"to_status"`
	FromObject          string `json:"from_object"`
	ToObject            string `json:"to_object"`
	PreservationOf      string `json:"preservation_of,omitempty"`
	PropositionDigest   string `json:"proposition_digest"`
	EvidenceDigest      string `json:"evidence_digest"`
	PreviousEventDigest string `json:"previous_event_digest,omitempty"`
	EventID             string `json:"event_id"`
	TransitionDigest    string `json:"transition_digest"`
	Stage               string `json:"stage"`
	Step                string `json:"step"`
	Reason              string `json:"reason"`
}

type Receipt struct {
	Schema                    string                 `json:"schema"`
	CaseID                    string                 `json:"case_id"`
	SubjectSHA                string                 `json:"subject_sha"`
	Producer                  string                 `json:"producer"`
	Consumer                  string                 `json:"consumer"`
	MetaOperation             string                 `json:"meta_operation"`
	ProofChoice               string                 `json:"proof_choice"`
	Stage                     string                 `json:"stage"`
	Step                      string                 `json:"step"`
	Reason                    string                 `json:"reason"`
	Decision                  string                 `json:"decision"`
	Resolution                string                 `json:"resolution"`
	Classification            string                 `json:"classification"`
	RawDecision               string                 `json:"raw_decision"`
	SemanticDecision          string                 `json:"semantic_decision"`
	MetaSourcePath            string                 `json:"meta_source_path"`
	MetaContractDigest        string                 `json:"meta_contract_digest"`
	DenominatorVersion        string                 `json:"denominator_version"`
	DenominatorCases          int                    `json:"denominator_cases"`
	ModeledSemanticComponents int                    `json:"modeled_semantic_components"`
	TotalSemanticComponents   int                    `json:"total_semantic_components"`
	SemanticCoverageBPS       int                    `json:"semantic_coverage_bps"`
	Before                    Snapshot               `json:"before"`
	After                     Snapshot               `json:"after"`
	TextualDelta              TextualDelta           `json:"textual_delta"`
	StructuralDelta           StructuralDelta        `json:"structural_delta"`
	SemanticComponentDelta    SemanticComponentDelta `json:"semantic_component_delta"`
	SemanticClaimDelta        ClaimDelta             `json:"semantic_claim_delta"`
	ClaimLedger               []Claim                `json:"claim_ledger"`
	ClaimTransitions          []ClaimTransition      `json:"claim_transitions"`
	TransitionCount           int                    `json:"transition_count"`
	TransitionHeadDigest      string                 `json:"transition_head_digest"`
	RepositoryWrites          int                    `json:"repository_writes"`
	ReceiptDigest             string                 `json:"receipt_digest"`
}
