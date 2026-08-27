package observereffect

const (
	LedgerSchema     = "gooo/observer-effect-ledger/v1"
	ReceiptSchema    = "gooo/observer-effect-receipt/v1"
	JudgmentSchema   = "gooo/observer-effect-judgment/v1"
	ExperimentName   = "meta-observer-self-effect"
	FixedDenominator = 12
)

type Source struct {
	Path       string `json:"path"`
	Digest     string `json:"digest"`
	GoooSource bool   `json:"gooo_source"`
}

type SnapshotDelta struct {
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	Changed      bool   `json:"changed"`
}

type Observation struct {
	RepositoryStorage SnapshotDelta `json:"repository_storage"`
	Environment       SnapshotDelta `json:"environment"`
	LogicalTime       SnapshotDelta `json:"logical_time"`
}

type Effect struct {
	Domain            string `json:"domain"`
	ObservedChanged   bool   `json:"observed_changed"`
	MutationAttempted bool   `json:"mutation_attempted"`
	BeforeDigest      string `json:"before_digest"`
	AfterDigest       string `json:"after_digest"`
	WriteCount        int    `json:"write_count"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	Status            string `json:"status"`
	Reason            string `json:"reason"`
}

type Unknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	ClaimID                string `json:"claim_id"`
	Persistent             bool   `json:"persistent"`
	Sequence               int    `json:"sequence"`
	PreviousState          string `json:"previous_state"`
	CurrentState           string `json:"current_state"`
	Transition             string `json:"transition"`
	PreviousEvidenceDigest string `json:"previous_evidence_digest"`
	EvidenceDigest         string `json:"evidence_digest"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Expected      string `json:"expected"`
	Actual        string `json:"actual"`
	Status        string `json:"status"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type Metrics struct {
	FixedDenominator     int `json:"fixed_denominator"`
	Satisfied            int `json:"satisfied"`
	CoverageBasisPoints  int `json:"coverage_basis_points"`
	ObservationSatisfied int `json:"observation_satisfied"`
	ObservationTotal     int `json:"observation_total"`
	EffectSatisfied      int `json:"effect_satisfied"`
	EffectTotal          int `json:"effect_total"`
	GuardrailSatisfied   int `json:"guardrail_satisfied"`
	GuardrailTotal       int `json:"guardrail_total"`
}

type Authority struct {
	RepositoryWrites    int  `json:"repository_writes"`
	OutputWrites        int  `json:"output_writes"`
	MutationAuthority   bool `json:"mutation_authority"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type Receipt struct {
	Schema            string          `json:"schema"`
	Kind              string          `json:"kind"`
	Producer          string          `json:"producer"`
	Consumer          string          `json:"consumer"`
	MetaOperation     string          `json:"meta_operation"`
	ProofChoice       string          `json:"proof_choice"`
	Subject           string          `json:"subject"`
	SubjectDigest     string          `json:"subject_digest"`
	Decision          string          `json:"decision"`
	Resolution        string          `json:"resolution"`
	RepositoryWrites  int             `json:"repository_writes"`
	MutationAuthority bool            `json:"mutation_authority"`
	Unknown           Unknown         `json:"unknown"`
	Coordinate        Unknown         `json:"coordinate"`
	Reason            string          `json:"reason"`
	ClaimTransition   ClaimTransition `json:"claim_transition"`
	EvidenceDigest    string          `json:"evidence_digest"`
	Digest            string          `json:"digest"`
}

type Report struct {
	Schema              string          `json:"schema"`
	Experiment          string          `json:"experiment"`
	Source              Source          `json:"source"`
	Observation         Observation     `json:"observation"`
	Effects             []Effect        `json:"effects"`
	ReceiptDigests      []string        `json:"receipt_digests"`
	Unknown             Unknown         `json:"unknown"`
	Coordinate          Unknown         `json:"coordinate"`
	Reason              string          `json:"reason"`
	ClaimTransition     ClaimTransition `json:"claim_transition"`
	Metrics             Metrics         `json:"metrics"`
	Authority           Authority       `json:"authority"`
	RepositoryWrites    int             `json:"repository_writes"`
	MutationAuthority   bool            `json:"mutation_authority"`
	PromotionAuthorized bool            `json:"promotion_authorized"`
	Decision            string          `json:"decision"`
	Resolution          string          `json:"resolution"`
	EvidenceDigest      string          `json:"evidence_digest"`
	Digest              string          `json:"digest"`
	Indicators          []Indicator     `json:"indicators"`
}
