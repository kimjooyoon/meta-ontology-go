package policycompilation

import (
	"errors"
	"fmt"
	"strings"
)

const PolicyCounterexampleProposalSchema = "gooo/meta-policy-counterexample-proposal/v1"

type PolicyRevisionCounterexample struct {
	ExpectedSourceDigest string         `json:"expected_source_digest"`
	Input                Case           `json:"case"`
	Observed             DecisionResult `json:"observed"`
}

type PolicyCounterexampleProposal struct {
	Schema                       string                       `json:"schema"`
	State                        string                       `json:"state"`
	Reason                       string                       `json:"reason"`
	SourceFile                   string                       `json:"source_file"`
	SourceDigest                 string                       `json:"source_digest"`
	Counterexample               PolicyRevisionCounterexample `json:"counterexample"`
	CounterexampleDigest         string                       `json:"canonical_counterexample_digest"`
	CounterexampleArtifactDigest string                       `json:"counterexample_artifact_digest,omitempty"`
	Revision                     *PolicyDecisionRevision      `json:"revision_request,omitempty"`
	OriginalPolicy               *CompiledPolicy              `json:"original_policy,omitempty"`
	CandidatePolicy              *CompiledPolicy              `json:"candidate_policy,omitempty"`
	CandidateSource              string                       `json:"candidate_source,omitempty"`
	ChangedCoordinates           []string                     `json:"changed_coordinates"`
	DerivedFields                []string                     `json:"derived_fields"`
	Pending                      *PolicyRevisionPending       `json:"pending,omitempty"`
	Admission                    PolicyRevisionPending        `json:"admission"`
	FailureDetail                string                       `json:"failure_detail,omitempty"`
	InputProvenance              string                       `json:"input_provenance"`
	CandidateExecution           string                       `json:"candidate_execution"`
	Improvement                  string                       `json:"improvement"`
	MutationAuthority            int                          `json:"mutation_authority"`
	PromotionAuthority           int                          `json:"promotion_authority"`
}

func DecodePolicyRevisionCounterexample(data []byte) (PolicyRevisionCounterexample, error) {
	var input PolicyRevisionCounterexample
	err := decodeStrictJSON(data, &input)
	return input, err
}

// ProposePolicyRevisionFromCounterexample derives a request, not an adoption.
// The caller's oracle and observation remain unverified claims. A source/result
// disagreement or stale snapshot cannot be repaired by changing their digests.
func ProposePolicyRevisionFromCounterexample(filename string, source []byte, expectedPackage, expectedNamespace string, input PolicyRevisionCounterexample) (PolicyCounterexampleProposal, error) {
	raw, err := canonicalJSON(input)
	if err != nil {
		return PolicyCounterexampleProposal{}, err
	}
	report := PolicyCounterexampleProposal{
		Schema: PolicyCounterexampleProposalSchema, State: "UNKNOWN",
		SourceFile: filename, SourceDigest: DigestBytes(source),
		Counterexample: input, CounterexampleDigest: DigestBytes(raw),
		ChangedCoordinates: []string{}, DerivedFields: []string{},
		Admission: revisionPending("INDEPENDENT_VALIDATION", "OBSERVE_REVISION_CANDIDATE",
			"INDEPENDENT_REVISION_EVIDENCE_MISSING", "RUN_INDEPENDENT_REVISION_OBSERVER"),
		InputProvenance: "CALLER_DECLARED_NOT_VERIFIED",
		CandidateExecution: "NOT_OBSERVED", Improvement: "UNKNOWN",
	}
	if input.ExpectedSourceDigest == "" {
		return declineCounterexample(report, "UNKNOWN", "SOURCE_PIN_MISSING", "DIRECT_MISSING", "PIN_ORIGINAL_SOURCE")
	}
	if !ValidDigest(input.ExpectedSourceDigest) || input.ExpectedSourceDigest != report.SourceDigest {
		return declineCounterexample(report, "REFUTED", "SOURCE_PIN_MISMATCH", "", "")
	}
	if strings.TrimSpace(input.Input.ID) == "" || strings.TrimSpace(input.Input.EvidenceClass) == "" || strings.TrimSpace(input.Input.Provenance) == "" {
		return declineCounterexample(report, "UNKNOWN", "COUNTEREXAMPLE_METADATA_MISSING", "DIRECT_MISSING", "RECORD_COUNTEREXAMPLE_PROVENANCE")
	}
	if !input.Input.ProducerAvailable || !input.Input.ConsumerAvailable ||
		input.Input.ObservedSourceDigest == "" || input.Input.ObservedArtifactSourceDigest == "" ||
		input.Input.ObservedGeneratedJudgeDigest == "" || input.Input.ObservedIndependentDigest == "" {
		return declineCounterexample(report, "UNKNOWN", "COUNTEREXAMPLE_EVIDENCE_MISSING", "DIRECT_MISSING", "OBSERVE_ORIGINAL_POLICY_EVIDENCE")
	}
	if input.Observed.CaseID == "" || input.Observed.Decision == "" || input.Observed.MatchedCondition == "" {
		return declineCounterexample(report, "UNKNOWN", "COUNTEREXAMPLE_RESULT_MISSING", "DIRECT_MISSING", "OBSERVE_ORIGINAL_POLICY_RESULT")
	}
	if input.Observed.CaseID != input.Input.ID || !knownCondition(input.Observed.MatchedCondition) ||
		!knownDecision(input.Observed.Decision) || !knownDecision(input.Input.ValidatorExpectation) {
		return declineCounterexample(report, "REFUTED", "COUNTEREXAMPLE_IDENTITY_OR_DECISION_INVALID", "", "")
	}
	if input.Observed.Decision == DecisionUnknown {
		return declineCounterexample(report, "UNKNOWN", "ORIGINAL_DECISION_UNKNOWN", "DEPENDENCY_BLOCKED", "RESOLVE_ORIGINAL_POLICY_UNKNOWN")
	}
	if input.Observed.Decision == input.Input.ValidatorExpectation {
		report.State, report.Reason = "NOT_PROPOSED", "NO_DECISION_DIFFERENCE_DECLARED"
		return report, nil
	}
	revision := PolicyDecisionRevision{
		ExpectedSourceDigest: input.ExpectedSourceDigest,
		Condition: input.Observed.MatchedCondition, FromDecision: input.Observed.Decision,
		ToDecision: input.Input.ValidatorExpectation,
	}
	proposal, err := ProposePolicyDecisionRevision(filename, source, expectedPackage, expectedNamespace, revision)
	if err != nil {
		report.FailureDetail = err.Error()
		return declineCounterexample(report, "REFUTED", "COUNTEREXAMPLE_RULE_NOT_BOUND", "", "")
	}
	if input.Input.ObservedSourceDigest != proposal.Original.SourceDigest ||
		input.Input.ObservedArtifactSourceDigest != proposal.Original.SourceDigest ||
		input.Input.ObservedGeneratedJudgeDigest != DigestBytes(GenerateJudge(proposal.Original)) ||
		input.Input.ObservedIndependentDigest != proposal.Original.SemanticDigest {
		return declineCounterexample(report, "UNKNOWN", "COUNTEREXAMPLE_SNAPSHOT_STALE", "STALE", "OBSERVE_CURRENT_ORIGINAL_POLICY")
	}
	if !sameResult(input.Observed, EvaluateSourcePolicy(proposal.Original, input.Input)) {
		return declineCounterexample(report, "REFUTED", "COUNTEREXAMPLE_RESULT_CONTRADICTS_SOURCE", "", "")
	}
	report.State, report.Reason = "PROPOSED", "SOURCE_BOUND_COUNTEREXAMPLE_PROPOSAL"
	report.Revision, report.OriginalPolicy, report.CandidatePolicy = &revision, &proposal.Original, &proposal.Candidate
	report.CandidateSource = proposal.CandidateSource
	report.ChangedCoordinates = append(report.ChangedCoordinates, proposal.ChangedCoordinates...)
	report.DerivedFields = []string{"condition", "from_decision", "to_decision"}
	return report, nil
}

func declineCounterexample(report PolicyCounterexampleProposal, state, reason, unknownClass, next string) (PolicyCounterexampleProposal, error) {
	report.State, report.Reason = state, reason
	if state == "UNKNOWN" {
		pending := revisionPending("COUNTEREXAMPLE_SELECTION", "DERIVE_REVISION_REQUEST", reason, next)
		pending.UnknownClass = unknownClass
		if unknownClass == "DEPENDENCY_BLOCKED" {
			pending.BlockedBy = []string{"counterexample.observed"}
		}
		report.Pending = &pending
	}
	if report.FailureDetail != "" {
		return report, fmt.Errorf("%s: %w", reason, errors.New(report.FailureDetail))
	}
	return report, fmt.Errorf("%s: %s", state, reason)
}
