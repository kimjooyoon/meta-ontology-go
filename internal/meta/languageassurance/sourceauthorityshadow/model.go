package sourceauthorityshadow

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityeval"

const (
	ReceiptSchema = "gooo/source-backed-authority-shadow-receipt/v1"
	Mode          = "SHADOW"
)

type Receipt struct {
	Schema             string                     `json:"schema"`
	Mode               string                     `json:"mode"`
	ExpectedSHA        string                     `json:"expected_sha"`
	SubjectSHA         string                     `json:"subject_sha"`
	InputDigest        string                     `json:"input_digest"`
	Observation        string                     `json:"observation"`
	Resolution         string                     `json:"resolution"`
	Enforcement        string                     `json:"enforcement"`
	Reason             string                     `json:"reason"`
	GateEffect         string                     `json:"gate_effect"`
	PromotionCreditBPS int                        `json:"promotion_credit_bps"`
	RepositoryWrites   int                        `json:"repository_writes"`
	Evaluation         sourceauthorityeval.Report `json:"evaluation"`
	Indicators         []Indicator                `json:"indicators"`
	ReceiptDigest      string                     `json:"receipt_digest"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Satisfied     bool   `json:"satisfied"`
}
