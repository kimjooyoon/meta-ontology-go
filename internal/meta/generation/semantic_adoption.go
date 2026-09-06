package generation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const (
	SemanticAdoptionProposalSchema      = "gooo/semantic-self-adoption-proposal/v1"
	SemanticAdoptionAuthorizationSchema = "gooo/semantic-self-adoption-authorization/v1"
	SemanticAdoptionEvidenceSchema      = "gooo/semantic-self-adoption-evidence/v1"
	SemanticAdoptionReportSchema        = "gooo/semantic-self-adoption-report/v1"
	SemanticAdoptionTarget              = "gooo.generate.semantic-lowering"
	SemanticAdoptionMode                = "bounded-in-memory-reuse"
	SemanticAdoptionAuthorizationMode   = "explicit-human-input"
	SemanticAdoptionUnknownReason       = "MISSING_EXPLICIT_AUTHORIZATION"
	SemanticAdoptionUnknownNext         = "PROVIDE_EXPLICIT_AUTHORIZATION"
	SemanticAdoptionUnknownClass        = "INCOMPLETE_EVIDENCE"
	SemanticAdoptionUnknownStage        = "ADOPTION"
	SemanticAdoptionUnknownStep         = "AUTHORIZE_CANDIDATE"
	SemanticAdoptionRefutedReason       = "ADOPTION_CONTRADICTION"
	SemanticAdoptionClosedReason        = "EXACT_AUTHORIZED_REUSE"
	SemanticAdoptionRefuted             = "REFUTED"
)

// SemanticAdoptionProposal is a caller-owned proposal derived from one stable
// observation candidate. It never grants execution or repository mutation.
type SemanticAdoptionProposal struct {
	Schema            string                       `json:"schema"`
	ObservationDigest string                       `json:"observation_digest"`
	ContractDigest    string                       `json:"contract_digest"`
	InputSourceDigest string                       `json:"input_source_digest"`
	Candidate         SemanticObservationCandidate `json:"candidate"`
	Target            string                       `json:"target"`
	Mode              string                       `json:"mode"`
	ExecutionAllowed  bool                         `json:"execution_allowed"`
	RepositoryWrites  int                          `json:"repository_writes"`
}

// SemanticAdoptionAuthorization is the explicit human-controlled input that
// permits a proposal to exercise the bounded compiler reuse path.
type SemanticAdoptionAuthorization struct {
	Schema               string `json:"schema"`
	AuthorizationID      string `json:"authorization_id"`
	AuthorizationMode    string `json:"authorization_mode"`
	ProposalDigest       string `json:"proposal_digest"`
	CandidateStableID    string `json:"candidate_stable_id"`
	CandidateInputDigest string `json:"candidate_input_digest"`
	ContractDigest       string `json:"contract_digest"`
	InputSourceDigest    string `json:"input_source_digest"`
	Authorized           bool   `json:"authorized"`
	RepositoryWrites     int    `json:"repository_writes"`
	LocalTestExecutions  int    `json:"local_test_executions"`
}

// SemanticAdoptionEvidence records the exact compiler result produced after
// authorization. Counts are execution observations, not inferred savings.
type SemanticAdoptionEvidence struct {
	Schema                string                `json:"schema"`
	ProposalDigest        string                `json:"proposal_digest"`
	AuthorizationDigest   string                `json:"authorization_digest"`
	CandidateStableID     string                `json:"candidate_stable_id"`
	InputDigest           string                `json:"input_digest"`
	BeforeOperationCount  int                   `json:"before_operation_count"`
	AfterOperationCount   int                   `json:"after_operation_count"`
	CacheMisses           int                   `json:"cache_misses"`
	CacheHits             int                   `json:"cache_hits"`
	ReuseApplied          bool                  `json:"reuse_applied"`
	BehaviorEqual         bool                  `json:"behavior_equal"`
	DeterminismEqual      bool                  `json:"determinism_equal"`
	BeforeOutputDigest    string                `json:"before_output_digest"`
	AdoptedOutputDigest   string                `json:"adopted_output_digest"`
	AdoptedReplayDigest   string                `json:"adopted_replay_digest"`
	BeforeSemanticDigest  string                `json:"before_semantic_digest"`
	AdoptedSemanticDigest string                `json:"adopted_semantic_digest"`
	Decision              string                `json:"decision"`
	Reason                string                `json:"reason"`
	Unknown               *EnvelopeUnknownState `json:"unknown"`
	RepositoryWrites      int                   `json:"repository_writes"`
	LocalTestExecutions   int                   `json:"local_test_executions"`
}

// SemanticAdoptionRuntimeMetrics are measured by the caller-owned compiler
// command. Zero-valued build/test fields mean those operations did not run.
type SemanticAdoptionRuntimeMetrics struct {
	WallMS          int64 `json:"wall_ms"`
	AllocationCount int64 `json:"allocation_count"`
	AllocationBytes int64 `json:"allocation_bytes"`
	BuildMS         int64 `json:"build_ms"`
	TestMS          int64 `json:"test_ms"`
	ExecutedTests   int64 `json:"executed_tests"`
	ReusedTests     int64 `json:"reused_tests"`
}

// SemanticAdoptionReport is the single caller-owned artifact for the example
// loop. Proposal and authorization are digest-linked; no source is rewritten.
type SemanticAdoptionReport struct {
	Schema               string                         `json:"schema"`
	Lifecycle            string                         `json:"lifecycle"`
	ObservationDigest    string                         `json:"observation_digest"`
	ProposalDigest       string                         `json:"proposal_digest"`
	AuthorizationDigest  string                         `json:"authorization_digest"`
	Proposal             SemanticAdoptionProposal       `json:"proposal"`
	Authorization        SemanticAdoptionAuthorization  `json:"authorization"`
	Evidence             SemanticAdoptionEvidence       `json:"evidence"`
	Observation          SemanticObservation            `json:"observation"`
	BeforeRuntimeMetrics SemanticAdoptionRuntimeMetrics `json:"before_runtime_metrics"`
	AfterRuntimeMetrics  SemanticAdoptionRuntimeMetrics `json:"after_runtime_metrics"`
	IndependentDecision  string                         `json:"independent_decision"`
	IndependentReason    string                         `json:"independent_reason"`
	RepositoryWrites     int                            `json:"repository_writes"`
	LocalTestExecutions  int                            `json:"local_test_executions"`
}

func AdoptionUnknownState() *EnvelopeUnknownState {
	return &EnvelopeUnknownState{
		Stage: SemanticAdoptionUnknownStage, Step: SemanticAdoptionUnknownStep,
		Reason: SemanticAdoptionUnknownReason, UnknownClass: SemanticAdoptionUnknownClass,
		NextOperation: SemanticAdoptionUnknownNext, BlockedBy: []string{"explicit_authorization"},
	}
}

func SemanticAdoptionProposalDigest(proposal SemanticAdoptionProposal) (string, error) {
	digest, err := cache.DigestOf(proposal)
	if err != nil {
		return "", fmt.Errorf("adoption proposal digest: %w", err)
	}
	return digest.String(), nil
}

func ValidateSemanticAdoptionProposal(proposal SemanticAdoptionProposal) error {
	if proposal.Schema != SemanticAdoptionProposalSchema || proposal.ObservationDigest == "" ||
		!cache.Digest(proposal.ObservationDigest).Known() || !cache.Digest(proposal.ContractDigest).Known() ||
		!cache.Digest(proposal.InputSourceDigest).Known() || proposal.Target != SemanticAdoptionTarget ||
		proposal.Mode != SemanticAdoptionMode || proposal.ExecutionAllowed || proposal.RepositoryWrites != 0 {
		return errors.New("semantic adoption proposal is not a bounded proposal-only artifact")
	}
	if proposal.Candidate.StableID == "" || proposal.Candidate.Phase != SemanticObservationPhase ||
		proposal.Candidate.OperationID != SemanticObservationOperationID ||
		!cache.Digest(proposal.Candidate.InputDigest).Known() || proposal.Candidate.ObservedCount < 2 ||
		proposal.Candidate.ExpectedReducibleCount != proposal.Candidate.ObservedCount-1 ||
		proposal.Candidate.SafetyAssessment != "UNKNOWN_NOT_INFERRED" ||
		proposal.Candidate.BenefitAssessment != "UNKNOWN_NOT_INFERRED" || len(proposal.Candidate.SourceSpans) == 0 {
		return errors.New("semantic adoption proposal candidate is invalid")
	}
	for _, span := range proposal.Candidate.SourceSpans {
		if !validSemanticObservationSpan(span) {
			return errors.New("semantic adoption proposal candidate span is invalid")
		}
	}
	return nil
}

func ValidateSemanticAdoptionAuthorization(authorization SemanticAdoptionAuthorization) error {
	if authorization.Schema != SemanticAdoptionAuthorizationSchema || authorization.AuthorizationID == "" ||
		authorization.AuthorizationMode != SemanticAdoptionAuthorizationMode ||
		!cache.Digest(authorization.ProposalDigest).Known() || authorization.CandidateStableID == "" ||
		!cache.Digest(authorization.CandidateInputDigest).Known() ||
		!cache.Digest(authorization.ContractDigest).Known() || !cache.Digest(authorization.InputSourceDigest).Known() ||
		authorization.RepositoryWrites != 0 || authorization.LocalTestExecutions != 0 {
		return errors.New("semantic adoption authorization is invalid")
	}
	return nil
}

func ValidateSemanticAdoptionEvidence(evidence SemanticAdoptionEvidence) error {
	if evidence.Schema != SemanticAdoptionEvidenceSchema || !cache.Digest(evidence.ProposalDigest).Known() ||
		!cache.Digest(evidence.AuthorizationDigest).Known() || evidence.CandidateStableID == "" ||
		!cache.Digest(evidence.InputDigest).Known() || evidence.RepositoryWrites != 0 || evidence.LocalTestExecutions != 0 {
		return errors.New("semantic adoption evidence is invalid")
	}
	if evidence.Decision == "UNKNOWN" {
		if evidence.Unknown == nil || evidence.Reason != SemanticAdoptionUnknownReason || !SameAdoptionUnknown(evidence.Unknown, AdoptionUnknownState()) ||
			evidence.BeforeOperationCount != 0 || evidence.AfterOperationCount != 0 || evidence.CacheMisses != 0 || evidence.CacheHits != 0 || evidence.ReuseApplied ||
			evidence.BehaviorEqual || evidence.DeterminismEqual || evidence.BeforeOutputDigest != "" || evidence.AdoptedOutputDigest != "" || evidence.AdoptedReplayDigest != "" ||
			evidence.BeforeSemanticDigest != "" || evidence.AdoptedSemanticDigest != "" {
			return errors.New("semantic adoption UNKNOWN evidence is not causal")
		}
		return nil
	}
	if evidence.Decision == SemanticAdoptionRefuted {
		if evidence.Reason != SemanticAdoptionRefutedReason || evidence.Unknown != nil {
			return errors.New("semantic adoption REFUTED evidence is not causal")
		}
		return nil
	}
	if evidence.Decision != "CLOSED" || evidence.Reason != SemanticAdoptionClosedReason || evidence.Unknown != nil {
		return errors.New("semantic adoption evidence decision is invalid")
	}
	return nil
}

func VerifySemanticAdoption(proposal SemanticAdoptionProposal, proposalDigest string, authorization SemanticAdoptionAuthorization, authorizationDigest string, evidence SemanticAdoptionEvidence) (string, string, *EnvelopeUnknownState, error) {
	if err := ValidateSemanticAdoptionProposal(proposal); err != nil {
		return "", "", nil, err
	}
	if !cache.Digest(proposalDigest).Known() || !cache.Digest(authorizationDigest).Known() {
		return "", "", nil, errors.New("semantic adoption artifact digest is unknown")
	}
	if err := ValidateSemanticAdoptionAuthorization(authorization); err != nil {
		return "", "", nil, err
	}
	if err := ValidateSemanticAdoptionEvidence(evidence); err != nil {
		return "", "", nil, err
	}
	if authorization.ProposalDigest != proposalDigest || authorization.CandidateStableID != proposal.Candidate.StableID ||
		authorization.CandidateInputDigest != proposal.Candidate.InputDigest ||
		authorization.ContractDigest != proposal.ContractDigest || authorization.InputSourceDigest != proposal.InputSourceDigest {
		return SemanticAdoptionRefuted, SemanticAdoptionRefutedReason, nil, nil
	}
	if !authorization.Authorized {
		return "UNKNOWN", SemanticAdoptionUnknownReason, AdoptionUnknownState(), nil
	}
	if evidence.ProposalDigest != proposalDigest || evidence.AuthorizationDigest != authorizationDigest ||
		evidence.CandidateStableID != proposal.Candidate.StableID || evidence.InputDigest != proposal.Candidate.InputDigest ||
		evidence.RepositoryWrites != 0 || evidence.LocalTestExecutions != 0 {
		return SemanticAdoptionRefuted, SemanticAdoptionRefutedReason, nil, nil
	}
	if evidence.BeforeOperationCount != 2 || evidence.AfterOperationCount != 1 || evidence.CacheMisses != 1 ||
		evidence.CacheHits != 1 || !evidence.ReuseApplied || !evidence.BehaviorEqual || !evidence.DeterminismEqual ||
		evidence.BeforeOutputDigest == "" || evidence.AdoptedOutputDigest == "" || evidence.AdoptedReplayDigest == "" ||
		evidence.BeforeOutputDigest != evidence.AdoptedOutputDigest || evidence.AdoptedOutputDigest != evidence.AdoptedReplayDigest ||
		evidence.BeforeSemanticDigest == "" || evidence.BeforeSemanticDigest != evidence.AdoptedSemanticDigest {
		return SemanticAdoptionRefuted, SemanticAdoptionRefutedReason, nil, nil
	}
	return "CLOSED", SemanticAdoptionClosedReason, nil, nil
}

func ValidateBoundSemanticAdoption(observation SemanticObservation, proposal SemanticAdoptionProposal, evidence SemanticAdoptionEvidence) error {
	if len(observation.Candidates) != 1 || observation.Candidates[0].StableID != proposal.Candidate.StableID ||
		observation.Candidates[0].InputDigest != proposal.Candidate.InputDigest ||
		evidence.Unknown != nil {
		return errors.New("semantic adoption is not bound to the observed candidate")
	}
	if evidence.BeforeOperationCount != observation.PairEvidence.BeforeOperationCount ||
		evidence.AfterOperationCount != observation.PairEvidence.AfterOperationCount ||
		!observation.PairEvidence.ChangeAdopted || observation.PairEvidence.BehaviorEqual != evidence.BehaviorEqual ||
		observation.PairEvidence.DeterminismEqual != evidence.DeterminismEqual {
		return errors.New("semantic adoption pair evidence is not bound")
	}
	if evidence.Decision == "CLOSED" {
		if evidence.Reason != SemanticAdoptionClosedReason || !observation.PairEvidence.BehaviorEqual || !observation.PairEvidence.DeterminismEqual {
			return errors.New("semantic adoption CLOSED evidence is not bound")
		}
		return nil
	}
	if evidence.Decision == SemanticAdoptionRefuted {
		if evidence.BehaviorEqual && evidence.DeterminismEqual {
			return errors.New("semantic adoption REFUTED evidence is not bound")
		}
		return nil
	}
	return nil
}

func SameAdoptionUnknown(left, right *EnvelopeUnknownState) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason &&
		left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation &&
		strings.Join(left.BlockedBy, "\x00") == strings.Join(right.BlockedBy, "\x00")
}
