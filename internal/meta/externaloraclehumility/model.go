package externaloraclehumility

type Claim struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	State string `json:"state"`
}

type SourceContract struct {
	Path      string  `json:"path"`
	SHA256    string  `json:"sha256"`
	Authority string  `json:"authority"`
	Claims    []Claim `json:"claims"`
}

type ReferenceContract struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	URL            string `json:"url"`
	Revision       string `json:"revision"`
	Locator        string `json:"locator"`
	DocumentSHA256 string `json:"document_sha256"`
	ClaimID        string `json:"claim_id"`
	Relation       string `json:"relation"`
	Authority      string `json:"authority"`
	Primary        bool   `json:"primary"`
}

type DenominatorContract struct {
	Version         string `json:"version"`
	Total           int    `json:"total"`
	BasisPointsGoal int    `json:"basis_points_required"`
}

type CaseContract struct {
	ID                 string `json:"id"`
	Input              string `json:"input"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedAuthority  string `json:"expected_authority"`
	ExpectedEffect     string `json:"expected_effect"`
}

type Contract struct {
	Schema           string              `json:"schema"`
	Version          int                 `json:"version"`
	Source           SourceContract      `json:"source"`
	FixedDenominator DenominatorContract `json:"fixed_denominator"`
	References       []ReferenceContract `json:"references"`
	Cases            []CaseContract      `json:"cases"`
}

type ReferenceEvidenceSet struct {
	Schema     string              `json:"schema"`
	CapturedOn string              `json:"captured_on"`
	References []ReferenceEvidence `json:"references"`
}

type ReferenceEvidence struct {
	ID             string `json:"id"`
	Available      bool   `json:"available"`
	DocumentSHA256 string `json:"document_sha256"`
	Proposition    string `json:"proposition"`
	EvidenceRole   string `json:"evidence_role"`
	Relation       string `json:"relation"`
	Authority      string `json:"authority"`
	Agreement      bool   `json:"agreement"`
}

type SourceReceipt struct {
	Schema        string  `json:"schema"`
	SubjectSHA    string  `json:"subject_sha"`
	SourcePath    string  `json:"source_path"`
	SourceSHA256  string  `json:"source_sha256"`
	Producer      string  `json:"producer"`
	Consumer      string  `json:"consumer"`
	MetaOperation string  `json:"meta_operation"`
	ProofChoice   string  `json:"proof_choice"`
	Stage         string  `json:"stage"`
	Step          string  `json:"step"`
	Reason        string  `json:"reason"`
	Claims        []Claim `json:"claims"`
}

type ClaimTransition struct {
	ClaimID       string `json:"claim_id"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Before        string `json:"before"`
	After         string `json:"after"`
	Reason        string `json:"reason"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Persisted     bool   `json:"persisted"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
}

type Criterion struct {
	ID            string
	Class         string
	ProofChoice   string
	Producer      string
	Consumer      string
	MetaOperation string
	Stage         string
	Step          string
	Unit          string
	Relation      string
	Target        int
}

type Input struct {
	Contract Contract
	Evidence ReferenceEvidenceSet
	Receipt  SourceReceipt
	Source   []byte
	Subject  string
}

type Report struct {
	Schema             string            `json:"schema"`
	SubjectSHA         string            `json:"subject_sha"`
	Decision           string            `json:"decision"`
	Resolution         string            `json:"resolution"`
	Reason             string            `json:"reason"`
	ReferenceAgreement string            `json:"reference_agreement"`
	SemanticAuthority  string            `json:"semantic_authority"`
	AuthorityGrant     string            `json:"authority_grant"`
	EnforcementEffect  string            `json:"enforcement_effect"`
	Completed          int               `json:"completed"`
	Total              int               `json:"total"`
	BasisPoints        int               `json:"basis_points"`
	UnknownIndicators  int               `json:"unknown_indicators"`
	OfficialMutations  int               `json:"official_mutations"`
	RepositoryWrites   int               `json:"repository_writes"`
	PromotionCount     int               `json:"promotion_count"`
	Producer           string            `json:"producer"`
	Consumer           string            `json:"consumer"`
	MetaOperation      string            `json:"meta_operation"`
	ProofChoice        string            `json:"proof_choice"`
	Stage              string            `json:"stage"`
	Step               string            `json:"step"`
	Indicators         []Indicator       `json:"indicators"`
	Transitions        []ClaimTransition `json:"claim_transitions"`
	Receipt            SourceReceipt     `json:"source_receipt"`
	ReportDigest       string            `json:"report_digest"`
}

type SuiteCase struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Authority          string `json:"authority"`
	Effect             string `json:"effect"`
	Passed             bool   `json:"passed"`
}

type Suite struct {
	Schema                 string      `json:"schema"`
	SubjectSHA             string      `json:"subject_sha"`
	Decision               string      `json:"decision"`
	Resolution             string      `json:"resolution"`
	Reason                 string      `json:"reason"`
	CaseDenominatorVersion string      `json:"case_denominator_version"`
	CasesTotal             int         `json:"cases_total"`
	CasesSatisfied         int         `json:"cases_satisfied"`
	CoverageBPS            int         `json:"coverage_bps"`
	OfficialMutations      int         `json:"official_mutations"`
	RepositoryWrites       int         `json:"repository_writes"`
	PromotionCount         int         `json:"promotion_count"`
	Cases                  []SuiteCase `json:"cases"`
	SuiteDigest            string      `json:"suite_digest"`
}
