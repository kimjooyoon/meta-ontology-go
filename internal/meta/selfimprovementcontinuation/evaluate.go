package selfimprovementcontinuation

import "errors"

const (
	ReasonExact             = "EXACT_CONTINUATION_IDENTITY"
	ReasonReplay            = "CONTINUATION_REQUEST_REPLAYED"
	ReasonIdempotent        = "IDEMPOTENT_DUPLICATE_CONTINUATION"
	ReasonMissingIdentity   = "MISSING_DISPATCH_IDENTITY"
	ReasonMissingReceipt    = "MISSING_CONTINUATION_ARTIFACT_RECEIPT"
	ReasonMissingLineage    = "MISSING_SOURCE_HEAD_LINEAGE"
	ReasonWorkflowMismatch  = "CONTINUATION_WORKFLOW_IDENTITY_MISMATCH"
	ReasonDigestMismatch    = "CONTINUATION_ARTIFACT_DIGEST_MISMATCH"
	ReasonUnauthorized      = "UNAUTHORIZED_CONTINUATION_DISPATCH"
	ReasonDuplicateConflict = "CONFLICTING_DUPLICATE_CONTINUATIONS"
	ReasonMalformedIdentity = "MALFORMED_CONTINUATION_IDENTITY"
)

func RequiredFieldNames() []string {
	return []string{
		"source_workflow_name", "source_workflow_path", "source_repository", "source_event",
		"source_ref", "source_head_sha", "source_run_id", "source_run_attempt", "source_artifact_name",
		"source_artifact_id", "source_artifact_archive_digest", "source_artifact_observed_digest",
		"source_receipt_digest", "target_workflow_name", "target_workflow_path", "dispatch_ref", "dispatch_mode",
	}
}

func Evaluate(program PolicyProgram, input ContinuationInput) ContinuationResolution {
	resolution := baseResolution(program, input)
	missing, contradictions := inspectInput(input)
	resolution.MissingFields = sortedStrings(missing)
	resolution.ContradictoryFields = sortedStrings(contradictions)
	if len(contradictions) > 0 {
		resolution.Decision, resolution.Resolution, resolution.Reason = DecisionRefuted, ResolutionExact, refutedReason(contradictions)
		resolution.Digest = resolutionDigest(resolution)
		return resolution
	}
	if len(missing) > 0 {
		resolution.Decision, resolution.Resolution = DecisionUnknown, ResolutionLower
		resolution.Reason = unknownReason(input, missing)
		resolution.Unknown = unknownFor(resolution.Reason)
		resolution.Digest = resolutionDigest(resolution)
		return resolution
	}
	if input.Replay {
		resolution.Decision, resolution.Resolution, resolution.Reason = DecisionClosed, ResolutionExact, ReasonReplay
	} else if input.DuplicateDispatch {
		resolution.Decision, resolution.Resolution, resolution.Reason = DecisionClosed, ResolutionExact, ReasonIdempotent
	} else {
		resolution.Decision, resolution.Resolution, resolution.Reason = DecisionClosed, ResolutionExact, ReasonExact
	}
	resolution.Metrics.ExactIdentityBindingsAfter = IdentityBindingsAfter
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func baseResolution(program PolicyProgram, input ContinuationInput) ContinuationResolution {
	return ContinuationResolution{
		Schema: ResolutionSchema, Input: input, Decision: DecisionUnknown, Resolution: ResolutionLower,
		RequestDigest:       requestDigest(ContinuationRequest{Schema: RequestSchema, Lifecycle: "REQUESTED", ContractID: ContractID, Input: input, Decision: Decision("REQUESTED"), Resolution: Resolution("UNRESOLVED"), Reason: "CI_CONTINUATION_REQUESTED", ExecutionAuthorized: false, RepositoryWrites: 0, LocalTestExecutions: 0}),
		ExecutionAuthorized: false, ExecutionGrants: 0, LiveGrantDecision: 0, LiveExecutionCount: 0,
		GrantConsumedUses: 0, RepositoryWrites: 0, LocalTestExecutions: 0,
		Metrics: Metrics{WorkflowRunContinuationEdgesBefore: DepthEdgesBefore, WorkflowRunContinuationEdgesAfter: DepthEdgesAfter,
			ExactIdentityBindingsBefore: IdentityBindingsBefore, ExactIdentityBindingsAfter: 0,
			ManualDispatchesBefore: 0, ManualDispatchesAfter: 0, LiveGrantDecision: 0, LiveGrants: 0,
			LiveExecutionCount: 0, GrantConsumedUses: 0, RepositoryWrites: 0, LocalTestExecutions: 0,
			CanonicalExecutionCount: 0, CanonicalCases: 0, SixFieldUnknowns: 0,
			UnauthorizedDispatches: input.UnauthorizedDispatches, FallbackAccepted: 0,
			IndependentReplayComparisons: 1, ArtifactFiles: ArtifactFiles, ArtifactTypes: ArtifactTypes,
			GoPhysicalLines: program.GoPhysicalLines, GoooPhysicalLines: program.GoooPhysicalLines,
			PerformanceImprovement: PerformanceUnknown, CounterexampleRunID: CounterexampleRunID},
	}
}

func inspectInput(input ContinuationInput) ([]string, []string) {
	missing := requiredMissing(input)
	contradictions := sortedStrings(append([]string(nil), input.ParseErrors...))
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
	if input.TargetWorkflowPath != "" && input.DispatchRef != "" && input.TargetWorkflowPath == V26WorkflowPath && input.DispatchMode != DispatchMode {
		add("dispatch_mode")
	}
	return missing, contradictions
}

func requiredMissing(input ContinuationInput) []string {
	missing := []string{}
	add := func(value string) {
		if !contains(missing, value) {
			missing = append(missing, value)
		}
	}
	if input.SourceWorkflowName == "" || input.SourceWorkflowPath == "" || input.SourceRepository == "" || input.SourceEvent == "" {
		add("source_workflow_identity")
	}
	if input.SourceRef == "" || input.DispatchRef == "" {
		add("source_ref")
	}
	if input.SourceHeadSHA == "" {
		add("source_head_sha")
	}
	if input.SourceRunID <= 0 {
		add("source_run_id")
	}
	if input.SourceRunAttempt <= 0 {
		add("source_run_attempt")
	}
	if input.SourceArtifactName == "" || input.SourceArtifactID <= 0 {
		add("source_artifact_identity")
	}
	if input.SourceArtifactArchiveDigest == "" || input.SourceArtifactObservedDigest == "" || input.SourceReceiptDigest == "" {
		add("source_artifact_receipt")
	}
	if input.TargetWorkflowName == "" || input.TargetWorkflowPath == "" {
		add("target_workflow_identity")
	}
	if input.DispatchMode == "" {
		add("dispatch_mode")
	}
	return missing
}

func artifactNameMatches(input ContinuationInput) bool {
	if input.SourceHeadSHA == "" {
		return true
	}
	if input.TargetWorkflowPath == V25WorkflowPath {
		return input.SourceArtifactName == "self-improvement-candidate-authorization-"+input.SourceHeadSHA
	}
	if input.TargetWorkflowPath == V26WorkflowPath {
		return input.SourceArtifactName == "self-improvement-execution-contract-"+input.SourceHeadSHA
	}
	return false
}

func expectedSource(input ContinuationInput) (string, string, string) {
	if input.TargetWorkflowPath == V26WorkflowPath {
		return V25WorkflowName, V25WorkflowPath, DispatchMode
	}
	return SourceWorkflowName, SourceWorkflowPath, SourceEvent
}

func targetPairMatches(name, path string) bool {
	return (name == V25WorkflowName && path == V25WorkflowPath) || (name == V26WorkflowName && path == V26WorkflowPath)
}

func unknownReason(input ContinuationInput, missing []string) string {
	if contains(missing, "source_artifact_receipt") || contains(missing, "source_artifact_identity") {
		return ReasonMissingReceipt
	}
	if contains(missing, "source_head_sha") || contains(missing, "source_ref") {
		return ReasonMissingLineage
	}
	return ReasonMissingIdentity
}

func unknownFor(reason string) *UnknownState {
	unknown := &UnknownState{Stage: "DISPATCH", Step: "1", Reason: reason, UnknownClass: "INCOMPLETE_EVIDENCE", NextOperation: "bind-exact-continuation-identity", BlockedBy: []string{"continuation_identity"}}
	switch reason {
	case ReasonMissingReceipt:
		unknown.Stage, unknown.Step, unknown.NextOperation, unknown.BlockedBy = "FETCH", "2", "restore-exact-continuation-artifact-receipt", []string{"continuation_artifact_receipt"}
	case ReasonMissingLineage:
		unknown.Stage, unknown.Step, unknown.NextOperation, unknown.BlockedBy = "LINEAGE", "3", "bind-exact-source-head-lineage", []string{"source_head_lineage"}
	}
	return unknown
}

func refutedReason(fields []string) string {
	if contains(fields, "unauthorized_dispatch") {
		return ReasonUnauthorized
	}
	if contains(fields, "conflicting_duplicate_dispatch") {
		return ReasonDuplicateConflict
	}
	if contains(fields, "source_run_id") || contains(fields, "source_run_attempt") || contains(fields, "source_artifact_id") {
		return ReasonMalformedIdentity
	}
	if contains(fields, "source_artifact_digest") || contains(fields, "source_artifact_archive_digest") || contains(fields, "source_artifact_observed_digest") {
		return ReasonDigestMismatch
	}
	if contains(fields, "source_workflow_name") || contains(fields, "source_workflow_path") || contains(fields, "source_repository") || contains(fields, "source_event") || contains(fields, "source_ref") || contains(fields, "dispatch_ref") || contains(fields, "target_workflow_name") || contains(fields, "target_workflow_path") || contains(fields, "target_workflow_identity") || contains(fields, "source_artifact_name") {
		return ReasonWorkflowMismatch
	}
	return ReasonWorkflowMismatch
}

func ValidateResolution(resolution ContinuationResolution) error {
	if resolution.Schema != ResolutionSchema || !validDigest(resolution.RequestDigest) || !validDigest(resolution.Digest) || resolution.Digest != resolutionDigest(resolution) {
		return errors.New("continuation resolution identity mismatch")
	}
	if resolution.ExecutionAuthorized || resolution.ExecutionGrants != 0 || resolution.LiveGrantDecision != 0 || resolution.LiveExecutionCount != 0 || resolution.GrantConsumedUses != 0 || resolution.RepositoryWrites != 0 || resolution.LocalTestExecutions != 0 {
		return errors.New("continuation crossed execution or grant boundary")
	}
	switch resolution.Decision {
	case DecisionClosed:
		if resolution.Resolution != ResolutionExact || resolution.Unknown != nil {
			return errors.New("continuation CLOSED resolution is not exact")
		}
	case DecisionUnknown:
		if resolution.Resolution != ResolutionLower || resolution.Unknown == nil || resolution.Unknown.Stage == "" || resolution.Unknown.Step == "" || resolution.Unknown.Reason == "" || resolution.Unknown.UnknownClass == "" || resolution.Unknown.NextOperation == "" || len(resolution.Unknown.BlockedBy) == 0 {
			return errors.New("continuation UNKNOWN resolution is not six-field causal")
		}
	case DecisionRefuted:
		if resolution.Resolution != ResolutionExact || resolution.Unknown != nil {
			return errors.New("continuation REFUTED resolution is not exact")
		}
	default:
		return errors.New("continuation decision is unknown")
	}
	return nil
}
