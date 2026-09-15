package selfimprovementexecutiongrant

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

const (
	ReasonAllow                 = "EXECUTION_GRANT_GRANTED_UNCONSUMED"
	ReasonDeny                  = "EXECUTION_GRANT_DENIED"
	ReasonReplay                = "EXECUTION_GRANT_REPLAYED_DETERMINISTICALLY"
	ReasonMissingDecision       = "MISSING_EXPLICIT_EXECUTION_GRANT_DECISION"
	ReasonMissingArtifact       = "MISSING_OR_EXPIRED_UPSTREAM_ARTIFACT"
	ReasonSourceRetrievalFailed = "SOURCE_ARTIFACT_RETRIEVAL_FAILED"
	ReasonIncompleteInput       = "INCOMPLETE_EXECUTION_GRANT_INPUT"
	ReasonV24Unknown            = "V24_AUTHORIZATION_UNKNOWN"
	ReasonV24Denied             = "V24_AUTHORIZATION_DENIED"
	ReasonV25Unknown            = "V25_PRE_EXECUTION_CONTRACT_UNKNOWN"
	ReasonV25Refuted            = "V25_PRE_EXECUTION_CONTRACT_REFUTED"
	ReasonDigestMismatch        = "EXECUTION_GRANT_DIGEST_BINDING_MISMATCH"
	ReasonScopeMismatch         = "EXECUTION_GRANT_SCOPE_BINDING_MISMATCH"
	ReasonDuplicate             = "CONFLICTING_DUPLICATE_EXECUTION_GRANTS"
	ReasonUnsafe                = "UNAUTHORIZED_OR_UNSAFE_EXECUTION_GRANT"
	ReasonUpstreamContradiction = "UPSTREAM_AUTHORIZATION_CONTRADICTION"
)

func Evaluate(program PolicyProgram, input GrantInput) GrantResolution {
	resolution := baseResolution(program, input)
	missing, contradictions := inspectInput(input)
	resolution.MissingFields = sortedStrings(missing)
	resolution.ContradictoryFields = sortedStrings(contradictions)
	resolution.Metrics.SixFieldUnknowns = 0
	resolution.Metrics.RefutedContradictions = boolInt(len(contradictions) > 0)

	if len(contradictions) > 0 {
		return finishRefuted(resolution, refutedReason(contradictions))
	}
	if sourceArtifactRetrievalFailed(input.Request.Source) {
		return finishUnknownWithObligations(resolution, ReasonSourceRetrievalFailed, "FETCH", "2", "SOURCE_ARTIFACT_RETRIEVAL", "retry-exact-source-artifact-retrieval", "source_artifact_retrieval", unknownObligations(input, "source_artifact_retrieval"), unknownFrontier(input, "retry-exact-source-artifact-retrieval"))
	}
	if len(input.DecisionInputs) == 0 {
		return finishUnknownWithObligations(resolution, ReasonMissingDecision, "GRANT", "1", "INCOMPLETE_EVIDENCE", "provide-explicit-execution-grant-decision", "explicit_execution_grant_decision", unknownObligations(input, "explicit_execution_grant_decision"), unknownFrontier(input, "provide-explicit-execution-grant-decision"))
	}
	if sourceMissing(input.Request.Source) || input.Request.Source.ArtifactExpired {
		return finishUnknownWithObligations(resolution, ReasonMissingArtifact, "BIND", "3", "UPSTREAM_ARTIFACT_UNAVAILABLE", "restore-exact-upstream-artifact", "upstream_artifact", unknownObligations(input, "upstream_artifact"), unknownFrontier(input, "restore-exact-upstream-artifact"))
	}
	if len(missing) > 0 {
		return finishUnknown(resolution, ReasonIncompleteInput, "BIND", "3", "INCOMPLETE_EVIDENCE", "bind-complete-execution-grant-input", "grant_input_binding")
	}
	if conflictingDecisions(input.DecisionInputs) {
		return finishRefuted(resolution, ReasonDuplicate)
	}
	decision := input.DecisionInputs[0]
	if decision.Decision == DecisionAllow {
		if input.Request.V24.AuthorizationResolution == candidate.AuthorizationUnknown {
			return finishUnknown(resolution, ReasonV24Unknown, "UPSTREAM", "4", "UPSTREAM_UNKNOWN", "resolve-v24-authorization", "v24_authorization_resolution")
		}
		if input.Request.V24.AuthorizationOutcome == candidate.AuthorizationDenied || input.Request.V24.AuthorizationDecision == candidate.AuthorizationDeny {
			return finishRefuted(resolution, ReasonV24Denied)
		}
		if input.Request.V25.Decision == string(v25.DecisionUnknown) {
			return finishUnknown(resolution, ReasonV25Unknown, "UPSTREAM", "5", "UPSTREAM_UNKNOWN", "resolve-v25-pre-execution-contract", "v25_pre_execution_contract")
		}
		if input.Request.V25.Decision == string(v25.DecisionRefuted) {
			return finishRefuted(resolution, ReasonV25Refuted)
		}
		if !v24AllowsGrant(input.Request.V24) || !v25AllowsGrant(input.Request.V25) {
			return finishRefuted(resolution, ReasonUpstreamContradiction)
		}
		return finishClosed(resolution, decision, true, ReasonAllow, ResolutionGrantedUnconsumed)
	}
	if decision.Decision == DecisionDeny {
		return finishClosed(resolution, decision, false, ReasonDeny, ResolutionDenied)
	}
	return finishRefuted(resolution, ReasonUnsafe)
}

func baseResolution(program PolicyProgram, input GrantInput) GrantResolution {
	sourceBound := boolInt(validArtifact(input.Request.Source))
	exactSourceDigest := boolInt(sourceBound == 1 && input.Request.Source.ObservedArtifactDigest != "" && input.Request.Source.ArtifactDigest == input.Request.Source.ObservedArtifactDigest)
	metrics := GrantMetrics{
		StructuralSeparateGrantEdgesBefore: 0, StructuralSeparateGrantEdgesAfter: 1,
		SourceArtifactBoundBefore: 0, SourceArtifactBoundAfter: sourceBound, SourceArtifactBound: sourceBound,
		SourceArtifactExpiredMisclassifiedBefore: 1, SourceArtifactExpiredMisclassifiedAfter: 0, SourceArtifactExpiredMisclassified: 0,
		ExactSourceDigestBoundBefore: 0, ExactSourceDigestBoundAfter: exactSourceDigest, ExactSourceDigestBound: exactSourceDigest,
		LiveGrantRequests: boolInt(input.Live), LiveGrants: 0, LiveExecutionCount: 0,
		LiveGrantsBefore: 0, LiveGrantsAfter: 0, ExecutionCountBefore: 0, ExecutionCountAfter: 0,
		CanonicalExecutionCount: 0, GrantConsumedUses: 0, RepositoryWrites: 0,
		RepositoryWritesBefore: 0, RepositoryWritesAfter: 0,
		LocalTestExecutions: 0, FallbackAccepted: 0, PerformanceImprovement: PerformanceUnknown,
		GoPhysicalLines: program.Inventory.GoPhysicalLines, GoooPhysicalLines: program.Inventory.GoooPhysicalLines,
		CounterexampleArtifactIDs: []int64{KnownFlawedArtifactID},
	}
	return GrantResolution{Schema: ResolutionSchema, RequestDigest: input.Request.Digest,
		GrantAllowsExecution: false, RemainingUses: 0, ConsumedUses: 0, ExecutionCount: 0,
		RepositoryWrites: 0, LocalTestExecutions: 0, FallbackAccepted: 0,
		ConsumptionObligation: ConsumptionObligation, OneUseEnforced: false, Metrics: metrics}
}

func finishUnknown(resolution GrantResolution, reason, stage, step, class, next, blocked string) GrantResolution {
	return finishUnknownWithObligations(resolution, reason, stage, step, class, next, blocked, []string{blocked}, []string{next})
}

func finishUnknownWithObligations(resolution GrantResolution, reason, stage, step, class, next, blocked string, obligations, frontier []string) GrantResolution {
	resolution.Decision, resolution.Resolution, resolution.Reason = DecisionUnknown, ResolutionLower, reason
	resolution.Unknown = &UnknownState{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blocked}
	resolution.Obligations = sortedStrings(obligations)
	resolution.Frontier = sortedStrings(frontier)
	resolution.Metrics.SixFieldUnknowns = 1
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func unknownObligations(input GrantInput, primary string) []string {
	obligations := []string{primary}
	if len(input.DecisionInputs) == 0 {
		obligations = append(obligations, "explicit_execution_grant_decision")
	}
	if sourceArtifactRetrievalFailed(input.Request.Source) {
		obligations = append(obligations, "source_artifact_retrieval")
	}
	if sourceMissing(input.Request.Source) || input.Request.Source.ArtifactExpired {
		obligations = append(obligations, "upstream_artifact")
	}
	if len(requiredMissing(input)) > 0 {
		obligations = append(obligations, "grant_input_binding")
	}
	return uniqueStrings(obligations)
}

func unknownFrontier(input GrantInput, primary string) []string {
	frontier := []string{primary}
	if len(input.DecisionInputs) == 0 {
		frontier = append(frontier, "provide-explicit-execution-grant-decision")
	}
	if sourceArtifactRetrievalFailed(input.Request.Source) {
		frontier = append(frontier, "retry-exact-source-artifact-retrieval")
	}
	if sourceMissing(input.Request.Source) || input.Request.Source.ArtifactExpired {
		frontier = append(frontier, "restore-exact-upstream-artifact")
	}
	if len(requiredMissing(input)) > 0 {
		frontier = append(frontier, "bind-complete-execution-grant-input")
	}
	return uniqueStrings(frontier)
}

func uniqueStrings(values []string) []string {
	result := []string{}
	for _, value := range values {
		if value != "" && !contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

func finishRefuted(resolution GrantResolution, reason string) GrantResolution {
	resolution.Decision, resolution.Resolution, resolution.Reason = DecisionRefuted, ResolutionExact, reason
	resolution.Metrics.RefutedContradictions = 1
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func finishClosed(resolution GrantResolution, input GrantDecisionInput, allows bool, reason string, outcome Resolution) GrantResolution {
	resolution.Decision, resolution.Resolution, resolution.Reason = DecisionClosed, outcome, reason
	resolution.DecisionInputs = []GrantDecisionInput{input}
	resolution.GrantAllowsExecution = allows
	resolution.ExecutionCount, resolution.ConsumedUses = 0, 0
	if allows {
		resolution.RemainingUses = 1
		resolution.Metrics.LiveGrants = boolInt(resolution.Metrics.LiveGrantRequests == 1)
		resolution.Metrics.LiveGrantsAfter = resolution.Metrics.LiveGrants
		resolution.Metrics.GrantRemainingUses = 1
	}
	resolution.Receipt = buildReceipt(resolution.RequestDigest, input, allows)
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func buildReceipt(requestDigest string, input GrantDecisionInput, allows bool) *GrantReceipt {
	if input.DecisionDigest == "" {
		input.DecisionDigest = decisionDigest(input)
	}
	receipt := GrantReceipt{Schema: ReceiptSchema,
		GrantID:       "execution-grant/" + strings.TrimPrefix(input.DecisionDigest, "sha256:"),
		RequestDigest: requestDigest, Decision: input.Decision, DecisionSource: input.DecisionSource,
		ActorEvidence: input.ActorEvidence, GrantAllowsExecution: allows, RemainingUses: boolInt(allows),
		ConsumedUses: 0, ExecutionCount: 0, ConsumptionStatus: ConsumptionPending,
		ConsumptionObligation: ConsumptionObligation, OneUseEnforcementState: "PENDING_NEXT_EXECUTOR"}
	receipt.Digest = receiptDigest(receipt)
	return &receipt
}

func inspectInput(input GrantInput) ([]string, []string) {
	missing := requiredMissing(input)
	contradictions := []string{}
	addContradiction := func(field string) {
		if slices.Contains(contradictions, field) {
			return
		}
		contradictions = append(contradictions, field)
	}
	if input.Request.Schema != "" && input.Request.Schema != RequestSchema {
		addContradiction("request_schema")
	}
	if input.Request.ContractID != "" && input.Request.ContractID != ContractID {
		addContradiction("grant_contract_id")
	}
	if input.Request.Digest != "" && (!validDigest(input.Request.Digest) || input.Request.Digest != requestDigest(input.Request)) {
		addContradiction("grant_request_digest")
	}
	if input.Request.Target != "" && input.Request.Target != GrantTarget {
		addContradiction("scope")
	}
	if input.Request.Mode != "" && input.Request.Mode != GrantMode {
		addContradiction("scope")
	}
	inspectV24(input.Request.V24, addContradiction)
	inspectV25(input.Request.V25, addContradiction)
	inspectSource(input.Request.Source, addContradiction)
	if conflictingDecisions(input.DecisionInputs) {
		addContradiction("conflicting_duplicate_grant")
	}
	for _, decision := range input.DecisionInputs {
		inspectDecision(input.Request, decision, addContradiction)
	}
	return missing, contradictions
}

func inspectV24(binding V24Binding, add func(string)) {
	for name, value := range map[string]string{"v24_request_digest": binding.RequestDigest, "v24_resolution_digest": binding.ResolutionDigest, "candidate_stable_id": binding.CandidateStableID, "candidate_digest": binding.CandidateDigest, "subject_sha": binding.SubjectSHA, "observation_digest": binding.ObservationDigest, "v24_contract_digest": binding.ContractDigest} {
		if value != "" && !validDigest(value) && name != "subject_sha" {
			add(name)
		}
	}
	if binding.SubjectSHA != "" && !validSHA(binding.SubjectSHA) {
		add("subject_sha")
	}
	if binding.RequestValid && binding.ResolutionValid && (binding.AuthorizationDecision != candidate.AuthorizationAllow && binding.AuthorizationDecision != candidate.AuthorizationDeny) {
		add("v24_authorization_decision")
	}
}

func inspectV25(binding V25Binding, add func(string)) {
	for name, value := range map[string]string{"v25_contract_digest": binding.ContractDigest, "candidate_stable_id": binding.CandidateStableID, "candidate_digest": binding.CandidateDigest, "subject_sha": binding.SubjectSHA, "observation_digest": binding.ObservationDigest, "candidate_input_digest": binding.CandidateInputDigest, "evaluator_registry_digest": binding.EvaluatorRegistryDigest, "toolchain_test_contract_identity": binding.ToolchainTestContractIdentity} {
		if value != "" && name == "subject_sha" && !validSHA(value) {
			add(name)
		} else if value != "" && name != "subject_sha" && !validDigest(value) {
			add(name)
		}
	}
	if binding.MaxExecutions < 0 || binding.MaxExecutions > MaxExecutions {
		add("max_executions")
	}
	if binding.RepositoryWritesAllowed || binding.ExecutionAuthorized {
		add("execution_safety")
	}
	if binding.MaxExecutions != 0 && binding.MaxExecutions != MaxExecutions {
		add("max_executions")
	}
	if binding.BoundedTarget != "" && binding.BoundedTarget != string(v25.KnownBoundedTarget) {
		add("scope")
	}
	if binding.Decision != "" && binding.Decision != string(v25.DecisionClosed) && binding.Decision != string(v25.DecisionUnknown) && binding.Decision != string(v25.DecisionRefuted) {
		add("v25_decision")
	}
}

func inspectSource(source SourceArtifact, add func(string)) {
	if source.WorkflowRunID < 0 || source.WorkflowRunAttempt < 0 || source.ArtifactID < 0 {
		add("source_artifact_identity")
	}
	if source.ArtifactDigest != "" && !validDigest(source.ArtifactDigest) {
		add("source_artifact_digest")
	}
	if source.ObservedArtifactDigest != "" && source.ArtifactDigest != source.ObservedArtifactDigest {
		add("source_artifact_digest")
	}
}

func inspectDecision(request GrantRequest, input GrantDecisionInput, add func(string)) {
	if input.Schema != "" && input.Schema != GrantDecisionSchema {
		add("grant_decision_schema")
	}
	if input.Decision != "" && input.Decision != DecisionAllow && input.Decision != DecisionDeny {
		add("grant_decision")
	}
	if input.RequestDigest != "" && (!validDigest(input.RequestDigest) || input.RequestDigest != request.Digest) {
		add("grant_request_digest")
	}
	if input.DecisionDigest != "" && (!validDigest(input.DecisionDigest) || input.DecisionDigest != decisionDigest(input)) {
		add("grant_decision_digest")
	}
	if input.V24 != request.V24 {
		add("v24_binding")
	}
	if input.V25 != request.V25 {
		add("v25_binding")
	}
	if input.Source != request.Source {
		add("source_artifact_binding")
	}
	if input.DecisionSource != DecisionSourceWorkflowDispatch && input.DecisionSource != DecisionSourceCanonical {
		add("unauthorized_grant")
	}
	if input.DecisionSource == DecisionSourceWorkflowDispatch && (input.ActorEvidence.EvidenceLabel != ActorEvidenceLabel || input.ActorEvidence.Event != DecisionSourceWorkflowDispatch || input.ActorEvidence.Repository == "" || input.ActorEvidence.Actor == "" || input.ActorEvidence.WorkflowRunID <= 0 || input.ActorEvidence.WorkflowRunAttempt <= 0) {
		add("unauthorized_grant")
	}
	if input.DecisionSource == DecisionSourceCanonical && input.ActorEvidence.EvidenceLabel != CanonicalEvidenceLabel {
		add("unauthorized_grant")
	}
}

func requiredMissing(input GrantInput) []string {
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

func v24Ready(binding V24Binding) bool {
	return binding.RequestSchema == candidate.AuthorizationRequestSchema && binding.ResolutionSchema == candidate.AuthorizationResolutionSchema && validDigest(binding.RequestDigest) && validDigest(binding.ResolutionDigest) && validDigest(binding.CandidateStableID) && validDigest(binding.CandidateDigest) && validSHA(binding.SubjectSHA) && validDigest(binding.ObservationDigest) && validDigest(binding.ContractDigest) && binding.RequestValid && binding.ResolutionValid
}

func v25Ready(binding V25Binding) bool {
	return binding.Schema == v25.Schema && binding.ContractID == v25.ContractID && validDigest(binding.ContractDigest) && validDigest(binding.CandidateStableID) && validDigest(binding.CandidateDigest) && validSHA(binding.SubjectSHA) && validDigest(binding.ObservationDigest) && validDigest(binding.CandidateInputDigest) && binding.OperationID == v25.KnownOperationID && binding.BoundedTarget == string(v25.KnownBoundedTarget) && validDigest(binding.EvaluatorRegistryDigest) && validDigest(binding.ToolchainTestContractIdentity) && binding.Valid
}

func v24AllowsGrant(binding V24Binding) bool {
	return v24Ready(binding) && binding.AuthorizationDecision == candidate.AuthorizationAllow && binding.AuthorizationResolution == candidate.AuthorizationClosed && binding.AuthorizationOutcome == candidate.AuthorizationAuthorized
}

func v25AllowsGrant(binding V25Binding) bool {
	return v25Ready(binding) && binding.Decision == string(v25.DecisionClosed) && binding.Resolution == string(v25.ResolutionDeclared) && !binding.ExecutionAuthorized && binding.ExecutionGrantRequired && binding.MaxExecutions == MaxExecutions && !binding.RepositoryWritesAllowed
}

func conflictingDecisions(inputs []GrantDecisionInput) bool {
	if len(inputs) < 2 {
		return false
	}
	first := inputs[0]
	first.DecisionDigest = ""
	for _, current := range inputs[1:] {
		candidate := current
		candidate.DecisionDigest = ""
		if !reflect.DeepEqual(first, candidate) {
			return true
		}
	}
	return false
}

func refutedReason(fields []string) string {
	if contains(fields, "conflicting_duplicate_grant") {
		return ReasonDuplicate
	}
	if contains(fields, "execution_safety") || contains(fields, "max_executions") {
		return ReasonUnsafe
	}
	if contains(fields, "unauthorized_grant") {
		return ReasonUnsafe
	}
	if contains(fields, "scope") {
		return ReasonScopeMismatch
	}
	if contains(fields, "grant_request_digest") || contains(fields, "grant_decision_digest") || contains(fields, "v24_binding") || contains(fields, "v25_binding") || contains(fields, "source_artifact_binding") || contains(fields, "candidate_digest") {
		return ReasonDigestMismatch
	}
	return ReasonDigestMismatch
}

func sourceMissing(source SourceArtifact) bool {
	return !validArtifact(source)
}

func sourceArtifactRetrievalFailed(source SourceArtifact) bool {
	return source.ArtifactRetrievalError != ""
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ValidateResolution(resolution GrantResolution) error {
	if resolution.Schema != ResolutionSchema || !validDigest(resolution.RequestDigest) || resolution.Digest != resolutionDigest(resolution) {
		return errors.New("execution grant resolution identity mismatch")
	}
	if resolution.GrantAllowsExecution && (resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionGrantedUnconsumed || resolution.RemainingUses != 1 || resolution.ConsumedUses != 0 || resolution.ExecutionCount != 0) {
		return errors.New("execution grant crossed its consumption boundary")
	}
	if resolution.ExecutionCount != 0 || resolution.ConsumedUses != 0 || resolution.RepositoryWrites != 0 || resolution.LocalTestExecutions != 0 || resolution.FallbackAccepted != 0 || resolution.OneUseEnforced {
		return errors.New("execution grant performed forbidden work")
	}
	switch resolution.Decision {
	case DecisionUnknown:
		if resolution.Resolution != ResolutionLower || resolution.Unknown == nil || resolution.Receipt != nil || len(resolution.DecisionInputs) != 0 || resolution.Unknown.Stage == "" || resolution.Unknown.Step == "" || resolution.Unknown.Reason == "" || resolution.Unknown.UnknownClass == "" || resolution.Unknown.NextOperation == "" || resolution.Unknown.BlockedBy == "" || len(resolution.Obligations) == 0 || len(resolution.Frontier) == 0 || !contains(resolution.Obligations, resolution.Unknown.BlockedBy) || !contains(resolution.Frontier, resolution.Unknown.NextOperation) {
			return errors.New("execution grant UNKNOWN resolution is not six-field causal")
		}
	case DecisionRefuted:
		if resolution.Resolution != ResolutionExact || resolution.Unknown != nil || resolution.Receipt != nil || resolution.Metrics.RefutedContradictions != 1 {
			return errors.New("execution grant REFUTED resolution is not exact")
		}
	case DecisionClosed:
		if len(resolution.DecisionInputs) != 1 || resolution.Receipt == nil || (resolution.Resolution != ResolutionDenied && resolution.Resolution != ResolutionGrantedUnconsumed) || resolution.Receipt.Digest != receiptDigest(*resolution.Receipt) {
			return errors.New("execution grant CLOSED resolution is not exact")
		}
	default:
		return errors.New("execution grant resolution decision is unknown")
	}
	return nil
}

func VerifyGrantResolution(resolution GrantResolution) error {
	if err := ValidateResolution(resolution); err != nil {
		return fmt.Errorf("grant resolution verification failed: %w", err)
	}
	return nil
}
