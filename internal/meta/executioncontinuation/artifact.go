package executioncontinuation

import (
	"context"
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
)

type Run struct {
	candidatedelivery.Run
	Branch string `json:"head_branch"`
}

func (c Client) observeRun(ctx context.Context, kind string, id int64, attempt int, complete bool) (Run, error) {
	for {
		var run Run
		if err := c.get(ctx, fmt.Sprintf("actions/runs/%d", id), &run); err != nil {
			return run, unknown("SOURCE_RUN_UNAVAILABLE")
		}
		if err := c.runIdentity(run, kind, id, attempt); err != nil {
			return run, err
		}
		if run.Status == "completed" {
			if run.Conclusion != "success" {
				return run, refuted("SOURCE_RUN_NOT_SUCCESSFUL")
			}
			return run, nil
		}
		if !complete {
			return run, nil
		}
		select {
		case <-ctx.Done():
			return run, unknown("SOURCE_RUN_COMPLETION_UNKNOWN")
		case <-time.After(2 * time.Second):
		}
	}
}

func (c Client) runIdentity(r Run, kind string, id int64, attempt int) error {
	if attempt != 1 || (r.ID != 0 && r.ID != id) || (r.Attempt != 0 && r.Attempt != attempt) ||
		(r.Repository.Name != "" && r.Repository.Name != c.Repository) ||
		(r.Path != "" && r.Path != policy.WorkflowPath(kind)) ||
		(r.Event != "" && r.Event != "workflow_dispatch") || (r.Branch != "" && r.Branch != "dev") ||
		(r.Head != "" && !validSHA(r.Head)) {
		return refuted("SOURCE_RUN_MISMATCH")
	}
	if r.ID == 0 || r.Attempt == 0 || r.Repository.Name == "" || r.Path == "" ||
		r.Event == "" || r.Branch == "" || r.Head == "" || r.Status == "" {
		return unknown("SOURCE_RUN_METADATA_UNKNOWN")
	}
	return nil
}

func (c Client) artifacts(ctx context.Context, run int64) ([]candidatedelivery.Artifact, error) {
	var all []candidatedelivery.Artifact
	total := -1
	for page := 1; page <= 8; page++ {
		var part struct {
			Total     *int                         `json:"total_count"`
			Artifacts []candidatedelivery.Artifact `json:"artifacts"`
		}
		if err := c.get(ctx, fmt.Sprintf("actions/runs/%d/artifacts?per_page=100&page=%d", run, page), &part); err != nil {
			return nil, unknown("ARTIFACT_POPULATION_UNKNOWN")
		}
		if part.Total == nil || part.Artifacts == nil || *part.Total < 0 || len(part.Artifacts) > 100 ||
			(total >= 0 && total != *part.Total) {
			return nil, unknown("ARTIFACT_POPULATION_UNKNOWN")
		}
		total = *part.Total
		all = append(all, part.Artifacts...)
		if len(all) == total {
			return all, nil
		}
		if len(all) > total {
			return nil, refuted("ARTIFACT_POPULATION_MISMATCH")
		}
	}
	return nil, unknown("ARTIFACT_POPULATION_INCOMPLETE")
}

func (c Client) readArtifact(ctx context.Context, run Run, name string, id int64, digest string, wanted []string) (map[string][]byte, error) {
	artifacts, err := c.artifacts(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	var selected []candidatedelivery.Artifact
	for _, a := range artifacts {
		if a.Name == name {
			selected = append(selected, a)
		}
	}
	if len(selected) == 0 {
		return nil, unknown("SOURCE_ARTIFACT_MISSING")
	}
	if len(selected) != 1 {
		return nil, refuted("SOURCE_ARTIFACT_DUPLICATED")
	}
	a := selected[0]
	if a.ID != id || a.Digest != digest || !validDigest(a.Digest) || a.Size <= 0 || a.Size > MaxBytes || a.Run.ID != run.ID || a.Run.Head != run.Head {
		return nil, refuted("SOURCE_ARTIFACT_MISMATCH")
	}
	if a.Expired {
		return nil, unknown("SOURCE_ARTIFACT_EXPIRED")
	}
	raw, err := c.request(ctx, "GET", fmt.Sprintf("actions/artifacts/%d/zip", id), nil)
	if err != nil {
		return nil, unknown("SOURCE_ARCHIVE_UNAVAILABLE")
	}
	if int64(len(raw)) != a.Size || Digest(raw) != a.Digest {
		return nil, refuted("SOURCE_ARCHIVE_MISMATCH")
	}
	files, err := candidatedelivery.ReadArchiveMembers(raw, wanted)
	if err != nil {
		return nil, refuted("SOURCE_ARCHIVE_INVALID")
	}
	for _, name := range wanted {
		if len(files[name]) == 0 {
			return nil, unknown("SOURCE_MEMBER_MISSING")
		}
	}
	return files, nil
}
