package selfimprovementexecutioncontract

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

// Verify is an independent consumer. It reclassifies the input without
// calling Evaluate, then compares the resulting decision coordinates and
// safety envelope to the emitted contract declaration.
func Verify(program PolicyProgram, input ContractInput, resolution ContractResolution) Verification {
	decision, contractResolution, reason := independentClassify(program, input)
	verification := Verification{
		Schema: VerificationSchema, ContractDigest: resolution.Digest,
		IndependentDecision: decision, IndependentResolution: contractResolution,
		IndependentReason: reason, IndependentReplayComparisons: 1,
		RepositoryWrites: 0, LocalTestExecutions: 0, ExecutionGrants: 0,
		Verified: resolution.Decision == decision && resolution.Resolution == contractResolution && resolution.Reason == reason &&
			!resolution.ExecutionAuthorized && resolution.ExecutionGrantRequired && resolution.MaxExecutions == 1 && !resolution.RepositoryWritesAllowed &&
			resolutionInputMatches(resolution, input) && ValidateResolution(resolution) == nil}
	verification.Digest = verificationDigest(verification)
	return verification
}

func independentClassify(program PolicyProgram, input ContractInput) (Decision, Resolution, string) {
	missing, contradictions := independentFields(input)
	if len(contradictions) > 0 {
		return DecisionRefuted, ResolutionExact, independentReason(contradictions)
	}
	if len(missing) > 0 || !independentMapping(input) || !independentPolicy(program) {
		return DecisionUnknown, ResolutionLower, independentUnknownReason(program, input, missing)
	}
	return DecisionClosed, ResolutionDeclared, ReasonDeclared
}

func independentFields(input ContractInput) ([]string, []string) {
	missing, contradictions := []string{}, []string{}
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
	if input.Candidate.ExecutionInputDigest == "" || input.ExecutionInput == nil || executionInputIncomplete(input.ExecutionInput) {
		missing = append(missing, "candidate_input_digest")
	} else {
		if input.Candidate.ExecutionInputDigest != input.Candidate.InputDigest ||
			input.Candidate.ExecutionInputDigest != input.ExecutionInput.Digest ||
			input.ExecutionInput.CandidateStableID != input.Candidate.StableID ||
			input.ExecutionInput.CandidateDigest != input.Candidate.Digest ||
			input.ExecutionInput.SubjectSHA != input.Candidate.SubjectSHA ||
			input.ExecutionInput.ObservationDigest != input.Candidate.ObservationDigest {
			contradictions = append(contradictions, "candidate_input_digest")
		}
		if err := valuewitnessinput.Validate(*input.ExecutionInput); err != nil {
			contradictions = append(contradictions, "candidate_input_digest")
		}
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
	if input.Registry.Schema == "" || !input.Registry.SafetyDeclared {
		missing = append(missing, "evaluator_registry_digest")
	}
	if input.Registry.InputAuthority == "" {
		missing = append(missing, "input_authority")
	}
	if input.Registry.OutputAuthority == "" {
		missing = append(missing, "output_authority")
	}
	if input.Registry.MaxExecutions == 0 {
		missing = append(missing, "max_executions")
	}
	if input.Candidate.StableID != "" && !validDigest(input.Candidate.StableID) {
		contradictions = append(contradictions, "candidate_stable_id")
	}
	if input.Candidate.Digest != "" && !validDigest(input.Candidate.Digest) {
		contradictions = append(contradictions, "candidate_digest")
	}
	if input.Candidate.SubjectSHA != "" && !validSHA(input.Candidate.SubjectSHA) {
		contradictions = append(contradictions, "subject_sha")
	}
	if input.Candidate.InputDigest != "" && !validDigest(input.Candidate.InputDigest) {
		contradictions = append(contradictions, "candidate_input_digest")
	}
	if input.Candidate.ExecutionInputDigest != "" && !validDigest(input.Candidate.ExecutionInputDigest) {
		contradictions = append(contradictions, "candidate_input_digest")
	}
	if input.Candidate.ObservationDigest != "" && !validDigest(input.Candidate.ObservationDigest) {
		contradictions = append(contradictions, "observation_digest")
	}
	if input.Observation.CandidateInputDigest != "" && input.Observation.CandidateInputDigest != input.Candidate.InputDigest {
		contradictions = append(contradictions, "candidate_input_digest")
	}
	if input.Observation.SourceObservationDigest != "" &&
		(!validDigest(input.Observation.SourceObservationDigest) || input.Observation.SourceObservationDigest != input.Candidate.ObservationDigest) {
		contradictions = append(contradictions, "observation_digest")
	}
	if input.Observation.CandidateStableID != "" && input.Observation.CandidateStableID != input.Candidate.StableID {
		contradictions = append(contradictions, "candidate_stable_id")
	}
	if input.Observation.SubjectSHA != "" && input.Observation.SubjectSHA != input.Candidate.SubjectSHA {
		contradictions = append(contradictions, "subject_sha")
	}
	if input.Observation.ObservationDigest != "" && input.Observation.ObservationDigest != observationDigest(input.Observation) {
		contradictions = append(contradictions, "observation_digest")
	}
	for _, span := range input.Observation.SourceSpans {
		if !validSpan(span) {
			contradictions = append(contradictions, "source_spans")
			break
		}
	}
	if input.Observation.ObservedCount < 0 {
		contradictions = append(contradictions, "observed_count")
	}
	known := KnownRegistry()
	if input.Observation.Schema != "" && input.Observation.Schema != ObservationSchema {
		contradictions = append(contradictions, "observation_schema")
	}
	if input.Observation.Phase != "" && input.Observation.Phase != Phase(KnownPhase) {
		contradictions = append(contradictions, "phase")
	}
	if input.Observation.BoundedTarget != "" && input.Observation.BoundedTarget != KnownBoundedTarget {
		contradictions = append(contradictions, "bounded_target")
	}
	if input.Registry.Schema != "" && input.Registry.Schema != known.Schema {
		contradictions = append(contradictions, "registry_schema")
	}
	if input.Registry.Phase != "" && input.Registry.Phase != known.Phase {
		contradictions = append(contradictions, "phase")
	}
	if input.Registry.OperationID != "" && input.Registry.OperationID != known.OperationID {
		contradictions = append(contradictions, "registry_operation")
	}
	if input.Registry.BoundedTarget != "" && input.Registry.BoundedTarget != known.BoundedTarget {
		contradictions = append(contradictions, "bounded_target")
	}
	if input.Registry.InputAuthority != "" && input.Registry.InputAuthority != known.InputAuthority {
		contradictions = append(contradictions, "input_authority")
	}
	if input.Registry.OutputAuthority != "" && input.Registry.OutputAuthority != known.OutputAuthority {
		contradictions = append(contradictions, "output_authority")
	}
	if input.Registry.EvaluatorRegistryDigest != "" && input.Registry.EvaluatorRegistryDigest != known.EvaluatorRegistryDigest {
		contradictions = append(contradictions, "evaluator_registry_digest")
	}
	if input.Registry.ToolchainTestContractIdentity != "" && input.Registry.ToolchainTestContractIdentity != known.ToolchainTestContractIdentity {
		contradictions = append(contradictions, "toolchain_test_contract_identity")
	}
	if input.Registry.RepositoryWritesAllowed {
		contradictions = append(contradictions, "repository_writes_allowed")
	}
	if input.Registry.MaxExecutions > 1 {
		contradictions = append(contradictions, "max_executions")
	}
	if input.Registry.MaxExecutions < 0 || input.Registry.SafetyDeclared && input.Registry.MaxExecutions != 1 {
		contradictions = append(contradictions, "max_executions")
	}
	if input.Authorization.ExecutionAllowed || input.Authorization.RepositoryWrites != 0 || input.Authorization.LocalTestExecutions != 0 {
		contradictions = append(contradictions, "execution_boundary")
	}
	if input.Authorization.RequestSchema != "" && input.Authorization.RequestSchema != selfimprovementcandidate.AuthorizationRequestSchema {
		contradictions = append(contradictions, "authorization_request")
	}
	if input.Authorization.ContractID != "" && input.Authorization.ContractID != selfimprovementcandidate.AuthorizationContractID {
		contradictions = append(contradictions, "authorization_contract")
	}
	if input.Authorization.Scope != "" && input.Authorization.Scope != selfimprovementcandidate.AuthorizationScope {
		contradictions = append(contradictions, "authorization_scope")
	}
	if input.Authorization.RequestDigest != "" && !validDigest(input.Authorization.RequestDigest) {
		contradictions = append(contradictions, "authorization_request")
	}
	if input.Authorization.ContractDigest != "" && !validDigest(input.Authorization.ContractDigest) {
		contradictions = append(contradictions, "authorization_contract")
	}
	if input.Authorization.LiveAuthorized != 0 || input.Authorization.LiveState != "" && input.Authorization.LiveState != "UNKNOWN" {
		contradictions = append(contradictions, "authorization_request")
	}
	return sortedStrings(missing), sortedStrings(contradictions)
}

func independentMapping(input ContractInput) bool {
	return input.Candidate.MetaOperation == CandidateMetaOperation && input.Candidate.ExperimentKind == CandidateExperimentKind &&
		input.Observation.OperationID == OperationID(KnownOperationID) && input.Observation.BoundedTarget == KnownBoundedTarget &&
		input.ExecutionInput != nil && !executionInputIncomplete(input.ExecutionInput) &&
		input.Candidate.InputDigest == input.Candidate.ExecutionInputDigest && input.Candidate.ExecutionInputDigest == input.ExecutionInput.Digest &&
		input.ExecutionInput.OperationID == KnownOperationID && input.ExecutionInput.BoundedTarget == KnownBoundedTarget &&
		input.ExecutionInput.Phase == KnownPhase && valuewitnessinput.Validate(*input.ExecutionInput) == nil
}

func independentPolicy(program PolicyProgram) bool {
	return program.Evidence.Schema == PolicySchema && program.Evidence.CaseCount == 9 && program.Evidence.ClosedCases == 3 && program.Evidence.UnknownCases == 3 && program.Evidence.RefutedCases == 3
}

func independentUnknownReason(program PolicyProgram, input ContractInput, missing []string) string {
	if !independentPolicy(program) {
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
	if len(missing) > 0 {
		return ReasonMissingField
	}
	return ReasonMissingObservation
}

func independentReason(contradictions []string) string {
	if contains(contradictions, "max_executions") || contains(contradictions, "repository_writes_allowed") || contains(contradictions, "execution_boundary") {
		return ReasonSafetyConflict
	}
	if contains(contradictions, "source_spans") {
		return ReasonSpanConflict
	}
	if contains(contradictions, "registry_schema") || contains(contradictions, "evaluator_registry_digest") || contains(contradictions, "toolchain_test_contract_identity") || contains(contradictions, "registry_operation") || contains(contradictions, "operation_id") {
		return ReasonRegistryConflict
	}
	if contains(contradictions, "authorization_request") || contains(contradictions, "authorization_contract") || contains(contradictions, "authorization_scope") {
		return ReasonAuthorizationConflict
	}
	if contains(contradictions, "candidate_input_digest") || contains(contradictions, "candidate_stable_id") || contains(contradictions, "subject_sha") {
		return ReasonCandidateConflict
	}
	return ReasonCandidateConflict
}

func resolutionInputMatches(resolution ContractResolution, input ContractInput) bool {
	if input.ExecutionInput == nil {
		return resolution.ExecutionInput == nil && resolution.ExecutionInputDigest == ""
	}
	return resolution.ExecutionInput != nil && resolution.ExecutionInputDigest == input.ExecutionInput.Digest &&
		resolution.CandidateInputDigest == input.ExecutionInput.Digest &&
		resolution.ExecutionInput.Digest == input.ExecutionInput.Digest &&
		resolution.ExecutionInput.CandidateStableID == input.Candidate.StableID &&
		resolution.ExecutionInput.CandidateDigest == input.Candidate.Digest &&
		resolution.ExecutionInput.SubjectSHA == input.Candidate.SubjectSHA &&
		resolution.ExecutionInput.ObservationDigest == input.Candidate.ObservationDigest
}

func VerifyResolution(resolution ContractResolution) error {
	if resolution.Schema != Schema || resolution.Digest == "" {
		return errors.New("missing execution contract resolution")
	}
	if resolution.ExecutionInput != nil && valuewitnessinput.Validate(*resolution.ExecutionInput) != nil {
		return errors.New("execution contract resolution carries invalid execution input")
	}
	if resolution.ExecutionInput != nil && (resolution.ExecutionInputDigest != resolution.ExecutionInput.Digest || resolution.CandidateInputDigest != resolution.ExecutionInput.Digest) {
		return errors.New("execution contract resolution input digest is not bound")
	}
	if resolution.Decision == DecisionClosed && resolution.ExecutionInput == nil {
		return errors.New("closed execution contract omitted execution input")
	}
	return ValidateResolution(resolution)
}
