package selfimprovementcontinuation

import "reflect"

func Verify(program PolicyProgram, request ContinuationRequest, resolution ContinuationResolution) Verification {
	decision, outcome, reason := independentClassify(program, request.Input)
	verification := Verification{
		Schema: VerificationSchema, RequestDigest: request.Digest, ResolutionDigest: resolution.Digest,
		IndependentDecision: decision, IndependentResolution: outcome, IndependentReason: reason,
		IndependentReplayComparisons: 1, ExecutionAuthorized: false, LiveGrantDecision: 0,
		LiveExecutionCount: 0, GrantConsumedUses: 0, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	verification.Verified = request.Digest == requestDigest(request) && resolution.RequestDigest == request.Digest && reflect.DeepEqual(request.Input, resolution.Input) && resolution.Decision == decision && resolution.Resolution == outcome && resolution.Reason == reason && !resolution.ExecutionAuthorized && resolution.ExecutionGrants == 0 && resolution.LiveGrantDecision == 0 && resolution.LiveExecutionCount == 0 && resolution.GrantConsumedUses == 0 && resolution.RepositoryWrites == 0 && resolution.LocalTestExecutions == 0 && ValidateResolution(resolution) == nil
	verification.Digest = verificationDigest(verification)
	return verification
}

func independentClassify(program PolicyProgram, input ContinuationInput) (Decision, Resolution, string) {
	if len(program.Policy.Cases) != 9 || program.Evidence.Schema != PolicySchema || program.Evidence.PolicyID != ContractID {
		return DecisionUnknown, ResolutionLower, ReasonMissingIdentity
	}
	missing, contradictions := independentFields(input)
	if len(contradictions) > 0 {
		return DecisionRefuted, ResolutionExact, refutedReason(contradictions)
	}
	if len(missing) > 0 {
		reason := unknownReason(input, missing)
		return DecisionUnknown, ResolutionLower, reason
	}
	if input.Replay {
		return DecisionClosed, ResolutionExact, ReasonReplay
	}
	if input.DuplicateDispatch {
		return DecisionClosed, ResolutionExact, ReasonIdempotent
	}
	return DecisionClosed, ResolutionExact, ReasonExact
}

func independentFields(input ContinuationInput) ([]string, []string) {
	missing, contradictions := requiredMissing(input), []string{}
	for _, field := range input.ParseErrors {
		if !contains(contradictions, field) {
			contradictions = append(contradictions, field)
		}
	}
	add := func(value string) {
		if !contains(contradictions, value) {
			contradictions = append(contradictions, value)
		}
	}
	expectedName, expectedPath, expectedEvent := expectedSource(input)
	if input.SourceWorkflowName != "" && input.SourceWorkflowName != expectedName {
		add("source_workflow_name")
	}
	if input.SourceWorkflowPath != "" && input.SourceWorkflowPath != expectedPath {
		add("source_workflow_path")
	}
	if input.SourceRepository != "" && input.SourceRepository != SourceRepository {
		add("source_repository")
	}
	if input.SourceEvent != "" && input.SourceEvent != expectedEvent {
		add("source_event")
	}
	if input.SourceRef != "" && input.SourceRef != SourceRef {
		add("source_ref")
	}
	if input.DispatchRef != "" && input.DispatchRef != DispatchRef {
		add("dispatch_ref")
	}
	if input.DispatchMode != "" && input.DispatchMode != DispatchMode {
		add("dispatch_mode")
	}
	if input.TargetWorkflowName != "" && input.TargetWorkflowName != V25WorkflowName && input.TargetWorkflowName != V26WorkflowName {
		add("target_workflow_name")
	}
	if input.TargetWorkflowPath != "" && input.TargetWorkflowPath != V25WorkflowPath && input.TargetWorkflowPath != V26WorkflowPath {
		add("target_workflow_path")
	}
	if input.TargetWorkflowName != "" && input.TargetWorkflowPath != "" && !targetPairMatches(input.TargetWorkflowName, input.TargetWorkflowPath) {
		add("target_workflow_identity")
	}
	if input.SourceHeadSHA != "" && !validSHA(input.SourceHeadSHA) {
		add("source_head_sha")
	}
	if input.SourceArtifactArchiveDigest != "" && !validDigest(input.SourceArtifactArchiveDigest) {
		add("source_artifact_archive_digest")
	}
	if input.SourceArtifactObservedDigest != "" && !validDigest(input.SourceArtifactObservedDigest) {
		add("source_artifact_observed_digest")
	}
	if input.SourceReceiptDigest != "" && !validDigest(input.SourceReceiptDigest) {
		add("source_receipt_digest")
	}
	if input.SourceArtifactArchiveDigest != "" && input.SourceArtifactObservedDigest != "" && input.SourceArtifactArchiveDigest != input.SourceArtifactObservedDigest {
		add("source_artifact_digest")
	}
	if input.DuplicateConflict {
		add("conflicting_duplicate_dispatch")
	}
	if input.ManualDispatches != 0 || input.UnauthorizedDispatches != 0 || input.ExecutionAuthorized || input.ExecutionGrants != 0 || input.LiveGrantDecision != 0 || input.LiveExecutionCount != 0 || input.GrantConsumedUses != 0 || input.RepositoryWrites != 0 || input.LocalTestExecutions != 0 {
		add("unauthorized_dispatch")
	}
	if input.SourceArtifactName != "" && !artifactNameMatches(input) {
		add("source_artifact_name")
	}
	return missing, contradictions
}
