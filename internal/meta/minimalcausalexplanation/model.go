package minimalcausalexplanation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	ReceiptSchema         = "gooo/meta-minimal-causal-explanation/v2"
	JudgmentSchema        = "gooo/meta-minimal-causal-explanation-judgment/v2"
	GraphSchema           = "gooo/meta-minimal-causal-graph/v2"
	SourceSchema          = "gooo/meta-minimal-causal-explanation-source/v2"
	ObservationSchema     = "gooo/meta-minimal-causal-explanation-observation/v1"
	RepositorySchema      = "gooo/meta-minimal-causal-explanation-repository/v1"
	IndependenceSchema    = "gooo/meta-minimal-causal-explanation-independence/v1"
	CompilerReceiptSchema = "gooo.language.value-witness/v2"
	ExplanationAuthority  = "PATH_SET"
	ExplanationTextRole   = "INCIDENTAL"
	StatusUnknown         = "UNKNOWN"
	StatusPass            = "PASS"
	StatusFailClosed      = "FAIL_CLOSED"
	StatusRefuted         = "REFUTED"
	ClaimOpen             = "OPEN"
	ClaimDischarged       = "DISCHARGED"
	ClaimRefuted          = "REFUTED"
	CaseAccepted          = "ACCEPTED"
	CaseRejected          = "REJECTED"
	SubsetMinimal         = "SUBSET_MINIMAL"
	NotSubsetMinimal      = "NOT_SUBSET_MINIMAL"
	CardinalityMinimum    = "CARDINALITY_MINIMUM"
	NotCardinalityMinimum = "NOT_CARDINALITY_MINIMUM"
	CardinalityUnknown    = "UNKNOWN"
)

var subjectPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type SourceReconstruction struct {
	ASTParsed                  bool   `json:"ast_parsed"`
	IRLowered                  bool   `json:"ir_lowered"`
	SemanticDigest             string `json:"semantic_digest"`
	GraphReconstructed         bool   `json:"graph_reconstructed"`
	PredicateReconstructed     bool   `json:"predicate_reconstructed"`
	ProducerPackageImportCount int    `json:"producer_package_import_count"`
	ProducerPackageImportTotal int    `json:"producer_package_import_total"`
}

type MetaOperation struct {
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	ProofChoice    string `json:"proof_choice"`
	EvidenceDigest string `json:"evidence_digest"`
}

type MetaProgram struct {
	Schema               string          `json:"schema"`
	Producer             string          `json:"producer"`
	Consumer             string          `json:"consumer"`
	IndicatorDenominator int             `json:"indicator_denominator"`
	MetaOperations       []MetaOperation `json:"meta_operations"`
}

type CausalNode struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Producer string `json:"producer"`
	Consumer string `json:"consumer"`
}

type CausalEdge struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	ViaActivity string `json:"via_activity"`
	Relation    string `json:"relation"`
	Causal      bool   `json:"causal"`
}

type CausalGraph struct {
	Schema       string       `json:"schema"`
	DecisionRule string       `json:"decision_rule"`
	Nodes        []CausalNode `json:"nodes"`
	Edges        []CausalEdge `json:"edges"`
	Digest       string       `json:"digest"`
}

type DecisionPredicate struct {
	Value           string   `json:"value"`
	RequiredRoles   []string `json:"required_roles"`
	DecisionOutput  string   `json:"decision_output"`
	PriorClaimState string   `json:"prior_claim_state"`
}

type Evidence struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	Origin     string `json:"origin"`
	Status     string `json:"status"`
	Digest     string `json:"digest"`
	Provenance string `json:"provenance"`
}

type Counterfactual struct {
	ExecutionID       string     `json:"execution_id"`
	RemovedEvidenceID string     `json:"removed_evidence_id"`
	Origin            string     `json:"origin"`
	BeforeDecision    string     `json:"before_decision"`
	AfterDecision     string     `json:"after_decision"`
	Changed           bool       `json:"changed"`
	Coordinate        Coordinate `json:"coordinate"`
	EvidenceDigest    string     `json:"evidence_digest"`
}

type CombinationSearch struct {
	CorpusEvidenceIDs                 []string `json:"corpus_evidence_ids"`
	Exhaustive                        bool     `json:"exhaustive"`
	EnumeratedCombinationTotal        int      `json:"enumerated_combination_total"`
	SmallerCombinationTotal           int      `json:"smaller_combination_total"`
	SufficientSmallerCombinationTotal int      `json:"sufficient_smaller_combination_total"`
}

type ExplanationPath struct {
	ID                   string            `json:"id"`
	EvidenceIDs          []string          `json:"evidence_ids"`
	EdgeIDs              []string          `json:"edge_ids"`
	Decision             string            `json:"decision"`
	Sufficient           bool              `json:"sufficient"`
	SubsetMinimal        string            `json:"subset_minimal"`
	CardinalityMinimum   string            `json:"cardinality_minimum"`
	SingleRemovalChanged int               `json:"single_removal_changed"`
	SingleRemovalTotal   int               `json:"single_removal_total"`
	Counterfactuals      []Counterfactual  `json:"counterfactuals"`
	CombinationSearch    CombinationSearch `json:"combination_search"`
}

type ExplanationCase struct {
	ID                     string            `json:"id"`
	Kind                   string            `json:"kind"`
	ExplanationText        string            `json:"explanation_text"`
	AvailableEvidenceTotal int               `json:"available_evidence_total"`
	Paths                  []ExplanationPath `json:"paths"`
	ExpectedDecision       string            `json:"expected_decision"`
	Verdict                string            `json:"verdict"`
}

type Indicator struct {
	ID             string `json:"id"`
	Class          string `json:"class"`
	MetaOperation  string `json:"meta_operation"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	ProofChoice    string `json:"proof_choice"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

type ClaimTransition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	EvidenceDigest           string     `json:"evidence_digest"`
	Provenance               string     `json:"provenance"`
	Coordinate               Coordinate `json:"coordinate"`
	PreviousTransitionDigest string     `json:"previous_transition_digest"`
	TransitionDigest         string     `json:"transition_digest"`
}

type ClaimRegression struct {
	ScenarioID               string            `json:"scenario_id"`
	ReceiptDecision          string            `json:"receipt_decision"`
	LegacyUnconditionalState string            `json:"legacy_unconditional_state"`
	CorrectState             string            `json:"correct_state"`
	Transitions              []ClaimTransition `json:"transitions"`
}

type RepositoryObservation struct {
	Schema              string `json:"schema"`
	Status              string `json:"status"`
	WorkspaceWrites     bool   `json:"workspace_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type RepositoryBoundary struct {
	Before              RepositoryObservation `json:"before"`
	After               RepositoryObservation `json:"after"`
	Writes              int                   `json:"writes"`
	PromotionAuthorized bool                  `json:"promotion_authorized"`
}

type Intervention struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	BeforeSourceDigest      string `json:"before_source_digest"`
	AfterSourceDigest       string `json:"after_source_digest"`
	BeforeSemanticDigest    string `json:"before_semantic_digest"`
	AfterSemanticDigest     string `json:"after_semantic_digest"`
	BeforeDecision          string `json:"before_decision"`
	AfterDecision           string `json:"after_decision"`
	SemanticChanged         bool   `json:"semantic_changed"`
	SemanticDigestPreserved bool   `json:"semantic_digest_preserved"`
	ResultPreserved         bool   `json:"result_preserved"`
	PathSetChanged          bool   `json:"path_set_changed"`
	MinimalityChanged       bool   `json:"minimality_changed"`
	ClaimTransitionChanged  bool   `json:"claim_transition_changed"`
	Provenance              string `json:"provenance"`
}

type Preservation struct {
	ClaimTotal      int    `json:"claim_total"`
	PreservedTotal  int    `json:"preserved_total"`
	TransitionTotal int    `json:"transition_total"`
	TransitionHead  string `json:"transition_head"`
	Policy          string `json:"policy"`
	RegressionTotal int    `json:"regression_total"`
}

type Summary struct {
	CasesTotal                     int    `json:"cases_total"`
	PathsObserved                  int    `json:"paths_observed"`
	ObservedEvidenceTotal          int    `json:"observed_evidence_total"`
	SyntheticEvidenceTotal         int    `json:"synthetic_evidence_total"`
	CandidateEvidenceTotal         int    `json:"candidate_evidence_total"`
	SufficientPaths                int    `json:"sufficient_paths"`
	SubsetMinimalNumerator         int    `json:"subset_minimal_numerator"`
	SubsetMinimalDenominator       int    `json:"subset_minimal_denominator"`
	CardinalityMinimumNumerator    int    `json:"cardinality_minimum_numerator"`
	CardinalityMinimumDenominator  int    `json:"cardinality_minimum_denominator"`
	CardinalityUnknownPaths        int    `json:"cardinality_unknown_paths"`
	InsufficientPaths              int    `json:"insufficient_paths"`
	CounterfactualExecutions       int    `json:"counterfactual_executions"`
	ChangedCounterfactuals         int    `json:"changed_counterfactuals"`
	CombinationExecutions          int    `json:"combination_executions"`
	ClaimTransitionTotal           int    `json:"claim_transition_total"`
	RegressionClaimTransitionTotal int    `json:"regression_claim_transition_total"`
	RepositoryWrites               int    `json:"repository_writes"`
	PromotionAuthorized            bool   `json:"promotion_authorized"`
	PathSetAuthoritative           bool   `json:"path_set_authoritative"`
	ExplanationTextRole            string `json:"explanation_text_role"`
	ObservedCounterfactuals        int    `json:"observed_counterfactuals"`
	SyntheticCounterfactuals       int    `json:"synthetic_counterfactuals"`
}

type Receipt struct {
	Schema                string               `json:"schema"`
	Source                SourceBinding        `json:"source"`
	Subject               Subject              `json:"subject"`
	Reconstruction        SourceReconstruction `json:"reconstruction"`
	Program               MetaProgram          `json:"program"`
	Predicate             DecisionPredicate    `json:"predicate"`
	Graph                 CausalGraph          `json:"graph"`
	Evidence              []Evidence           `json:"evidence"`
	ObservedReceiptDigest string               `json:"observed_receipt_digest"`
	Cases                 []ExplanationCase    `json:"cases"`
	Summary               Summary              `json:"summary"`
	Repository            RepositoryBoundary   `json:"repository"`
	Preservation          Preservation         `json:"preservation"`
	ClaimTransitions      []ClaimTransition    `json:"claim_transitions"`
	ClaimRegression       ClaimRegression      `json:"claim_regression"`
	Interventions         []Intervention       `json:"interventions"`
	Indicators            []Indicator          `json:"indicators"`
	Decision              string               `json:"decision"`
	Resolution            string               `json:"resolution"`
	Authority             Authority            `json:"authority"`
	ReceiptDigest         string               `json:"receipt_digest"`
}

type SourceBinding struct {
	Schema         string `json:"schema"`
	Path           string `json:"path"`
	Digest         string `json:"digest"`
	Lines          int    `json:"lines"`
	SemanticDigest string `json:"semantic_digest"`
}

type Subject struct {
	Repository string `json:"repository"`
	SHA        string `json:"sha"`
}

type Authority struct {
	RepositoryWorkspaceWrites  bool `json:"repository_workspace_writes"`
	PromotionAuthorized        bool `json:"promotion_authorized"`
	SemanticMutationAuthorized bool `json:"semantic_mutation_authorized"`
}

type Judgment struct {
	Schema                    string `json:"schema"`
	Status                    string `json:"status"`
	Decision                  string `json:"decision"`
	Resolution                string `json:"resolution"`
	SourceReconstructed       bool   `json:"source_reconstructed"`
	PathSetVerified           bool   `json:"path_set_verified"`
	CounterfactualsVerified   bool   `json:"counterfactuals_verified"`
	ClaimsPreserved           bool   `json:"claims_preserved"`
	InterventionsVerified     bool   `json:"interventions_verified"`
	ProducerImportCount       int    `json:"producer_import_count"`
	ProducerImportDenominator int    `json:"producer_import_denominator"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	ReceiptDigest             string `json:"receipt_digest"`
	JudgmentDigest            string `json:"judgment_digest"`
}

type RawCompilerReceipt struct {
	Schema              string `json:"schema"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	Resolution          string `json:"resolution"`
	HeadSHA             string `json:"head_sha"`
	SourcePath          string `json:"source_path"`
	SourceDigest        string `json:"source_digest"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	CoreIRFingerprint   string `json:"core_ir_fingerprint"`
	ValueProgram        string `json:"value_program"`
}

func digestValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func contentDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n") + boolInt(data[len(data)-1] != '\n')
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func validSubject(repository, sha string) error {
	if strings.TrimSpace(repository) == "" || !subjectPattern.MatchString(sha) {
		return fmt.Errorf("subject repository and lowercase 40-character SHA are required")
	}
	return nil
}

func ReceiptDigest(receipt Receipt) string {
	receipt.ReceiptDigest = ""
	digest, _ := digestValue(receipt)
	return digest
}

func digestString(value string) string {
	return contentDigest([]byte(value))
}
