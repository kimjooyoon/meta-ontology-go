package publicworkflowlineage

import "fmt"

func (policy Policy) EvaluateReadOnlyObservation(input Input) ReadOnlyObservationEvaluation {
	evaluation := Evaluate(input)
	result := ReadOnlyObservationEvaluation{
		Schema:               ObservationSchema,
		Eligibility:          ObservationDenied,
		Decision:             evaluation.Decision,
		LineageState:         evaluation.LineageState,
		Reason:               evaluation.Reason,
		EvidenceReuseAllowed: false,
		PromotionAllowed:     false,
		SourceFailureKept:    evaluation.ProductFailureKept,
		Unknown:              evaluation.Unknown,
	}
	if !policy.readOnlyObservationPermissionsValid() {
		result.Reason = "canonical Gooo read-only observation permissions are unavailable"
		return result
	}
	if evaluation.Decision != DecisionClosed && !evaluation.ProductFailureKept {
		return result
	}
	result.Eligibility = ObservationAllowed
	result.ExactSourceIdentity = true
	result.TimingObservationEligible = true
	result.OperationObservationEligible = true
	return result
}

func (policy Policy) readOnlyObservationPermissionsValid() bool {
	return policy.ReadOnlyPermissions.WorkflowWindow == ReadOnlyPermission &&
		policy.ReadOnlyPermissions.VerificationRuntime == ReadOnlyPermission &&
		policy.ReadOnlyPermissions.EvidenceReuse == ExactSuccessReuse &&
		policy.ReadOnlyPermissions.Promotion == NoPromotionPermission
}

func Evaluate(input Input) Evaluation {
	evaluation := Evaluation{Decision: DecisionClosed, LineageState: StateExact, ExactSubjectBinding: true, ProvenanceState: ProvenanceExact}
	if input.Trigger.ConsumerWorkflow != "CI effort observation" || input.Trigger.ConsumerRunID <= 0 || input.Trigger.ConsumerRunAttempt <= 0 || input.Trigger.ConsumerSubjectSHA == "" || input.Trigger.ConsumerRef == "" {
		return refuted(StateMismatch, "consumer identity is incomplete or does not belong to the observer workflow", true, false)
	}
	if input.Trigger.CandidateSubjectSHA != "" && input.Trigger.CandidateSubjectSHA != input.Trigger.SourceSubjectSHA {
		return refuted(StateCurrentDevFallback, "current dev or merge head was offered as a substitute for the trigger subject", true, true)
	}
	if sourceIdentityMissing(input) {
		return unknownEvaluation(UnknownClassDirectMissing, "source workflow run API identity is absent or incomplete for the exact workflow_run.id key")
	}
	if mismatch := sourceIdentityContradiction(input); mismatch != "" {
		return refuted(StateMismatch, mismatch, true, false)
	}
	if mismatch := triggerMismatch(input); mismatch != "" {
		return refuted(StateMismatch, mismatch, true, false)
	}
	if input.Source.Status != "completed" {
		return unknownEvaluation(UnknownClassDirectMissing, "source workflow run is not completed for the exact trigger")
	}
	if input.Source.Conclusion != "success" {
		evaluation = refuted(StateMismatch, "source product or test failure remains a source failure", false, false)
		evaluation.ProductFailureKept = true
		return evaluation
	}
	artifact, count := matchingArtifact(input)
	if count > 1 {
		return refuted(StateTampered, "source artifact identity is duplicated", true, false)
	}
	if count == 0 {
		class, state, reason := missingClass(input)
		return Evaluation{Decision: DecisionUnknown, LineageState: state, Reason: reason, ExactSubjectBinding: true, ProvenanceState: ProvenanceUnknown, Unknown: unknown(class, reason)}
	}
	if artifact.RunID <= 0 || artifact.RunAttempt <= 0 || artifact.Size <= 0 || artifact.Digest == "" || artifact.SubjectSHA == "" || artifact.SubjectBinding == "" {
		return unknownEvaluation(UnknownClassDirectMissing, "exact source artifact metadata or payload subject binding is absent")
	}
	if artifact.RunID != input.Source.ID || artifact.RunAttempt != input.Source.RunAttempt || artifact.Name != input.ExpectedArtifactName || artifact.SubjectSHA != input.Trigger.SourceSubjectSHA || input.ExpectedArtifactSubjectBinding == "" || artifact.SubjectBinding != input.ExpectedArtifactSubjectBinding || input.ExpectedDigest != "" && artifact.Digest != input.ExpectedDigest || artifact.PayloadDigest != "" && artifact.PayloadDigest != artifact.Digest {
		return refuted(StateTampered, "source artifact subject or digest does not match the exact trigger lineage", true, false)
	}
	return Evaluation{Decision: DecisionClosed, LineageState: StateExact, Reason: "exact API workflow run, attempt, subject, artifact, and digest matched", ExactSubjectBinding: true, ProvenanceState: ProvenanceExact, ArtifactResolved: true, ArtifactIdentity: fmt.Sprintf("%d:%s:%s", artifact.ID, artifact.Name, artifact.Digest)}
}

func sourceIdentityMissing(input Input) bool {
	source := input.Source
	return source.ID <= 0 || source.APIQueryRunID <= 0 || source.ResolvedBy == "" || source.Workflow == "" || source.WorkflowPath == "" || source.WorkflowID <= 0 || source.Event == "" || source.HeadSHA == "" || source.Repository == "" || source.APIRepositoryName == "" || source.Status == "" || source.RunAttempt <= 0 || source.RefState == "" || source.Status == "completed" && source.Conclusion == ""
}

func sourceIdentityContradiction(input Input) string {
	source := input.Source
	if source.APIQueryRunID != source.ID {
		return "source API response was not resolved by the exact workflow_run.id query key"
	}
	if input.ExpectedSourceAPIKey == "" || input.ExpectedSourceAPIKey != "workflow_run.id" || source.ResolvedBy != "actions-run-api:"+input.ExpectedSourceAPIKey {
		return "source workflow run was not resolved by the declared Actions run API key"
	}
	if input.ExpectedWorkflow == "" || source.Workflow != input.ExpectedWorkflow || source.WorkflowPath != source.Workflow {
		return "source workflow API identity does not match the declared CI workflow path"
	}
	if input.ExpectedRepository == "" || source.APIRepositoryName != input.ExpectedRepository || source.Repository != input.ExpectedRepository {
		return "source API repository identity does not match the declared repository"
	}
	if source.APIRepositoryName != source.Repository {
		return "source API repository and head repository identities disagree"
	}
	return ""
}

func triggerMismatch(input Input) string {
	trigger, source := input.Trigger, input.Source
	if trigger.SourceWorkflow == "" || trigger.SourceWorkflow != source.Workflow {
		return "source workflow identity does not match the triggering workflow"
	}
	if trigger.SourceRunID == 0 || trigger.SourceRunID != source.ID {
		return "source run identity does not match the triggering workflow run"
	}
	if trigger.SourceRunAttempt == 0 || trigger.SourceRunAttempt != source.RunAttempt {
		return "source run attempt does not match the triggering workflow attempt"
	}
	if trigger.SourceSubjectSHA == "" || trigger.SourceSubjectSHA != source.HeadSHA {
		return "source subject SHA does not match the triggering workflow subject"
	}
	if trigger.SourceRefState != "" && trigger.SourceRefState != RefStateMissing && source.RefState != "" && source.RefState != RefStateMissing && trigger.SourceRefState != source.RefState {
		return "source ref observation state does not match the API response"
	}
	if source.RefState == RefStateValue && trigger.SourceRef != source.Ref {
		return "source ref observation does not match the API response"
	}
	if source.RefState != RefStateValue && source.Ref != "" {
		return "source ref has a value despite its declared null or empty observation state"
	}
	if trigger.SourceHeadBranch != "" && source.HeadBranch != "" && trigger.SourceHeadBranch != source.HeadBranch {
		return "source head_branch observation does not match the API response"
	}
	if trigger.SourceEvent != "" && source.Event != "" && trigger.SourceEvent != source.Event {
		return "source event does not match the triggering workflow event"
	}
	if trigger.SourceRepository != "" && source.Repository != "" && trigger.SourceRepository != source.Repository {
		return "source repository does not match the triggering workflow repository"
	}
	return ""
}

func matchingArtifact(input Input) (Artifact, int) {
	var found Artifact
	count := 0
	for _, artifact := range input.Artifacts.Artifacts {
		if artifact.Name == input.ExpectedArtifactName && !artifact.Expired {
			found = artifact
			count++
		}
	}
	return found, count
}

func missingClass(input Input) (string, string, string) {
	if input.Artifacts.LookupStatus == "unavailable" || input.Source.ID == 0 || input.Source.Status != "completed" {
		return UnknownClassDirectMissing, StateDirectMissing, "source artifact lookup is absent, pending, or unavailable for the exact trigger"
	}
	return UnknownClassStale, StateStale, "source artifact is absent or expired for the exact completed trigger"
}

func unknown(class, reason string) *CausalUnknown {
	next := "resolve_exact_source_artifact_identity"
	if class == UnknownClassDirectMissing {
		next = "retry_exact_trigger_artifact_lookup"
	}
	return &CausalUnknown{Stage: "workflow-lineage", Step: "source-artifact-resolution", Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: []string{"workflow_run.id", "source-run-artifact", "subject-sha", "run-attempt"}}
}

func unknownEvaluation(class, reason string) Evaluation {
	return Evaluation{Decision: DecisionUnknown, LineageState: StateDirectMissing, Reason: reason, ExactSubjectBinding: true, ProvenanceState: ProvenanceUnknown, Unknown: unknown(class, reason)}
}

func refuted(state, reason string, mismatch, fallback bool) Evaluation {
	return Evaluation{Decision: DecisionRefuted, LineageState: state, Reason: reason, ProvenanceState: ProvenanceRefuted, MismatchDetected: mismatch, FallbackAttempted: fallback, FallbackRejected: fallback}
}
