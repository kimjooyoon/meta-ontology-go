package verify

type operationSpec struct {
	ID                  string `json:"id"`
	Activity            string `json:"activity"`
	ProofChoice         string `json:"proof_choice"`
	InputEntity         string `json:"input_entity"`
	OutputEntity        string `json:"output_entity"`
	Mode                string `json:"mode"`
	Ordinal             int    `json:"ordinal"`
	RepositoryWrites    bool   `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type resolvedBinding struct {
	IndicatorID     string `json:"indicator_id"`
	ProofChoice     string `json:"proof_choice"`
	OperationID     string `json:"operation_id"`
	Activity        string `json:"activity"`
	Mode            string `json:"mode"`
	EvidenceDigest  string `json:"evidence_digest"`
	OperationDigest string `json:"operation_digest"`
}

type programStep struct {
	Index           int      `json:"index"`
	OperationID     string   `json:"operation_id"`
	Activity        string   `json:"activity"`
	Mode            string   `json:"mode"`
	DependsOn       []string `json:"depends_on"`
	InputEntity     string   `json:"input_entity"`
	OutputEntity    string   `json:"output_entity"`
	OperationDigest string   `json:"operation_digest"`
}
