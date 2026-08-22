package improvement

type Indicator struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Before int64  `json:"before"`
	After  int64  `json:"after"`
	Delta  int64  `json:"delta"`
	Unit   string `json:"unit"`
}

type Proof struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Passed bool   `json:"passed"`
}

type Transition struct {
	Schema             string      `json:"schema"`
	Decision           Decision    `json:"decision"`
	ReasonCode         string      `json:"reason_code"`
	Comparable         bool        `json:"comparable"`
	ContractSchema     string      `json:"contract_schema"`
	RegistryDigest     string      `json:"registry_digest"`
	BeforeCompleted    int64       `json:"before_completed"`
	AfterCompleted     int64       `json:"after_completed"`
	Total              int64       `json:"total"`
	CompletedDelta     int64       `json:"completed_delta"`
	BeforeBasisPoints  int64       `json:"before_basis_points"`
	AfterBasisPoints   int64       `json:"after_basis_points"`
	BasisPointsDelta   int64       `json:"basis_points_delta"`
	Gains              int64       `json:"gains"`
	Regressions        int64       `json:"regressions"`
	BeforeUnresolved   int64       `json:"before_unresolved"`
	AfterUnresolved    int64       `json:"after_unresolved"`
	Indicators         []Indicator `json:"indicators"`
	Proofs             []Proof     `json:"proofs"`
	Digest             string      `json:"digest"`
}

type inspection struct {
	statuses   map[string]EvidenceStatus
	unresolved int64
	reason     string
}
