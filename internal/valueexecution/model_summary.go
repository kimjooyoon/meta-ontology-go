package valueexecution

type Summary struct {
	ValueCasesPassed       int        `json:"value_cases_passed"`
	ValueCasesTotal        int        `json:"value_cases_total"`
	CounterexamplesPassed  int        `json:"counterexamples_passed"`
	CounterexamplesTotal   int        `json:"counterexamples_total"`
	ValueOutputsObserved   int        `json:"value_outputs_observed"`
	DeterministicReplays   int        `json:"deterministic_replays"`
	RepositoryWrites       int        `json:"repository_writes"`
	CoreIRProgramPreserved           Coordinate `json:"core_ir_program_preserved"`
	CoreIRFingerprintSensitive       Coordinate `json:"core_ir_fingerprint_sensitive"`
	CoreIRUnknownAttributeFailClosed Coordinate `json:"core_ir_unknown_attribute_fail_closed"`
}

type Authority struct {
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
	PromotionAuthorized          bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized  bool `json:"automatic_adoption_authorized"`
}
