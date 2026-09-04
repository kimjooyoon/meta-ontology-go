package publicworkflowlineage

import "fmt"

func Evaluate(input Input) Evaluation {
	evaluation := Evaluation{Decision: DecisionClosed, LineageState: StateExact, ExactSubjectBinding: true}
	if input.Trigger.ConsumerWorkflow != "CI effort observation" || input.Trigger.ConsumerRunID <= 0 || input.Trigger.ConsumerRunAttempt <= 0 || input.Trigger.ConsumerSubjectSHA == "" || input.Trigger.ConsumerRef == "" {
		return refuted(StateMismatch, "consumer identity is incomplete or does not belong to the observer workflow", true, false)
	}
	if input.Trigger.CandidateSubjectSHA != "" && input.Trigger.CandidateSubjectSHA != input.Trigger.SourceSubjectSHA {
		return refuted(StateCurrentDevFallback, "current dev or merge head was offered as a substitute for the trigger subject", true, true)
	}
	if mismatch := triggerMismatch(input); mismatch != "" {
		return refuted(StateMismatch, mismatch, true, false)
	}
	if input.Source.Conclusion == "failure" {
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
		return Evaluation{Decision: DecisionUnknown, LineageState: state, Reason: reason, ExactSubjectBinding: true, Unknown: unknown(class, reason)}
	}
	if artifact.RunID != input.Source.ID || artifact.Name != input.ExpectedArtifactName || artifact.SubjectSHA != "" && artifact.SubjectSHA != input.Trigger.SourceSubjectSHA || artifact.Digest == "" || input.ExpectedDigest != "" && artifact.Digest != input.ExpectedDigest || artifact.PayloadDigest != "" && artifact.PayloadDigest != artifact.Digest {
		return refuted(StateTampered, "source artifact subject or digest does not match the exact trigger lineage", true, false)
	}
	return Evaluation{Decision: DecisionClosed, LineageState: StateExact, Reason: "exact workflow, run attempt, subject, artifact, and digest matched", ExactSubjectBinding: true, ArtifactResolved: true, ArtifactIdentity: fmt.Sprintf("%d:%s:%s", artifact.ID, artifact.Name, artifact.Digest)}
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
	if trigger.SourceRef != "" && source.Ref != "" && trigger.SourceRef != source.Ref {
		return "source ref does not match the triggering workflow ref"
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
	return &CausalUnknown{Stage: "workflow-lineage", Step: "source-artifact-resolution", Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: []string{"source-run-artifact", "subject-sha", "run-attempt"}}
}

func refuted(state, reason string, mismatch, fallback bool) Evaluation {
	return Evaluation{Decision: DecisionRefuted, LineageState: state, Reason: reason, MismatchDetected: mismatch, FallbackAttempted: fallback, FallbackRejected: fallback}
}
