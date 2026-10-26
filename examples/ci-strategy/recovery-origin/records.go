package recoveryorigin

import (
	"fmt"
	"time"
)

type runRecord struct {
	ID               int64
	RunAttempt       int64  `json:"run_attempt"`
	HeadSHA          string `json:"head_sha"`
	Path, Conclusion string
	UpdatedAt        time.Time `json:"updated_at"`
	Repository       struct {
		FullName string `json:"full_name"`
	}
}

type jobRecords struct {
	Jobs []struct {
		ID         int64
		RunID      int64  `json:"run_id"`
		RunAttempt int64  `json:"run_attempt"`
		HeadSHA    string `json:"head_sha"`
		Steps      []struct {
			Number     int64
			Conclusion string
		}
	}
}

type artifactRecord struct {
	Name, Digest string
	WorkflowRun  struct {
		ID      int64
		HeadSHA string `json:"head_sha"`
	} `json:"workflow_run"`
}

func copyOrigin(run runRecord, jobs jobRecords, artifact artifactRecord, jobID int64, out *R_ReceiptOrigin) error {
	out.F_repository, out.F_source = run.Repository.FullName, run.HeadSHA
	out.F_run_id, out.F_attempt = run.ID, run.RunAttempt
	out.F_workflow_path, out.F_conclusion = run.Path, run.Conclusion
	out.F_plan_reference = fmt.Sprintf("https://github.com/%s/blob/%s/%s", run.Repository.FullName, run.HeadSHA, run.Path)
	out.F_reference = fmt.Sprintf("https://github.com/%s/actions/runs/%d/job/%d", run.Repository.FullName, run.ID, jobID)
	if !run.UpdatedAt.IsZero() {
		out.F_completed_at = run.UpdatedAt.Unix()
	}
	out.F_artifact_run_id, out.F_artifact_source = artifact.WorkflowRun.ID, artifact.WorkflowRun.HeadSHA
	out.F_artifact_sha256 = artifact.Digest
	out.F_artifact_name = artifact.Name
	found := 0
	for _, job := range jobs.Jobs {
		if job.ID != jobID {
			continue
		}
		found++
		out.F_job_id, out.F_job_run_id = job.ID, job.RunID
		out.F_job_attempt, out.F_job_source = job.RunAttempt, job.HeadSHA
		for _, step := range job.Steps {
			out.F_step_numbers = append(out.F_step_numbers, step.Number)
			out.F_step_conclusions = append(out.F_step_conclusions, step.Conclusion)
			out.F_step_references = append(out.F_step_references, fmt.Sprintf("%s#step:%d:1", out.F_reference, step.Number))
		}
	}
	if found != 1 {
		return fmt.Errorf("job %d: expected one record, found %d", jobID, found)
	}
	return nil
}
