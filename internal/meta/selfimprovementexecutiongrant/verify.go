package selfimprovementexecutiongrant

import (
	"errors"
	"reflect"

	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

// Verify is an independent consumer. It reclassifies the typed input without
// calling Evaluate and checks the emitted receipt's safety boundary.
func Verify(program PolicyProgram, input GrantInput, resolution GrantResolution) Verification {
	decision, outcome, reason := independentClassify(program, input)
	verification := Verification{Schema: VerificationSchema, RequestDigest: input.Request.Digest,
		ResolutionDigest: resolution.Digest, IndependentDecision: decision, IndependentResolution: outcome,
		IndependentReason: reason, IndependentReplayComparisons: 1, LiveGrants: 0,
		ExecutionCount: 0, GrantConsumedUses: 0, RepositoryWrites: 0, LocalTestExecutions: 0}
	verification.Verified = resolution.Decision == decision && resolution.Resolution == outcome && resolution.Reason == reason && resolution.RequestDigest == input.Request.Digest && !resolution.OneUseEnforced && resolution.ExecutionCount == 0 && resolution.ConsumedUses == 0 && resolution.RepositoryWrites == 0 && resolution.LocalTestExecutions == 0 && ValidateResolution(resolution) == nil
	if resolution.GrantAllowsExecution {
		verification.LiveGrants = boolInt(input.Live)
	}
	verification.Digest = verificationDigest(verification)
	return verification
}

func independentClassify(program PolicyProgram, input GrantInput) (Decision, Resolution, string) {
	if !independentPolicy(program) {
		return DecisionUnknown, ResolutionLower, ReasonIncompleteInput
	}
	if fields := independentContradictions(input); len(fields) > 0 {
		return DecisionRefuted, ResolutionExact, independentRefutedReason(fields)
	}
	if input.Request.Source.ArtifactRetrievalError != "" {
		return DecisionUnknown, ResolutionLower, ReasonSourceRetrievalFailed
	}
	if len(input.DecisionInputs) == 0 {
		return DecisionUnknown, ResolutionLower, ReasonMissingDecision
	}
	if independentSourceMissing(input.Request.Source) {
		return DecisionUnknown, ResolutionLower, ReasonMissingArtifact
	}
	if len(independentMissing(input)) > 0 {
		return DecisionUnknown, ResolutionLower, ReasonIncompleteInput
	}
	if independentConflict(input.DecisionInputs) {
		return DecisionRefuted, ResolutionExact, ReasonDuplicate
	}
	decision := input.DecisionInputs[0]
	if decision.Decision == DecisionAllow {
		if input.Request.V24.AuthorizationResolution == candidate.AuthorizationUnknown {
			return DecisionUnknown, ResolutionLower, ReasonV24Unknown
		}
		if input.Request.V24.AuthorizationDecision == candidate.AuthorizationDeny || input.Request.V24.AuthorizationOutcome == candidate.AuthorizationDenied {
			return DecisionRefuted, ResolutionExact, ReasonV24Denied
		}
		if input.Request.V25.Decision == string(v25.DecisionUnknown) {
			return DecisionUnknown, ResolutionLower, ReasonV25Unknown
		}
		if input.Request.V25.Decision == string(v25.DecisionRefuted) {
			return DecisionRefuted, ResolutionExact, ReasonV25Refuted
		}
		if !independentV24Allows(input.Request.V24) || !independentV25Allows(input.Request.V25) {
			return DecisionRefuted, ResolutionExact, ReasonUpstreamContradiction
		}
		return DecisionClosed, ResolutionGrantedUnconsumed, ReasonAllow
	}
	if decision.Decision == DecisionDeny {
		return DecisionClosed, ResolutionDenied, ReasonDeny
	}
	return DecisionRefuted, ResolutionExact, ReasonUnsafe
}

func independentPolicy(program PolicyProgram) bool {
	_, _, err := canonicalExecutorProgramBinding(program)
	return err == nil && program.Evidence.Schema == PolicySchema && program.Evidence.PolicyID == ContractID && program.Evidence.CaseCount == 9 && program.Evidence.ClosedCases == 3 && program.Evidence.UnknownCases == 3 && program.Evidence.RefutedCases == 3
}

func independentSourceMissing(source SourceArtifact) bool {
	return !validArtifact(source)
}

func independentMissing(input GrantInput) []string {
	missing := []string{}
	add := func(name string) {
		if !contains(missing, name) {
			missing = append(missing, name)
		}
	}
	v24 := input.Request.V24
	if v24.RequestDigest == "" {
		add("v24_request_digest")
	}
	if v24.ResolutionDigest == "" {
		add("v24_resolution_digest")
	}
	if v24.CandidateStableID == "" {
		add("candidate_stable_id")
	}
	if v24.CandidateDigest == "" {
		add("candidate_digest")
	}
	if v24.SubjectSHA == "" {
		add("subject_sha")
	}
	if v24.ObservationDigest == "" {
		add("observation_digest")
	}
	if v24.ContractDigest == "" {
		add("v24_contract_digest")
	}
	v25 := input.Request.V25
	if v25.ContractDigest == "" {
		add("v25_contract_digest")
	}
	if v25.OperationID == "" {
		add("operation_id")
	}
	if v25.BoundedTarget == "" {
		add("scope")
	}
	if v25.EvaluatorRegistryDigest == "" {
		add("evaluator_registry_digest")
	}
	if v25.ToolchainTestContractIdentity == "" {
		add("toolchain_test_contract_identity")
	}
	if v25.MaxExecutions == 0 {
		add("max_executions")
	}
	source := input.Request.Source
	if source.Repository == "" {
		add("source_repository")
	}
	if source.WorkflowRunID == 0 {
		add("source_workflow_run_id")
	}
	if source.WorkflowRunAttempt == 0 {
		add("source_workflow_run_attempt")
	}
	if source.ArtifactID == 0 {
		add("source_artifact_id")
	}
	if source.ArtifactDigest == "" {
		add("source_artifact_digest")
	}
	if !source.ArtifactExpiryKnown {
		add("source_artifact_expiry")
	}
	return missing
}

func independentContradictions(input GrantInput) []string {
	fields := []string{}
	add := func(name string) {
		if !contains(fields, name) {
			fields = append(fields, name)
		}
	}
	if input.Request.Schema != RequestSchema || input.Request.ContractID != ContractID || !validDigest(input.Request.Digest) || input.Request.Digest != requestDigest(input.Request) {
		add("grant_request_digest")
	}
	if input.Request.Target != GrantTarget || input.Request.Mode != GrantMode {
		add("scope")
	}
	contract := input.Request.V25
	if contract.RepositoryWritesAllowed || contract.ExecutionAuthorized || contract.MaxExecutions > MaxExecutions || contract.MaxExecutions < 0 {
		add("execution_safety")
	}
	if contract.BoundedTarget != "" && contract.BoundedTarget != string(v25.KnownBoundedTarget) {
		add("scope")
	}
	if input.Request.Source.ArtifactDigest != "" && !validDigest(input.Request.Source.ArtifactDigest) {
		add("source_artifact_digest")
	}
	if input.Request.Source.ObservedArtifactDigest != "" && input.Request.Source.ArtifactDigest != input.Request.Source.ObservedArtifactDigest {
		add("source_artifact_digest")
	}
	if independentConflict(input.DecisionInputs) {
		add("conflicting_duplicate_grant")
	}
	for _, decision := range input.DecisionInputs {
		if decision.Schema != GrantDecisionSchema || (decision.Decision != DecisionAllow && decision.Decision != DecisionDeny) || decision.RequestDigest != input.Request.Digest || decision.V24 != input.Request.V24 || decision.V25 != input.Request.V25 || decision.Source != input.Request.Source || !independentActorValid(decision, input.Live) {
			add("unauthorized_grant")
		}
		if decision.DecisionDigest != "" && decision.DecisionDigest != decisionDigest(decision) {
			add("grant_decision_digest")
		}
	}
	return fields
}

func independentActorValid(input GrantDecisionInput, live bool) bool {
	if input.DecisionSource == DecisionSourceCanonical {
		return !live && input.ActorEvidence.EvidenceLabel == CanonicalEvidenceLabel
	}
	return input.DecisionSource == DecisionSourceWorkflowDispatch && input.ActorEvidence.EvidenceLabel == ActorEvidenceLabel && input.ActorEvidence.Event == DecisionSourceWorkflowDispatch && input.ActorEvidence.Repository != "" && input.ActorEvidence.Actor != "" && input.ActorEvidence.WorkflowRunID > 0 && input.ActorEvidence.WorkflowRunAttempt > 0
}

func independentV24Allows(binding V24Binding) bool {
	return binding.RequestValid && binding.ResolutionValid && binding.RequestSchema == candidate.AuthorizationRequestSchema && binding.ResolutionSchema == candidate.AuthorizationResolutionSchema && binding.AuthorizationDecision == candidate.AuthorizationAllow && binding.AuthorizationResolution == candidate.AuthorizationClosed && binding.AuthorizationOutcome == candidate.AuthorizationAuthorized
}

func independentV25Allows(binding V25Binding) bool {
	return binding.Valid && binding.Schema == v25.Schema && binding.ContractID == v25.ContractID && binding.Decision == string(v25.DecisionClosed) && binding.Resolution == string(v25.ResolutionDeclared) && binding.OperationID == v25.KnownOperationID && binding.BoundedTarget == string(v25.KnownBoundedTarget) && binding.MaxExecutions == MaxExecutions && !binding.RepositoryWritesAllowed && !binding.ExecutionAuthorized && binding.ExecutionGrantRequired
}

func independentConflict(inputs []GrantDecisionInput) bool {
	if len(inputs) < 2 {
		return false
	}
	first := inputs[0]
	first.DecisionDigest = ""
	for _, current := range inputs[1:] {
		copy := current
		copy.DecisionDigest = ""
		if !reflect.DeepEqual(first, copy) {
			return true
		}
	}
	return false
}

func independentRefutedReason(fields []string) string {
	if contains(fields, "conflicting_duplicate_grant") {
		return ReasonDuplicate
	}
	if contains(fields, "execution_safety") {
		return ReasonUnsafe
	}
	if contains(fields, "unauthorized_grant") {
		return ReasonUnsafe
	}
	if contains(fields, "scope") {
		return ReasonScopeMismatch
	}
	return ReasonDigestMismatch
}

func ValidateGrantReceipt(receipt GrantReceipt) error {
	if receipt.Schema != ReceiptSchema || receipt.GrantID == "" || !validDigest(receipt.RequestDigest) || receipt.ExecutionCount != 0 || receipt.ConsumedUses != 0 || receipt.ConsumptionObligation != ConsumptionObligation || receipt.OneUseEnforcementState != "PENDING_NEXT_EXECUTOR" || receipt.Digest != receiptDigest(receipt) {
		return errors.New("execution grant receipt is not an unconsumed declaration")
	}
	return nil
}
