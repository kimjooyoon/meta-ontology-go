package foundationseed

type Source struct {
	CurrentHeadSHA          string `json:"current_head_sha"`
	ImmediatePredecessorSHA string `json:"immediate_predecessor_sha"`
	ResolutionDigest        string `json:"resolution_digest"`
	ResolutionValid         bool   `json:"resolution_valid"`
	HeadBound               bool   `json:"head_bound"`
	ObservedAttempts        int    `json:"observed_attempts"`
	MissingAttempts         int    `json:"missing_attempts"`
	SearchLimit             int    `json:"search_limit"`
	SearchComplete          bool   `json:"search_complete"`
	MissingComplete         bool   `json:"missing_complete"`
	Contiguous              bool   `json:"contiguous"`
	SelectedAncestors       int    `json:"selected_ancestors"`
	ValidCandidates         int    `json:"valid_candidates"`
	AmbiguousCandidates     int    `json:"ambiguous_candidates"`
	RepositoryWrites        int    `json:"repository_writes"`
	ReadinessDeltaClaims    *int   `json:"readiness_delta_claims"`
	ExactExhaustion         bool   `json:"exact_exhaustion"`
	AuthorityDenied         bool   `json:"authority_denied"`
}

type Indicator struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Choice string `json:"choice"`
	Value  int    `json:"value"`
	Target int    `json:"target"`
	Passed bool   `json:"passed"`
}
