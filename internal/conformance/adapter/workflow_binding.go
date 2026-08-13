package adapter

import (
	"fmt"
	"strings"
)

// WorkflowEvidenceStatus distinguishes absent, unverified, and verified CI data.
type WorkflowEvidenceStatus string

const (
	WorkflowEvidenceMissing    WorkflowEvidenceStatus = "missing"
	WorkflowEvidenceUnverified WorkflowEvidenceStatus = "unverified"
	WorkflowEvidenceVerified   WorkflowEvidenceStatus = "verified"
)

// WorkflowBinding is captured by the observer outside the producer response.
type WorkflowBinding struct {
	Status        WorkflowEvidenceStatus `json:"status"`
	Repository    string                 `json:"repository,omitempty"`
	BaseSHA       string                 `json:"base_sha,omitempty"`
	HeadSHA       string                 `json:"head_sha,omitempty"`
	EventRef      string                 `json:"event_ref,omitempty"`
	CheckoutRef   string                 `json:"checkout_ref,omitempty"`
	Run           string                 `json:"run,omitempty"`
	Attempt       int                    `json:"attempt,omitempty"`
	ArtifactCount int                    `json:"artifact_count"`
	Jobs          []ReceiptJob           `json:"jobs,omitempty"`
}

func missingWorkflowBinding() WorkflowBinding {
	return WorkflowBinding{Status: WorkflowEvidenceMissing}
}

func (w WorkflowBinding) clone() WorkflowBinding {
	w.Jobs = append([]ReceiptJob{}, w.Jobs...)
	return w
}

func (w WorkflowBinding) validate() error {
	switch w.Status {
	case WorkflowEvidenceMissing:
		if !workflowFieldsEmpty(w) {
			return fmt.Errorf("missing workflow evidence has fields")
		}
	case WorkflowEvidenceUnverified:
		if w.ArtifactCount < 0 {
			return fmt.Errorf("workflow artifact_count cannot be negative")
		}
	case WorkflowEvidenceVerified:
		return w.validateVerified()
	default:
		return fmt.Errorf("unsupported workflow evidence status %q", w.Status)
	}
	return nil
}

type verifiedWorkflowEvidence struct {
	binding WorkflowBinding
}

func newVerifiedWorkflowEvidence(workflow WorkflowBinding) (verifiedWorkflowEvidence, error) {
	if workflow.Status != WorkflowEvidenceVerified {
		return verifiedWorkflowEvidence{}, fmt.Errorf("verified workflow status is required")
	}
	if err := workflow.validate(); err != nil {
		return verifiedWorkflowEvidence{}, err
	}
	return verifiedWorkflowEvidence{binding: workflow.clone()}, nil
}

func (w WorkflowBinding) validateVerified() error {
	for name, value := range map[string]string{
		"repository": w.Repository, "event_ref": w.EventRef,
		"checkout_ref": w.CheckoutRef, "run": w.Run,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("workflow %s is required", name)
		}
	}
	if !validCommitID(w.BaseSHA) || !validCommitID(w.HeadSHA) {
		return fmt.Errorf("workflow base_sha and head_sha are invalid")
	}
	if w.Attempt < 1 || w.ArtifactCount < 1 {
		return fmt.Errorf("verified workflow requires attempt and artifact")
	}
	return validateReceiptJobs(w.Jobs, w.HeadSHA, ReceiptProvenanceVerified)
}

func workflowFieldsEmpty(w WorkflowBinding) bool {
	return w.Repository == "" && w.BaseSHA == "" && w.HeadSHA == "" &&
		w.EventRef == "" && w.CheckoutRef == "" && w.Run == "" &&
		w.Attempt == 0 && w.ArtifactCount == 0 && len(w.Jobs) == 0
}
