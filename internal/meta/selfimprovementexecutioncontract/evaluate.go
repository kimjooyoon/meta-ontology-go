package selfimprovementexecutioncontract

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

const (
	ObservationSchema           = "gooo/self-improvement-execution-observation/v1"
	ReasonDeclared              = "PRE_EXECUTION_CONTRACT_DECLARED"
	ReasonReplayed              = "PRE_EXECUTION_CONTRACT_REPLAYED"
	ReasonSafetyBoundary        = "EXECUTION_REQUIRES_SEPARATE_GRANT"
	ReasonMissingObservation    = "MISSING_EXACT_OBSERVATION_EVIDENCE"
	ReasonUnsupportedMapping    = "UNSUPPORTED_OPERATION_MAPPING"
	ReasonMissingRegistry       = "MISSING_REGISTRY_IDENTITY"
	ReasonMissingField          = "MISSING_PRE_EXECUTION_FIELD"
	ReasonCandidateConflict     = "CANDIDATE_OBSERVATION_CONTRADICTION"
	ReasonSpanConflict          = "SOURCE_SPAN_CONTRADICTION"
	ReasonSafetyConflict        = "SAFETY_ENVELOPE_CONTRADICTION"
	ReasonRegistryConflict      = "REGISTRY_CONTRADICTION"
	ReasonAuthorizationConflict = "AUTHORIZATION_CONTRADICTION"
	ReasonMissingExecutionInput = "MISSING_EXECUTION_INPUT_ARTIFACT"
)

func Evaluate(program PolicyProgram, input ContractInput) ContractResolution {
	resolution := baseResolution(program, input)
	missing, contradictions := inspectInput(input)
	resolution.MissingFields = sortedStrings(missing)
	resolution.ContradictoryFields = sortedStrings(contradictions)
	resolution.Metrics.BoundFields = max(PreExecutionRequiredField-len(requiredMissing(input)), 0)
	resolution.Metrics.MissingFields = len(resolution.MissingFields)
	resolution.Metrics.ContradictoryFields = len(resolution.ContradictoryFields)
	resolution.Metrics.StructuralCandidateToOperationAfter = boolInt(operationMapped(input))

	if len(contradictions) != 0 {
		resolution.Decision = DecisionRefuted
		resolution.Resolution = ResolutionExact
		resolution.Reason = contradictions[0]
		if contains(contradictions, "source_spans") {
			resolution.Reason = ReasonSpanConflict
		}
		if contains(contradictions, "max_executions") || contains(contradictions, "repository_writes_allowed") || contains(contradictions, "execution_allowed") {
			resolution.Reason = ReasonSafetyConflict
		}
		if contains(contradictions, "candidate_input_digest") || contains(contradictions, "candidate_stable_id") || contains(contradictions, "subject_sha") {
			resolution.Reason = ReasonCandidateConflict
		}
		if contains(contradictions, "registry_schema") || contains(contradictions, "evaluator_registry_digest") || contains(contradictions, "toolchain_test_contract_identity") || contains(contradictions, "registry_operation") || contains(contradictions, "operation_id") {
			resolution.Reason = ReasonRegistryConflict
		}
		if contains(contradictions, "authorization_request") || contains(contradictions, "authorization_contract") || contains(contradictions, "authorization_scope") || contains(contradictions, "live_authorized") || contains(contradictions, "live_state") {
			resolution.Reason = ReasonAuthorizationConflict
		}
		resolution.Digest = resolutionDigest(resolution)
		return resolution
	}

	if len(missing) != 0 || !operationMapped(input) || !policyReady(program) {
		resolution.Decision = DecisionUnknown
		resolution.Resolution = ResolutionLower
		resolution.Reason = unknownReason(program, input, missing)
		resolution.Unknown = unknownFor(resolution.Reason)
		resolution.Digest = resolutionDigest(resolution)
		return resolution
	}

	resolution.Decision = DecisionClosed
	resolution.Resolution = ResolutionDeclared
	resolution.Reason = ReasonDeclared
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func baseResolution(program PolicyProgram, input ContractInput) ContractResolution {
	return ContractResolution{
		Schema: Schema, ContractID: ContractID, Policy: program.Evidence,
		RequiredFields:    RequiredFieldNames(),
		CandidateStableID: input.Candidate.StableID, CandidateDigest: input.Candidate.Digest,
		SubjectSHA: input.Candidate.SubjectSHA, ObservationDigest: firstNonEmpty(input.Observation.ObservationDigest, input.Candidate.ObservationDigest),
		CandidateInputDigest: input.Candidate.InputDigest, ExecutionInputDigest: input.Candidate.ExecutionInputDigest, ExecutionInput: input.ExecutionInput, Phase: input.Observation.Phase,
		OperationID: input.Observation.OperationID, BoundedTarget: input.Observation.BoundedTarget,
		EvaluatorRegistryDigest: input.Registry.EvaluatorRegistryDigest,
		ToolchainTestContractID: input.Registry.ToolchainTestContractIdentity,
		InputAuthority:          input.Registry.InputAuthority, OutputAuthority: input.Registry.OutputAuthority,
		MaxExecutions: 1, RepositoryWritesAllowed: false, ExecutionAuthorized: false,
		ExecutionGrantRequired: true, ExecutionGrantBlockedBy: ExecutionGrantBlockedBy,
		OutputEvidenceDeferred: true, RuntimeResultDeferred: true,
		Metrics: Metrics{PreExecutionRequiredFields: PreExecutionRequiredField,
			StructuralCandidateToOperationBefore: 0, LiveExecutionCount: 0,
			CanonicalExecutionCount: 0, ExecutionGrants: 0, RepositoryWrites: 0,
			LocalTestExecutions: 0, FallbackAccepted: 0, PerformanceImprovement: PerformanceUnknown,
			GoPhysicalLines: program.Inventory.GoPhysicalLines, GoooPhysicalLines: program.Inventory.GoooPhysicalLines},
	}
}

func inspectInput(input ContractInput) ([]string, []string) {
	missing, contradictions := requiredMissing(input), []string{}
	addMissing := func(field string) {
		if !contains(missing, field) {
			missing = append(missing, field)
		}
	}
	addContradiction := func(field string) {
		if !contains(contradictions, field) {
			contradictions = append(contradictions, field)
		}
	}

	if input.Candidate.StableID != "" && !validDigest(input.Candidate.StableID) {
		addContradiction("candidate_stable_id")
	}
	if input.Candidate.Digest != "" && !validDigest(input.Candidate.Digest) {
		addContradiction("candidate_digest")
	}
	if input.Candidate.SubjectSHA != "" && !validSHA(input.Candidate.SubjectSHA) {
		addContradiction("subject_sha")
	}
	if input.Candidate.ObservationDigest != "" && !validDigest(input.Candidate.ObservationDigest) {
		addContradiction("observation_digest")
	}
	if input.Candidate.InputDigest != "" && !validDigest(input.Candidate.InputDigest) {
		addContradiction("candidate_input_digest")
	}
	if input.Candidate.ExecutionInputDigest != "" && !validDigest(input.Candidate.ExecutionInputDigest) {
		addContradiction("candidate_input_digest")
	}
	if input.ExecutionInput == nil || executionInputIncomplete(input.ExecutionInput) {
		addMissing("candidate_input_digest")
	} else {
		if input.Candidate.ExecutionInputDigest != input.Candidate.InputDigest ||
			input.Candidate.ExecutionInputDigest != input.ExecutionInput.Digest ||
			input.ExecutionInput.CandidateStableID != input.Candidate.StableID ||
			input.ExecutionInput.CandidateDigest != input.Candidate.Digest ||
			input.ExecutionInput.SubjectSHA != input.Candidate.SubjectSHA ||
			input.ExecutionInput.ObservationDigest != input.Candidate.ObservationDigest {
			addContradiction("candidate_input_digest")
		}
		if valuewitnessinput.Validate(*input.ExecutionInput) != nil {
			addContradiction("candidate_input_digest")
		}
	}

	observation := input.Observation
	if observation.Schema != "" && observation.Schema != ObservationSchema {
		addContradiction("observation_schema")
	}
	if observation.ObservationDigest != "" {
		if !validDigest(observation.ObservationDigest) || observation.ObservationDigest != observationDigest(observation) {
			addContradiction("observation_digest")
		}
	}
	if observation.CandidateInputDigest != "" && input.Candidate.InputDigest != "" && observation.CandidateInputDigest != input.Candidate.InputDigest {
		addContradiction("candidate_input_digest")
	}
	if observation.SourceObservationDigest != "" {
		if !validDigest(observation.SourceObservationDigest) || observation.SourceObservationDigest != input.Candidate.ObservationDigest {
			addContradiction("observation_digest")
		}
	}
	if observation.CandidateStableID != "" && input.Candidate.StableID != "" && observation.CandidateStableID != input.Candidate.StableID {
		addContradiction("candidate_stable_id")
	}
	if observation.SubjectSHA != "" && input.Candidate.SubjectSHA != "" && observation.SubjectSHA != input.Candidate.SubjectSHA {
		addContradiction("subject_sha")
	}
	if len(observation.SourceSpans) > 0 {
		for _, span := range observation.SourceSpans {
			if !validSpan(span) {
				addContradiction("source_spans")
				break
			}
		}
	}
	if observation.ObservedCount < 0 {
		addContradiction("observed_count")
	}
	if observation.Phase != "" && observation.Phase != Phase(KnownPhase) {
		addContradiction("phase")
	}
	if observation.OperationID != "" && observation.OperationID != OperationID(KnownOperationID) {
		addMissing("operation_mapping")
	}
	if observation.BoundedTarget != "" && observation.BoundedTarget != KnownBoundedTarget {
		addContradiction("bounded_target")
	}

	registry := input.Registry
	known := KnownRegistry()
	if registry.Schema != "" && registry.Schema != known.Schema {
		addContradiction("registry_schema")
	}
	if registry.EvaluatorRegistryDigest != "" && registry.EvaluatorRegistryDigest != known.EvaluatorRegistryDigest {
		addContradiction("evaluator_registry_digest")
	}
	if registry.ToolchainTestContractIdentity != "" && registry.ToolchainTestContractIdentity != known.ToolchainTestContractIdentity {
		addContradiction("toolchain_test_contract_identity")
	}
	if registry.Phase != "" && registry.Phase != known.Phase {
		addContradiction("phase")
	}
	if registry.OperationID != "" && registry.OperationID != known.OperationID {
		addContradiction("operation_id")
	}
	if registry.BoundedTarget != "" && registry.BoundedTarget != known.BoundedTarget {
		addContradiction("bounded_target")
	}
	if registry.InputAuthority != "" && registry.InputAuthority != known.InputAuthority {
		addContradiction("input_authority")
	}
	if registry.OutputAuthority != "" && registry.OutputAuthority != known.OutputAuthority {
		addContradiction("output_authority")
	}
	if registry.MaxExecutions > 1 || registry.MaxExecutions < 0 {
		addContradiction("max_executions")
	}
	if registry.RepositoryWritesAllowed {
		addContradiction("repository_writes_allowed")
	}
	if registry.SafetyDeclared && registry.MaxExecutions != 1 {
		addContradiction("max_executions")
	}

	authorization := input.Authorization
	if authorization.RequestSchema != "" && authorization.RequestSchema != selfimprovementcandidate.AuthorizationRequestSchema {
		addContradiction("authorization_request")
	}
	if authorization.RequestDigest != "" && !validDigest(authorization.RequestDigest) {
		addContradiction("authorization_request")
	}
	if authorization.ContractID != "" && authorization.ContractID != selfimprovementcandidate.AuthorizationContractID {
		addContradiction("authorization_contract")
	}
	if authorization.ContractDigest != "" && !validDigest(authorization.ContractDigest) {
		addContradiction("authorization_contract")
	}
	if authorization.Scope != "" && authorization.Scope != selfimprovementcandidate.AuthorizationScope {
		addContradiction("authorization_scope")
	}
	if authorization.ExecutionAllowed {
		addContradiction("execution_allowed")
	}
	if authorization.RepositoryWrites != 0 {
		addContradiction("repository_writes_allowed")
	}
	if authorization.LocalTestExecutions != 0 {
		addContradiction("local_test_executions")
	}
	if authorization.LiveAuthorized != 0 {
		addContradiction("live_authorized")
	}
	if authorization.LiveState != "" && authorization.LiveState != "UNKNOWN" {
		addContradiction("live_state")
	}

	if input.Candidate.MetaOperation == "" || input.Candidate.MetaOperation != CandidateMetaOperation || input.Candidate.ExperimentKind == "" || input.Candidate.ExperimentKind != CandidateExperimentKind {
		addMissing("operation_mapping")
	}
	if input.Registry.Schema == "" || input.Registry.EvaluatorRegistryDigest == "" || input.Registry.ToolchainTestContractIdentity == "" || !input.Registry.SafetyDeclared {
		addMissing("evaluator_registry_digest")
	}
	if input.Registry.ToolchainTestContractIdentity == "" {
		addMissing("toolchain_test_contract_identity")
	}
	if input.Registry.InputAuthority == "" {
		addMissing("input_authority")
	}
	if input.Registry.OutputAuthority == "" {
		addMissing("output_authority")
	}
	if input.Registry.MaxExecutions == 0 || !input.Registry.SafetyDeclared {
		addMissing("max_executions")
	}
	if !input.Registry.SafetyDeclared {
		addMissing("repository_writes_allowed")
	}

	return missing, contradictions
}

func requiredMissing(input ContractInput) []string {
	missing := []string{}
	if input.Candidate.StableID == "" {
		missing = append(missing, "candidate_stable_id")
	}
	if input.Candidate.Digest == "" {
		missing = append(missing, "candidate_digest")
	}
	if input.Candidate.SubjectSHA == "" {
		missing = append(missing, "subject_sha")
	}
	if input.Candidate.ObservationDigest == "" {
		missing = append(missing, "observation_digest")
	}
	if input.Candidate.InputDigest == "" {
		missing = append(missing, "candidate_input_digest")
	}
	if input.Candidate.InputDigest != "" && (input.Candidate.ExecutionInputDigest == "" || input.ExecutionInput == nil || executionInputIncomplete(input.ExecutionInput)) {
		missing = append(missing, "candidate_input_digest")
	}
	if input.Observation.Phase == "" {
		missing = append(missing, "phase")
	}
	if input.Observation.OperationID == "" {
		missing = append(missing, "operation_id")
	}
	if input.Observation.BoundedTarget == "" {
		missing = append(missing, "bounded_target")
	}
	if len(input.Observation.SourceSpans) == 0 {
		missing = append(missing, "source_spans")
	}
	if !input.Observation.ObservedCountKnown {
		missing = append(missing, "observed_count")
	}
	if input.Registry.EvaluatorRegistryDigest == "" {
		missing = append(missing, "evaluator_registry_digest")
	}
	if input.Registry.ToolchainTestContractIdentity == "" {
		missing = append(missing, "toolchain_test_contract_identity")
	}
	return missing
}

func operationMapped(input ContractInput) bool {
	return input.Candidate.MetaOperation == CandidateMetaOperation && input.Candidate.ExperimentKind == CandidateExperimentKind &&
		input.Observation.OperationID == OperationID(KnownOperationID) && input.Observation.BoundedTarget == KnownBoundedTarget &&
		input.ExecutionInput != nil && !executionInputIncomplete(input.ExecutionInput) &&
		input.Candidate.InputDigest == input.Candidate.ExecutionInputDigest && input.Candidate.ExecutionInputDigest == input.ExecutionInput.Digest &&
		input.ExecutionInput.OperationID == KnownOperationID && input.ExecutionInput.BoundedTarget == KnownBoundedTarget &&
		input.ExecutionInput.Phase == KnownPhase && valuewitnessinput.Validate(*input.ExecutionInput) == nil
}

func policyReady(program PolicyProgram) bool {
	return program.Evidence.Schema == PolicySchema && program.Evidence.CaseCount == 9 &&
		program.Evidence.ClosedCases == 3 && program.Evidence.UnknownCases == 3 && program.Evidence.RefutedCases == 3
}

func unknownReason(program PolicyProgram, input ContractInput, missing []string) string {
	if !policyReady(program) {
		return ReasonMissingField
	}
	if input.Registry.EvaluatorRegistryDigest == "" || input.Registry.ToolchainTestContractIdentity == "" || input.Registry.Schema == "" || !input.Registry.SafetyDeclared {
		return ReasonMissingRegistry
	}
	if input.ExecutionInput == nil || executionInputIncomplete(input.ExecutionInput) {
		return ReasonMissingExecutionInput
	}
	if input.Candidate.MetaOperation == "" || input.Candidate.MetaOperation != CandidateMetaOperation || input.Candidate.ExperimentKind == "" || input.Candidate.ExperimentKind != CandidateExperimentKind || (input.Observation.OperationID != "" && input.Observation.OperationID != OperationID(KnownOperationID)) {
		return ReasonUnsupportedMapping
	}
	if len(input.Observation.SourceSpans) == 0 || !input.Observation.ObservedCountKnown || input.Observation.ObservationDigest == "" {
		return ReasonMissingObservation
	}
	if len(missing) != 0 {
		return ReasonMissingField
	}
	return ReasonMissingObservation
}

func unknownFor(reason string) *UnknownState {
	unknown := UnknownState{Stage: "BIND", Step: "1", Reason: reason, UnknownClass: "INCOMPLETE_EVIDENCE", NextOperation: "bind-exact-pre-execution-evidence", BlockedBy: []string{"pre_execution_evidence"}}
	switch reason {
	case ReasonUnsupportedMapping:
		unknown.Stage, unknown.Step, unknown.UnknownClass, unknown.NextOperation, unknown.BlockedBy = "BIND", "2", "UNSUPPORTED_MAPPING", "register-supported-operation", []string{"operation_registry"}
	case ReasonMissingRegistry:
		unknown.Stage, unknown.Step, unknown.NextOperation, unknown.BlockedBy = "BIND", "3", "bind-fixed-registry-identity", []string{"registry_identity"}
	case ReasonMissingObservation:
		unknown.Stage, unknown.Step, unknown.NextOperation, unknown.BlockedBy = "OBSERVE", "4", "bind-exact-observation-evidence", []string{"observation_evidence"}
	case ReasonMissingExecutionInput:
		unknown.Stage, unknown.Step, unknown.UnknownClass, unknown.NextOperation, unknown.BlockedBy = "BIND", "5", "INCOMPLETE_EVIDENCE", "bind-exact-execution-input", []string{"execution_input_artifact"}
	}
	return &unknown
}

func observationDigest(value ObservationEvidence) string {
	value.ObservationDigest = ""
	return digestJSON(value)
}

func validSpan(span SourceSpan) bool {
	return strings.TrimSpace(span.SourceID) != "" && !strings.HasPrefix(span.SourceID, "/") && span.StartLine > 0 && span.EndLine >= span.StartLine
}

func executionInputIncomplete(input *valuewitnessinput.ExecutionInput) bool {
	return input == nil || input.Schema == "" || input.ContractID == "" || input.CandidateStableID == "" ||
		input.CandidateDigest == "" || input.SubjectSHA == "" || input.ObservationDigest == "" ||
		input.OperationID == "" || input.BoundedTarget == "" || input.Phase == "" || input.Source.Path == "" ||
		input.Source.Bytes == "" || input.Source.Digest == "" || input.Activity.Name == "" ||
		input.Activity.SemanticFingerprint == "" || len(input.Corpus) == 0 || input.CorpusDigest == "" ||
		input.EvaluatorRegistry.Schema == "" || input.EvaluatorRegistry.Digest == "" || input.ToolchainTestContractID == "" ||
		input.OutputSchema == "" || input.InputAuthority == "" || input.OutputAuthority == "" || input.MaxExecutions == 0
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func contains(values []string, target string) bool {
	return slices.Contains(values, target)
}
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func ValidateResolution(resolution ContractResolution) error {
	if resolution.Schema != Schema || resolution.ContractID != ContractID {
		return errors.New("execution contract resolution identity mismatch")
	}
	if resolution.ExecutionAuthorized || resolution.ExecutionGrantRequired != true || resolution.ExecutionGrantBlockedBy != ExecutionGrantBlockedBy || resolution.MaxExecutions != 1 || resolution.RepositoryWritesAllowed || !resolution.OutputEvidenceDeferred || !resolution.RuntimeResultDeferred {
		return errors.New("execution contract crossed its non-executing boundary")
	}
	if resolution.Decision == DecisionUnknown {
		if resolution.Unknown == nil || resolution.Unknown.Stage == "" || resolution.Unknown.Step == "" || resolution.Unknown.Reason == "" || resolution.Unknown.UnknownClass == "" || resolution.Unknown.NextOperation == "" || len(resolution.Unknown.BlockedBy) == 0 {
			return errors.New("unknown resolution does not have six causal fields")
		}
	}
	if resolution.Digest != resolutionDigest(resolution) {
		return fmt.Errorf("execution contract resolution digest mismatch")
	}
	return nil
}
