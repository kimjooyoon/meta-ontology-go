package languagereadiness

const (
	ContractSchema = "gooo/self-improving-language-obligations/v1"
	SnapshotSchema = "gooo/language-readiness-snapshot/v1"
)

type Obligation struct {
	ID          string `json:"id"`
	Area        string `json:"area"`
	ProofChoice string `json:"proof_choice"`
	ConceptID   string `json:"concept_id"`
}

type ObligationResult struct {
	Obligation
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type Summary struct {
	Completed       int `json:"completed"`
	Total           int `json:"total"`
	NotSatisfied    int `json:"not_satisfied"`
	Unresolved      int `json:"unresolved"`
	ReadinessBPS    int `json:"readiness_bps"`
	RatioNumerator  int `json:"ratio_numerator"`
	RatioDenominator int `json:"ratio_denominator"`
}

type Indicator struct {
	MetricID     string `json:"metric_id"`
	Class        string `json:"class"`
	ProofChoice  string `json:"proof_choice"`
	Producer     string `json:"producer"`
	Consumer     string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value        int    `json:"value"`
	Target       int    `json:"target"`
	Satisfied    bool   `json:"satisfied"`
}

type Snapshot struct {
	Schema               string             `json:"schema"`
	ContractSchema       string             `json:"contract_schema"`
	Decision             string             `json:"decision"`
	Reason               string             `json:"reason"`
	RegistryDigest       string             `json:"registry_digest"`
	SourceArtifactDigest string             `json:"source_artifact_digest,omitempty"`
	Summary              Summary            `json:"summary"`
	Obligations          []ObligationResult `json:"obligations"`
	Indicators           []Indicator        `json:"indicators"`
	RepositoryWrites     int                `json:"repository_writes"`
	Digest               string             `json:"digest"`
}
