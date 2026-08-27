package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

const (
	ContractSchema      = "gooo/invariant-transformation-contract/v2"
	ReceiptSchema       = "gooo/invariant-transformation-receipt/v2"
	ReportSchema        = "gooo/invariant-transformation-report/v2"
	DenominatorID       = "gooo/invariant-transformation-denominator/v2"
	SourcePath          = "examples/invariant-transformation/main.gooo"
	ProducerID          = "invarianttransformation.producer"
	ConsumerID          = "invarianttransformation.independent-judge"
	AuthorityOp         = "authorize-bounded-transformation-witness"
	EffectNoWrite       = "NO_EFFECT"
	EffectApproved      = "APPROVED_ARTIFACT_RECORDED"
	DecisionAllowed     = "AUTHORIZED"
	DecisionBlocked     = "BLOCKED"
	DecisionRefuted     = "REFUTED"
	DecisionPass        = "PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	StatusOpen          = "OPEN"
	StatusDischarged    = "DISCHARGED"
	StatusRefuted       = "REFUTED"
	ProofFoundation     = "FOUNDATION"
	ProofCoherence      = "COHERENCE"
	ProofRegression     = "REGRESSION"
	AuthorityScope      = "BOUNDED_TRANSFORMATION_RECEIPT_OR_TEMP_ARTIFACT_EMISSION"
	InputDomainID       = "bounded-fixture-input-domain-v1"
	InvariantID         = "candidate-output-equals-expected-v1"
	ReceiptProvisional  = "PROVISIONAL_NO_EFFECT"
	ReceiptExecuted     = "EFFECT_EXECUTED"
	UnknownEffectScope  = "UNKNOWN"
	ExecutorID          = "invarianttransformation.post-judgment-executor"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ValueSpec struct {
	ID                string     `json:"id"`
	Kind              string     `json:"kind"`
	Producer          string     `json:"producer"`
	Consumer          string     `json:"consumer"`
	MetaOperation     string     `json:"meta_operation"`
	ProofChoice       string     `json:"proof_choice"`
	VerificationCheck string     `json:"verification_check"`
	Coordinate        Coordinate `json:"coordinate"`
}

// ValidatorExpectation is a labeled validator expectation. It never selects
// a source case, supplies a recipe, or authorizes an operation/effect.
type ValidatorExpectation struct {
	CaseID              string `json:"case_id"`
	ExpectedDecision    string `json:"expected_decision"`
	ExpectedResolution  string `json:"expected_resolution"`
	ExpectedReason      string `json:"expected_reason"`
	ExpectedStatus      string `json:"expected_status"`
	ExpectedEffectCount int    `json:"expected_effect_count"`
}

type Contract struct {
	Schema                string                 `json:"schema"`
	Version               int                    `json:"version"`
	Denominator           string                 `json:"denominator_id"`
	Source                string                 `json:"source"`
	Values                []ValueSpec            `json:"values"`
	ValidatorExpectations []ValidatorExpectation `json:"validator_expectations"`
}

type Transition struct {
	ClaimID                  string     `json:"claim_id"`
	From                     string     `json:"from"`
	To                       string     `json:"to"`
	Coordinate               Coordinate `json:"coordinate"`
	PropositionDigest        string     `json:"proposition_digest"`
	PriorStateDigest         string     `json:"prior_state_digest"`
	EvidenceDigest           string     `json:"evidence_digest"`
	PreviousTransitionDigest string     `json:"previous_transition_digest"`
	CurrentTransitionDigest  string     `json:"current_transition_digest"`
}

type MetaValue struct {
	ID                string     `json:"id"`
	Kind              string     `json:"kind"`
	Value             string     `json:"value"`
	EvidenceDigest    string     `json:"evidence_digest"`
	Producer          string     `json:"producer"`
	Consumer          string     `json:"consumer"`
	MetaOperation     string     `json:"meta_operation"`
	ProofChoice       string     `json:"proof_choice"`
	VerificationCheck string     `json:"verification_check"`
	Coordinate        Coordinate `json:"coordinate"`
}

type Claim struct {
	ID                string       `json:"id"`
	Status            string       `json:"status"`
	Reason            string       `json:"reason"`
	VerificationCheck string       `json:"verification_check"`
	Coordinate        Coordinate   `json:"coordinate"`
	TargetDigest      string       `json:"target_digest"`
	PriorStateDigest  string       `json:"prior_state_digest"`
	EvidenceDigests   []string     `json:"evidence_digests"`
	Transitions       []Transition `json:"transitions"`
}

type TransformationEvidence struct {
	SourceDigest             string `json:"source_digest"`
	SemanticSourceDigest     string `json:"semantic_source_digest"`
	CaseStableID             string `json:"case_stable_id"`
	ActivityStableID         string `json:"activity_stable_id"`
	OperationID              string `json:"operation_id"`
	InputDomainID            string `json:"input_domain_id"`
	InvariantID              string `json:"invariant_id"`
	EffectIntent             string `json:"effect_intent"`
	InputValue               int64  `json:"input_value"`
	CandidateOperation       string `json:"candidate_operation"`
	CandidateResult          int64  `json:"candidate_result"`
	ExpectedValue            int64  `json:"expected_value"`
	Invariant                string `json:"invariant"`
	CandidateDigest          string `json:"candidate_digest"`
	SemanticBeforeDigest     string `json:"semantic_before_digest"`
	SemanticAfterDigest      string `json:"semantic_after_digest"`
	ExpectedSemanticDigest   string `json:"expected_semantic_digest"`
	ReplayRecipe             string `json:"replay_recipe"`
	BaselineInputValue       int64  `json:"baseline_input_value"`
	BaselineOperation        string `json:"baseline_operation"`
	BaselineOutput           int64  `json:"baseline_output"`
	BaselineDigest           string `json:"baseline_digest"`
	ReplayInputValue         int64  `json:"replay_input_value"`
	ReplayOperation          string `json:"replay_operation"`
	ReplayOutput             int64  `json:"replay_output"`
	ReplayDigest             string `json:"replay_digest,omitempty"`
	ReplaySemanticDigest     string `json:"replay_semantic_digest,omitempty"`
	ReplayEvidenceDigest     string `json:"replay_evidence_digest,omitempty"`
	RegressionWitnessPresent bool   `json:"regression_witness_present"`
	ReplayCount              int    `json:"replay_count"`
	ReplayFailureStage       string `json:"replay_failure_stage,omitempty"`
	ReplayFailureStep        string `json:"replay_failure_step,omitempty"`
	ReplayFailureReason      string `json:"replay_failure_reason,omitempty"`
}

type CandidateComputation struct {
	Operation string `json:"operation"`
	Input     int64  `json:"input"`
	Output    int64  `json:"output"`
}

type ArtifactEvidence struct {
	Path                         string `json:"path"`
	ContentDigest                string `json:"content_digest"`
	Size                         int    `json:"size"`
	CaseID                       string `json:"case_id"`
	SubjectSHA                   string `json:"subject_sha"`
	AuthorizationDigest          string `json:"authorization_digest"`
	Producer                     string `json:"producer"`
	Executor                     string `json:"executor"`
	Consumer                     string `json:"consumer"`
	EffectReceiptDigest          string `json:"effect_receipt_digest"`
	RepositoryNetStatusUnchanged bool   `json:"repository_net_status_unchanged"`
}

type Effect struct {
	Kind                              string           `json:"kind"`
	ArtifactID                        string           `json:"artifact_id,omitempty"`
	ArtifactDigest                    string           `json:"artifact_digest,omitempty"`
	ArtifactPath                      string           `json:"artifact_path,omitempty"`
	ArtifactSize                      int              `json:"artifact_size,omitempty"`
	Artifact                          ArtifactEvidence `json:"artifact,omitempty"`
	CaseID                            string           `json:"case_id"`
	SubjectSHA                        string           `json:"subject_sha"`
	Intent                            string           `json:"intent"`
	AuthorizationDigest               string           `json:"authorization_digest"`
	ExecutionReceiptDigest            string           `json:"execution_receipt_digest"`
	Producer                          string           `json:"producer"`
	Executor                          string           `json:"executor"`
	Consumer                          string           `json:"consumer"`
	MetaOperation                     string           `json:"meta_operation"`
	TempArtifactWriteAuthorized       bool             `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool             `json:"repository_net_status_unchanged"`
	RepositoryActualOrTransientWrites string           `json:"repository_actual_or_transient_writes"`
}

type Receipt struct {
	Schema                            string                 `json:"schema"`
	CaseID                            string                 `json:"case_id"`
	CaseKind                          string                 `json:"case_kind"`
	ActivityStableID                  string                 `json:"activity_stable_id"`
	HeadSHA                           string                 `json:"head_sha"`
	SourcePath                        string                 `json:"source_path"`
	SourceDigest                      string                 `json:"source_digest"`
	SemanticSourceDigest              string                 `json:"semantic_source_digest"`
	ContractDigest                    string                 `json:"contract_digest"`
	ValidatorContractDigest           string                 `json:"validator_contract_digest"`
	Producer                          string                 `json:"producer"`
	Consumer                          string                 `json:"consumer"`
	MetaOperation                     string                 `json:"meta_operation"`
	ProofChoice                       string                 `json:"proof_choice"`
	Values                            []MetaValue            `json:"values"`
	Claims                            []Claim                `json:"claims"`
	Evidence                          TransformationEvidence `json:"evidence"`
	Decision                          string                 `json:"decision"`
	Resolution                        string                 `json:"resolution"`
	Reason                            string                 `json:"reason"`
	Phase                             string                 `json:"phase"`
	Effects                           []Effect               `json:"effects"`
	AuthorizationDigest               string                 `json:"authorization_digest"`
	TempArtifactWriteAuthorized       bool                   `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool                   `json:"repository_net_status_unchanged"`
	RepositoryActualOrTransientWrites string                 `json:"repository_actual_or_transient_writes"`
	RepositoryWritesObserved          bool                   `json:"repository_writes_observed"`
	RepositoryWrites                  int                    `json:"repository_writes"`
	MutationAuthority                 bool                   `json:"mutation_authority"`
	AuthorityScope                    string                 `json:"authority_scope"`
	Digest                            string                 `json:"digest"`
}

type Judgment struct {
	Decision            string `json:"decision"`
	Resolution          string `json:"resolution"`
	Reason              string `json:"reason"`
	Status              string `json:"status"`
	Independent         bool   `json:"independent"`
	CheckedClaims       int    `json:"checked_claims"`
	DischargedClaims    int    `json:"discharged_claims"`
	OpenClaims          int    `json:"open_claims"`
	RefutedClaims       int    `json:"refuted_claims"`
	Effects             int    `json:"effects"`
	AuthorizationDigest string `json:"authorization_digest"`
}

type CaseResult struct {
	Expectation                  ValidatorExpectation `json:"validator_expectation"`
	ProvisionalReceiptDigest     string               `json:"provisional_receipt_digest"`
	AuthorizationReceiptDigest   string               `json:"authorization_receipt_digest"`
	ExecutedEffects              int                  `json:"executed_effects"`
	IndependentlyObservedEffects int                  `json:"independently_observed_effects"`
	Receipt                      Receipt              `json:"receipt"`
	Judgment                     Judgment             `json:"judgment"`
	Satisfied                    bool                 `json:"satisfied"`
}

type Summary struct {
	CasesTotal                        int    `json:"cases_total"`
	CasesSatisfied                    int    `json:"cases_satisfied"`
	AuthorizedCases                   int    `json:"authorized_cases"`
	RefutedCases                      int    `json:"refuted_cases"`
	OpenCases                         int    `json:"open_cases"`
	ClaimsTotal                       int    `json:"claims_total"`
	UniqueClaimInstances              int    `json:"unique_claim_instances"`
	ClaimTemplates                    int    `json:"claim_templates"`
	DischargedClaims                  int    `json:"discharged_claims"`
	RefutedClaims                     int    `json:"refuted_claims"`
	OpenClaims                        int    `json:"open_claims"`
	TransitionEvents                  int    `json:"transition_events"`
	AcceptedTransitions               int    `json:"accepted_transitions"`
	SourceDerivedCases                int    `json:"source_derived_cases"`
	BoundedInputDomainObservations    int    `json:"bounded_input_domain_observations"`
	BoundedInputDomainDenominator     int    `json:"bounded_input_domain_denominator"`
	InputDomainCoverageBPS            int    `json:"input_domain_coverage_bps"`
	ProvisionalReceipts               int    `json:"provisional_receipts"`
	AuthorizationReceipts             int    `json:"authorization_receipts"`
	ExecutedEffects                   int    `json:"executed_effects"`
	IndependentlyObservedEffects      int    `json:"independently_observed_effects"`
	UnknownEffectScopes               int    `json:"unknown_effect_scopes"`
	ApprovedArtifactEffects           int    `json:"approved_artifact_effects"`
	RepositoryWrites                  int    `json:"repository_writes"`
	MutationAuthority                 int    `json:"mutation_authority"`
	CoverageBPS                       int    `json:"coverage_bps"`
	CorrectionCount                   int    `json:"correction_count"`
	CorrectionDenominator             int    `json:"correction_denominator"`
	RepositoryNetStatusUnchanged      bool   `json:"repository_net_status_unchanged"`
	TempArtifactWriteAuthorized       bool   `json:"temp_artifact_write_authorized"`
	RepositoryActualOrTransientWrites string `json:"repository_actual_or_transient_writes"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
}

type Report struct {
	Schema                  string       `json:"schema"`
	HeadSHA                 string       `json:"head_sha"`
	SourcePath              string       `json:"source_path"`
	SourceDigest            string       `json:"source_digest"`
	SemanticSourceDigest    string       `json:"semantic_source_digest"`
	ContractDigest          string       `json:"contract_digest"`
	ValidatorContractDigest string       `json:"validator_contract_digest"`
	DenominatorID           string       `json:"denominator_id"`
	DenominatorTotal        int          `json:"denominator_total"`
	Decision                string       `json:"decision"`
	Resolution              string       `json:"resolution"`
	Reason                  string       `json:"reason"`
	Cases                   []CaseResult `json:"cases"`
	Indicators              []Indicator  `json:"indicators"`
	Summary                 Summary      `json:"summary"`
	NotClaimed              []string     `json:"not_claimed"`
	Digest                  string       `json:"digest"`
}

func CanonicalValueSpecs() []ValueSpec {
	return []ValueSpec{
		{ID: "precondition", Kind: "PRECONDITION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "bind-transformation-precondition", ProofChoice: ProofFoundation, VerificationCheck: "source-digest-and-semantic-source-digest", Coordinate: Coordinate{"PRECONDITION", "bind-source-snapshot", "EXACT_SOURCE_SNAPSHOT"}},
		{ID: "transformation", Kind: "TRANSFORMATION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "observe-transformation-result", ProofChoice: ProofCoherence, VerificationCheck: "execute-candidate-and-bind-output", Coordinate: Coordinate{"TRANSFORMATION", "observe-candidate", "TRANSFORMATION_OBSERVED"}},
		{ID: "postcondition", Kind: "POSTCONDITION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "compare-postcondition", ProofChoice: ProofCoherence, VerificationCheck: "compare-lowered-semantic-output-to-invariant", Coordinate: Coordinate{"POSTCONDITION", "compare-semantic-value", "SEMANTIC_POSTCONDITION"}},
		{ID: "regression-witness", Kind: "REGRESSION_WITNESS", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "replay-regression-witness", ProofChoice: ProofRegression, VerificationCheck: "execute-baseline-and-replay-and-compare", Coordinate: Coordinate{"REGRESSION", "execute-replay", "REGRESSION_REPLAY_OBSERVED"}},
	}
}

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, Version: 2, Denominator: DenominatorID, Source: SourcePath,
		Values: CanonicalValueSpecs(),
		ValidatorExpectations: []ValidatorExpectation{
			{CaseID: "preserved-translation", ExpectedDecision: DecisionAllowed, ExpectedResolution: ResolutionExact, ExpectedReason: "ALL_INVARIANTS_DISCHARGED", ExpectedStatus: StatusDischarged, ExpectedEffectCount: 0},
			{CaseID: "semantic-violation", ExpectedDecision: DecisionRefuted, ExpectedResolution: ResolutionInvariant, ExpectedReason: "SEMANTIC_POSTCONDITION_REFUTED", ExpectedStatus: StatusRefuted, ExpectedEffectCount: 0},
			{CaseID: "missing-regression-witness", ExpectedDecision: DecisionBlocked, ExpectedResolution: ResolutionLower, ExpectedReason: "REGRESSION_REPLAY_RECIPE_UNAVAILABLE", ExpectedStatus: StatusOpen, ExpectedEffectCount: 0},
			{CaseID: "approved-artifact", ExpectedDecision: DecisionAllowed, ExpectedResolution: ResolutionExact, ExpectedReason: "ALL_INVARIANTS_DISCHARGED", ExpectedStatus: StatusDischarged, ExpectedEffectCount: 1},
		},
	}
}

func ValidatorExpectationFor(contract Contract, caseID string) (ValidatorExpectation, bool) {
	for _, expectation := range contract.ValidatorExpectations {
		if expectation.CaseID == caseID {
			return expectation, true
		}
	}
	return ValidatorExpectation{}, false
}

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Digest(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(raw)
}

func SemanticDigest(value int64) string {
	return Digest([]string{"int64", strconv.FormatInt(value, 10)})
}

func CandidateDigest(operation string, input, output int64) string {
	return Digest(CandidateComputation{Operation: operation, Input: input, Output: output})
}

func PostconditionDigest(before, after, expected string) string {
	return Digest([]string{before, after, expected})
}

func ReplayDigest(before, after string) string {
	return Digest([]string{before, after, "replay-1"})
}

func NewTransition(claimID, from, to string, coordinate Coordinate, evidenceDigest string) Transition {
	prior := Digest([]string{"claim-state", claimID, from})
	proposition := Digest([]string{"claim-proposition", claimID, coordinate.Stage, coordinate.Step, coordinate.Reason, evidenceDigest})
	previous := Digest([]string{"transition-genesis", claimID, prior})
	transition := Transition{ClaimID: claimID, From: from, To: to, Coordinate: coordinate, PropositionDigest: proposition, PriorStateDigest: prior, EvidenceDigest: evidenceDigest, PreviousTransitionDigest: previous}
	transition.CurrentTransitionDigest = TransitionDigest(transition)
	return transition
}

func TransitionDigest(transition Transition) string {
	transition.CurrentTransitionDigest = ""
	return Digest(transition)
}

func ValueContractDigest() string {
	return Digest(CanonicalValueSpecs())
}

func ValidatorContractDigest() string {
	return Digest(CanonicalContract().ValidatorExpectations)
}

func AuthorizationDigest(receipt Receipt) string {
	receipt.Effects = nil
	receipt.Phase = ReceiptProvisional
	receipt.AuthorizationDigest = ""
	receipt.TempArtifactWriteAuthorized = false
	receipt.Digest = ""
	return Digest(receipt)
}

func EffectExecutionDigest(effect Effect) string {
	effect.ExecutionReceiptDigest = ""
	effect.Artifact.EffectReceiptDigest = ""
	return Digest(effect)
}

func SealReceipt(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = Digest(receipt)
	return receipt
}

func SealReport(report Report) Report {
	report.Digest = ""
	report.Digest = Digest(report)
	return report
}

func ValidDigest(value string) bool { return digestPattern.MatchString(value) }
func ValidHead(value string) bool   { return headPattern.MatchString(value) }

func ValidateContract(contract Contract) error {
	if Digest(contract) != Digest(CanonicalContract()) {
		return fmt.Errorf("invariant transformation validator contract drift")
	}
	return nil
}
