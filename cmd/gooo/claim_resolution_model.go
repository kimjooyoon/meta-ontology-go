package main

const (
	claimResolutionSchema      = "gooo/claim-resolution/v1"
	claimResolutionCandidateID = "gooo.primitive.claim-resolution-tuple.v1"
	claimDecisionObserved      = "CLAIM_RESOLUTION_OBSERVED"
	claimDecisionFailed        = "FAIL_CLOSED"
	claimStateClosed           = "CLOSED"
	claimStateUnknown          = "UNKNOWN"
	claimStateRefuted          = "REFUTED"
	claimResolutionNone        = "NONE"
)

type claimResolutionClaim struct {
	State         string  `json:"state"`
	Stage         *string `json:"stage"`
	Step          *string `json:"step"`
	Reason        string  `json:"reason"`
	UnknownClass  *string `json:"unknown_class"`
	NextOperation string  `json:"next_operation"`
}

type claimResolutionSubject struct {
	SourceFile          string `json:"source_file"`
	SourceDigest        string `json:"source_digest"`
	Activity            string `json:"activity"`
	ActivityOccurrences int    `json:"activity_occurrences"`
	ValueProgramDigest  string `json:"value_program_digest"`
	Binding             string `json:"binding"`
}

type claimResolutionContract struct {
	Version              string   `json:"version"`
	BaseFields           []string `json:"base_fields"`
	States               []string `json:"states"`
	UnknownClassRequired bool     `json:"unknown_class_required"`
}

type claimResolutionSummary struct {
	FieldsObserved      int `json:"fields_observed"`
	FieldsTotal         int `json:"fields_total"`
	ResolutionsObserved int `json:"resolutions_observed"`
	ResolutionsTotal    int `json:"resolutions_total"`
	RepositoryWrites    int `json:"repository_writes"`
}

type claimResolutionIndicator struct {
	ID       string `json:"id"`
	Value    int    `json:"value"`
	Total    int    `json:"total"`
	Unit     string `json:"unit"`
	Class    string `json:"class"`
	Activity string `json:"activity"`
}

type claimResolutionAuthority struct {
	Source                 string `json:"source"`
	CoreMutationAuthorized bool   `json:"core_mutation_authorized"`
	RepositoryWrites       int    `json:"repository_writes"`
}

type claimResolutionReport struct {
	Schema     string                     `json:"schema"`
	Candidate  string                     `json:"candidate_id"`
	Decision   string                     `json:"decision"`
	Subject    claimResolutionSubject     `json:"subject"`
	Contract   claimResolutionContract    `json:"contract"`
	Claim      claimResolutionClaim       `json:"claim"`
	Summary    claimResolutionSummary     `json:"summary"`
	Indicators []claimResolutionIndicator `json:"indicators"`
	Authority  claimResolutionAuthority   `json:"authority"`
}
