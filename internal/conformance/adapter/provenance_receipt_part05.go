package adapter

import (
	"encoding/hex"
	"fmt"
)

func validateReceiptJobs(jobs []ReceiptJob, head string, status ReceiptProvenanceStatus) error {
	if len(jobs) != len(receiptJobNames) {
		return fmt.Errorf("receipt requires exactly six jobs")
	}
	seen := make(map[string]struct{}, len(jobs))
	for index, job := range jobs {
		if job.Name != receiptJobNames[index] || job.HeadSHA != head {
			return fmt.Errorf("receipt job %q is not bound to head", job.Name)
		}
		if _, exists := seen[job.Name]; exists {
			return fmt.Errorf("receipt job %q is duplicated", job.Name)
		}
		if !validCommitID(job.HeadSHA) || !validJobStatus(job.Status) {
			return fmt.Errorf("receipt job %q is malformed", job.Name)
		}
		if (job.Status == "completed" && !validJobConclusion(job.Conclusion)) ||
			(job.Status != "completed" && job.Conclusion != "") {
			return fmt.Errorf("receipt job %q has an invalid conclusion state", job.Name)
		}
		seen[job.Name] = struct{}{}
	}
	if status != ReceiptProvenanceVerified && status != ReceiptProvenanceDeferred {
		return fmt.Errorf("unsupported receipt provenance_status %q", status)
	}
	if status == ReceiptProvenanceVerified {
		for _, job := range jobs {
			if job.Status != "completed" || job.Conclusion != "success" {
				return fmt.Errorf("verified receipt has a non-success job %q", job.Name)
			}
		}
	}
	return nil
}
func validJobStatus(status string) bool {
	return status == "queued" || status == "in_progress" || status == "completed"
}
func validJobConclusion(conclusion string) bool {
	return conclusion == "success" || conclusion == "failure" ||
		conclusion == "cancelled" || conclusion == "skipped" || conclusion == "neutral"
}
func validCommitID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
