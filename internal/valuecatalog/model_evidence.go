package valuecatalog

type Summary struct {
	BaselineCasesPassed         int        `json:"baseline_cases_passed"`
	BaselineCasesTotal          int        `json:"baseline_cases_total"`
	ExtensionCasesPassed        int        `json:"extension_cases_passed"`
	ExtensionCasesTotal         int        `json:"extension_cases_total"`
	UnknownCounterexamplePassed bool       `json:"unknown_counterexample_passed"`
	RepositoryWrites            int        `json:"repository_writes"`
	CoreFingerprintSensitive    Coordinate `json:"core_fingerprint_sensitive"`
}

type Authority struct {
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
	PromotionAuthorized          bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized  bool `json:"automatic_adoption_authorized"`
}

type Indicator struct {
	ID            string `json:"id"`
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type OperationSpecMetrics struct {
	MetricID                 string `json:"metric_id"`
	FixedAxisTotal           int    `json:"fixed_axis_total"`
	VerifiedTotal            int    `json:"verified_total"`
	CoverageBasisPoints      int    `json:"coverage_basis_points"`
	UnknownPathCount         int    `json:"unknown_path_count"`
	OpenClaims               int    `json:"open_claims"`
	DischargedClaims         int    `json:"discharged_claims"`
	TransitionEventTotal     int    `json:"transition_event_total"`
	RegistrationEventTotal   int    `json:"registration_event_total"`
	EvidenceAcceptedTotal    int    `json:"evidence_accepted_event_total"`
	EvidenceUnavailableTotal int    `json:"evidence_unavailable_event_total"`
}

type Claim struct {
	ClaimID        string `json:"claim_id"`
	Stage          string `json:"stage"`
	Statement      string `json:"statement"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type ProcessCoordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}
