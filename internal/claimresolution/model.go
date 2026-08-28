package claimresolution

const (
	Schema           = "gooo/claim-resolution/v1"
	CandidateID      = "gooo.primitive.claim-resolution-tuple.v1"
	DecisionObserved = "CLAIM_RESOLUTION_OBSERVED"
	DecisionFailed   = "FAIL_CLOSED"
	StateClosed      = "CLOSED"
	StateUnknown     = "UNKNOWN"
	StateRefuted     = "REFUTED"
	None             = "NONE"
)

type Claim struct {
	State         string  `json:"state"`
	Stage         *string `json:"stage"`
	Step          *string `json:"step"`
	Reason        string  `json:"reason"`
	UnknownClass  *string `json:"unknown_class"`
	NextOperation string  `json:"next_operation"`
}

type Subject struct {
	SourceFile         string `json:"source_file"`
	SourceDigest       string `json:"source_digest"`
	Activity           string `json:"activity"`
	ActivityOccurrences int    `json:"activity_occurrences"`
	ValueProgramDigest string `json:"value_program_digest"`
	Binding            string `json:"binding"`
}

type Contract struct {
	Version              string   `json:"version"`
	BaseFields           []string `json:"base_fields"`
	States               []string `json:"states"`
	UnknownClassRequired bool     `json:"unknown_class_required"`
}

type Summary struct {
	FieldsObserved      int `json:"fields_observed"`
	FieldsTotal         int `json:"fields_total"`
	ResolutionsObserved int `json:"resolutions_observed"`
	ResolutionsTotal    int `json:"resolutions_total"`
	RepositoryWrites    int `json:"repository_writes"`
}

type Indicator struct {
	ID       string `json:"id"`
	Value    int    `json:"value"`
	Total    int    `json:"total"`
	Unit     string `json:"unit"`
	Class    string `json:"class"`
	Activity string `json:"activity"`
}

type Authority struct {
	Source                 string `json:"source"`
	CoreMutationAuthorized bool   `json:"core_mutation_authorized"`
	RepositoryWrites       int    `json:"repository_writes"`
}

type Report struct {
	Schema     string      `json:"schema"`
	Candidate  string      `json:"candidate_id"`
	Decision   string      `json:"decision"`
	Subject    Subject     `json:"subject"`
	Contract   Contract    `json:"contract"`
	Claim      Claim       `json:"claim"`
	Summary    Summary     `json:"summary"`
	Indicators []Indicator `json:"indicators"`
	Authority  Authority   `json:"authority"`
}
