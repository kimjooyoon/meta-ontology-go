package guardedpromotion

import (
	"context"
	"fmt"
)

func (collector *Collector) collectRepository(
	ctx context.Context, repositoryPath string, source *Source,
) error {
	var response repositoryResponse
	if err := collector.getJSON(ctx, repositoryPath, &response); err != nil {
		return err
	}
	source.ObservedRepository = response.FullName
	if response.FullName != source.RequestedRepository || response.DefaultBranch == "" {
		return fmt.Errorf("repository identity is not exact")
	}
	source.DefaultBranch = response.DefaultBranch
	return nil
}

func (collector *Collector) collectWorkflow(
	ctx context.Context, repositoryPath string, source *Source, runID int64,
) error {
	var response workflowRunResponse
	path := fmt.Sprintf("%s/actions/runs/%d", repositoryPath, runID)
	if err := collector.getJSON(ctx, path, &response); err != nil {
		return err
	}
	if response.ID != runID {
		return fmt.Errorf("source workflow run identity is not exact")
	}
	source.Workflow = WorkflowEvidence{
		RunID: response.ID, Name: response.Name, Path: response.Path,
		Event: response.Event, Status: response.Status, Conclusion: response.Conclusion,
		HeadSHA: response.HeadSHA, HeadBranch: response.HeadBranch,
	}
	return nil
}
