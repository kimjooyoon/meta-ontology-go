package adapter

import (
	"fmt"
	"strings"
)

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
