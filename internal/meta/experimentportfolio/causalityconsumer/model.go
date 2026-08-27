package causalityconsumer

// The consumer owns the wire-model copy used at the producer/consumer
// boundary. It intentionally imports no producer package.
type Contract struct {
	Schema                               string              `json:"schema"`
	ID                                   string              `json:"contract_id"`
	ReceiptSchema                        string              `json:"receipt_schema"`
	ReportSchema                         string              `json:"report_schema"`
	Version                              int                 `json:"version"`
	PredecessorContractID                string              `json:"predecessor_contract_id"`
	PredecessorContractPath              string              `json:"predecessor_contract_path"`
	PredecessorContractDigest            string              `json:"predecessor_contract_digest"`
	UpgradeReason                        string              `json:"upgrade_reason"`
	CausalityManifestSchema              string              `json:"causality_manifest_schema"`
	CausalityManifestPath                string              `json:"causality_manifest_path"`
	CausalityCoordinateID                string              `json:"causality_coordinate_id"`
	CausalityDenominator                 int                 `json:"causality_denominator"`
	CausalityTransitionDenominator       int                 `json:"causality_transition_denominator"`
	CausalityTransitionDenominatorReason string              `json:"causality_transition_denominator_reason"`
	Candidates                           []CandidateContract `json:"candidates"`
	CoordinateIDs                        []string            `json:"coordinate_ids"`
	CoordinateDenominators               map[string]int      `json:"coordinate_denominators"`
	CounterexampleSlots                  int                 `json:"counterexample_slots"`
	UnknownLocationSlots                 int                 `json:"unknown_location_slots"`
	NotClaimed                           []string            `json:"not_claimed"`
}

type CandidateContract struct {
	ID            string `json:"id"`
	SourcePath    string `json:"source_path"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Coordinate struct {
	ID            string `json:"id"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Status        string `json:"status"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type Counterexample struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Claim    string `json:"claim"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}

type UnknownLocation struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ExtensionEvidence struct {
	ID       string `json:"id"`
	Claim    string `json:"claim"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}

type ClaimTransition struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Receipt struct {
	Schema            string              `json:"schema"`
	SubjectSHA        string              `json:"subject_sha"`
	CandidateID       string              `json:"candidate_id"`
	SourcePath        string              `json:"source_path"`
	SourceDigest      string              `json:"source_digest"`
	Producer          string              `json:"producer"`
	Consumer          string              `json:"consumer"`
	MetaOperation     string              `json:"meta_operation"`
	ProofChoice       string              `json:"proof_choice"`
	SemanticValue     string              `json:"semantic_value"`
	Decision          string              `json:"decision"`
	ClaimTransitions  []ClaimTransition   `json:"claim_transitions"`
	CoordinateVector  []Coordinate        `json:"coordinate_vector"`
	Counterexamples   []Counterexample    `json:"counterexamples"`
	UnknownLocations  []UnknownLocation   `json:"unknown_locations"`
	ExtensionEvidence []ExtensionEvidence `json:"extension_evidence"`
	RepositoryWrites  int                 `json:"repository_writes"`
	MutationAuthority bool                `json:"mutation_authority"`
	FactsDigest       string              `json:"facts_digest"`
	Digest            string              `json:"digest"`
}

type SourceObservation struct {
	SourcePath    string `json:"source_path"`
	SourceDigest  string `json:"source_digest"`
	SemanticValue string `json:"semantic_value"`
}

type CausalityManifest struct {
	Schema                      string                  `json:"schema"`
	ManifestID                  string                  `json:"manifest_id"`
	Version                     int                     `json:"version"`
	PredecessorContractID       string                  `json:"predecessor_contract_id"`
	PredecessorContractDigest   string                  `json:"predecessor_contract_digest"`
	CoordinateID                string                  `json:"coordinate_id"`
	CasesPerCandidate           int                     `json:"cases_per_candidate"`
	TransitionDenominator       int                     `json:"transition_denominator"`
	TransitionDenominatorReason string                  `json:"transition_denominator_reason"`
	RequiredReceiptFields       []string                `json:"required_receipt_fields"`
	Cases                       []CausalityCaseContract `json:"cases"`
}

type CausalityCaseContract struct {
	CandidateID          string   `json:"candidate_id"`
	SourcePath           string   `json:"source_path"`
	OperationValueBefore string   `json:"operation_value_before"`
	OperationValueAfter  string   `json:"operation_value_after"`
	NonSemanticComment   string   `json:"non_semantic_comment"`
	RequiredChangeFields []string `json:"required_change_fields"`
}

type CausalityCaseInput struct {
	CaseID      string            `json:"case_id"`
	Kind        string            `json:"kind"`
	Observation SourceObservation `json:"observation"`
	Receipt     Receipt           `json:"receipt"`
}

type CausalitySampleInput struct {
	CandidateID string             `json:"candidate_id"`
	Baseline    CausalityCaseInput `json:"baseline"`
	Semantic    CausalityCaseInput `json:"semantic"`
	NonSemantic CausalityCaseInput `json:"nonsemantic"`
}

type CausalityInput struct {
	SubjectSHA string                 `json:"subject_sha"`
	Contract   Contract               `json:"contract"`
	Manifest   CausalityManifest      `json:"manifest"`
	Samples    []CausalitySampleInput `json:"samples"`
}

type CausalCaseResult struct {
	CaseID                  string            `json:"case_id"`
	Kind                    string            `json:"kind"`
	SourcePath              string            `json:"source_path"`
	SourceDigest            string            `json:"source_digest"`
	SemanticValue           string            `json:"semantic_value"`
	ReceiptDigest           string            `json:"receipt_digest"`
	Decision                string            `json:"decision"`
	Status                  string            `json:"status"`
	Stage                   string            `json:"stage"`
	Step                    string            `json:"step"`
	Reason                  string            `json:"reason"`
	ClaimTransitions        []ClaimTransition `json:"claim_transitions"`
	ReceiptClaimTransitions []ClaimTransition `json:"receipt_claim_transitions"`
	CoordinateVector        []Coordinate      `json:"coordinate_vector"`
}

type CausalitySampleResult struct {
	CandidateID                       string           `json:"candidate_id"`
	Baseline                          CausalCaseResult `json:"baseline"`
	Semantic                          CausalCaseResult `json:"semantic"`
	NonSemantic                       CausalCaseResult `json:"nonsemantic"`
	SourceSemanticValueChanged        bool             `json:"source_semantic_value_changed"`
	SemanticProjectionChanged         bool             `json:"semantic_projection_changed"`
	DecisionChanged                   bool             `json:"decision_changed"`
	ClaimTransitionsChanged           bool             `json:"claim_transitions_changed"`
	SourceDigestChanged               bool             `json:"source_digest_changed"`
	NonSemanticSourceDigestChanged    bool             `json:"nonsemantic_source_digest_changed"`
	NonSemanticSemanticValuePreserved bool             `json:"nonsemantic_semantic_value_preserved"`
	NonSemanticProjectionChanged      bool             `json:"nonsemantic_projection_changed"`
	NonSemanticDecisionChanged        bool             `json:"nonsemantic_decision_changed"`
	RequiredChangeFields              []string         `json:"required_change_fields"`
	ChangedFields                     []string         `json:"changed_fields"`
	DigestOnlyBinding                 bool             `json:"digest_only_binding"`
	HardcodedFixture                  bool             `json:"hardcoded_fixture"`
	Status                            string           `json:"status"`
	Stage                             string           `json:"stage"`
	Step                              string           `json:"step"`
	Reason                            string           `json:"reason"`
}

type CausalCaseCount struct {
	Observed int `json:"observed"`
	Total    int `json:"total"`
}

type TransitionBucket struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

type CausalityTransitionSummary struct {
	FixedDenominator int              `json:"fixed_denominator"`
	Refuted          TransitionBucket `json:"refuted"`
	Discharged       TransitionBucket `json:"discharged"`
	Open             TransitionBucket `json:"open"`
	Reason           string           `json:"reason"`
}

type CausalityUnknown struct {
	CandidateID string `json:"candidate_id"`
	CaseID      string `json:"case_id"`
	Stage       string `json:"stage"`
	Step        string `json:"step"`
	Reason      string `json:"reason"`
}

type CausalitySummary struct {
	CausalCases           CausalCaseCount `json:"causal_cases"`
	DigestOnlyCases       int             `json:"digest_only_cases"`
	HardcodedFixtureCases int             `json:"hardcoded_fixture_cases"`
	Unknowns              int             `json:"unknowns"`
}

type CausalityReport struct {
	Schema            string                     `json:"schema"`
	Decision          string                     `json:"decision"`
	Resolution        string                     `json:"resolution"`
	Reason            string                     `json:"reason"`
	Interpretation    string                     `json:"interpretation"`
	SubjectSHA        string                     `json:"subject_sha"`
	ContractID        string                     `json:"contract_id"`
	Manifest          CausalityManifest          `json:"manifest"`
	Samples           []CausalitySampleResult    `json:"samples"`
	Summary           CausalitySummary           `json:"summary"`
	TransitionSummary CausalityTransitionSummary `json:"transition_summary"`
	UnknownFindings   []CausalityUnknown         `json:"unknown_findings"`
	NotClaimed        []string                   `json:"not_claimed"`
	FactsDigest       string                     `json:"facts_digest"`
	Digest            string                     `json:"digest"`
}
