package artifactemit

const symbolicValueContractSchema = "gooo/symbolic-invocation-value-contract/v1"

type SymbolicValueContract struct {
	Schema               string                           `json:"schema"`
	SubjectSHA           string                           `json:"subject_sha"`
	MetricID             string                           `json:"metric_id"`
	Decision             string                           `json:"decision"`
	Resolution           string                           `json:"resolution"`
	Reason               string                           `json:"reason"`
	SourceArtifactDigest string                           `json:"source_artifact_digest"`
	Rules                []SymbolicValueContractRule      `json:"rules"`
	Default              SymbolicValueContractDefault     `json:"default"`
	Coordinates          SymbolicValueContractCoordinates `json:"coordinates"`
	Classes              []SymbolicValueContractClass     `json:"classes"`
	Indicators           []SymbolicValueContractIndicator `json:"indicators"`
	Views                []SymbolicValueContractView      `json:"views"`
	Proofs               []SymbolicValueContractProof     `json:"proofs"`
	Effects              SymbolicValueContractEffects     `json:"effects"`
	PromotionCreditBPS   int                              `json:"promotion_credit_bps"`
	NotClaimed           []string                         `json:"not_claimed"`
	Digest               string                           `json:"digest,omitempty"`
}

type SymbolicValueContractRule struct {
	ID            string                         `json:"id"`
	Match         SymbolicValueContractRuleMatch `json:"match"`
	Decision      string                         `json:"decision"`
	Resolution    string                         `json:"resolution"`
	Reason        string                         `json:"reason"`
	ProofChoice   string                         `json:"proof_choice"`
	MetaOperation string                         `json:"meta_operation"`
}

type SymbolicValueContractRuleMatch struct {
	Activity string `json:"activity"`
	Inputs   string `json:"inputs"`
}

type SymbolicValueContractDefault struct {
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type SymbolicValueContractCoordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type SymbolicValueContractClass struct {
	Class     string `json:"class"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}
