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
	ReceiptSchema              = "gooo/meta-minimal-causal-explanation/v1"
	JudgmentSchema             = "gooo/meta-minimal-causal-explanation-judgment/v1"
	GraphSchema                = "gooo/meta-minimal-causal-graph/v1"
	SourceSchema               = "gooo/meta-minimal-causal-explanation-source/v1"
	ExplanationAuthority       = "PATH_SET"
	ExplanationTextRole        = "INCIDENTAL"
	CaseTotal                  = 3
	PathTotal                  = 3
	MinimalPathTotal           = 1
	SufficientPathTotal        = 2
	CounterfactualTotal        = 7
	ChangedCounterfactualTotal = 6
	ClaimTotal                 = 6
	TransitionTotal            = ClaimTotal * 2
	IndicatorTotal             = 12
)

const (
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	CaseAccepted       = "ACCEPTED"
	CaseRejected       = "REJECTED"
	ClaimOpen          = "OPEN"
	ClaimDischarged    = "DISCHARGED"
)

var subjectPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type MetaOperation struct {
	ID          string `json:"id"`
	Activity    string `json:"activity"`
	Producer    string `json:"producer"`
	Consumer    string `json:"consumer"`
	ProofChoice string `json:"proof_choice"`
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
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type CausalGraph struct {
	Schema       string       `json:"schema"`
	DecisionRule string       `json:"decision_rule"`
	Nodes        []CausalNode `json:"nodes"`
	Edges        []CausalEdge `json:"edges"`
	Digest       string       `json:"digest"`
}

type Counterfactual struct {
	RemovedEvidenceID string     `json:"removed_evidence_id"`
	BeforeDecision    string     `json:"before_decision"`
	AfterDecision     string     `json:"after_decision"`
	Changed           bool       `json:"changed"`
	Coordinate        Coordinate `json:"coordinate"`
}

type ExplanationPath struct {
	ID              string           `json:"id"`
	EvidenceIDs     []string         `json:"evidence_ids"`
	EdgeIDs         []string         `json:"edge_ids"`
	Decision        string           `json:"decision"`
	Sufficient      bool             `json:"sufficient"`
	Minimal         bool             `json:"minimal"`
	Counterfactuals []Counterfactual `json:"counterfactuals"`
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
	Coordinate               Coordinate `json:"coordinate"`
	PreviousTransitionDigest string     `json:"previous_transition_digest"`
	TransitionDigest         string     `json:"transition_digest"`
}

type Preservation struct {
	ClaimTotal      int    `json:"claim_total"`
	PreservedTotal  int    `json:"preserved_total"`
	TransitionTotal int    `json:"transition_total"`
	TransitionHead  string `json:"transition_head"`
	Policy          string `json:"policy"`
}

type Summary struct {
	CasesTotal                 int    `json:"cases_total"`
	CasesAccepted              int    `json:"cases_accepted"`
	PathsObserved              int    `json:"paths_observed"`
	MinimalSufficientPaths     int    `json:"minimal_sufficient_paths"`
	SufficientNonminimalPaths  int    `json:"sufficient_nonminimal_paths"`
	InsufficientPaths          int    `json:"insufficient_paths"`
	CounterfactualTotal        int    `json:"counterfactual_total"`
	ChangedCounterfactualTotal int    `json:"changed_counterfactual_total"`
	CandidateEvidenceTotal     int    `json:"candidate_evidence_total"`
	RepositoryWrites           int    `json:"repository_writes"`
	PromotionAuthorized        bool   `json:"promotion_authorized"`
	PathSetAuthoritative       bool   `json:"path_set_authoritative"`
	ExplanationTextRole        string `json:"explanation_text_role"`
}

type Receipt struct {
	Schema           string            `json:"schema"`
	Source           SourceBinding     `json:"source"`
	Subject          Subject           `json:"subject"`
	Program          MetaProgram       `json:"program"`
	Graph            CausalGraph       `json:"graph"`
	Cases            []ExplanationCase `json:"cases"`
	Summary          Summary           `json:"summary"`
	Preservation     Preservation      `json:"preservation"`
	ClaimTransitions []ClaimTransition `json:"claim_transitions"`
	Indicators       []Indicator       `json:"indicators"`
	Decision         string            `json:"decision"`
	Resolution       string            `json:"resolution"`
	Authority        Authority         `json:"authority"`
	ReceiptDigest    string            `json:"receipt_digest"`
}

type SourceBinding struct {
	Schema string `json:"schema"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Lines  int    `json:"lines"`
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
	Schema                  string `json:"schema"`
	Status                  string `json:"status"`
	Decision                string `json:"decision"`
	Resolution              string `json:"resolution"`
	PathSetVerified         bool   `json:"path_set_verified"`
	CounterfactualsVerified bool   `json:"counterfactuals_verified"`
	ClaimsPreserved         bool   `json:"claims_preserved"`
	PromotionAuthorized     bool   `json:"promotion_authorized"`
	ReceiptDigest           string `json:"receipt_digest"`
	JudgmentDigest          string `json:"judgment_digest"`
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
	return strings.Count(string(data), "\n")
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
