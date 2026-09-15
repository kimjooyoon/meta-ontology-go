package guardedpromotion

import (
	"context"
	"fmt"
	"net/url"
)

type promotionCandidate struct {
	run      workflowRunResponse
	artifact artifactResponse
}

func (collector *Collector) collectPromotion(
	ctx context.Context, repositoryPath string, source *Source,
) error {
	runs, err := collector.promotionRuns(ctx, repositoryPath, source.PredecessorSHA)
	if err != nil {
		return err
	}
	source.ObservedRuns = len(runs)
	candidates := make([]promotionCandidate, 0, 1)
	for _, run := range runs {
		artifacts, listErr := collector.runArtifacts(ctx, repositoryPath, run.ID)
		if listErr != nil {
			return listErr
		}
		source.ObservedArtifacts += len(artifacts)
		for _, artifact := range artifacts {
			if !artifact.Expired && artifact.Name == PromotionArtifactBase+source.PredecessorSHA {
				candidates = append(candidates, promotionCandidate{run: run, artifact: artifact})
			}
		}
	}
	source.ValidCandidates = len(candidates)
	if len(candidates) == 0 {
		source.UnresolvedCandidates = 1
		return nil
	}
	if len(candidates) > 1 {
		source.AmbiguousCandidates = len(candidates)
		return nil
	}
	return collector.bindCandidate(ctx, repositoryPath, source, candidates[0])
}

func (collector *Collector) promotionRuns(
	ctx context.Context, repositoryPath, predecessorSHA string,
) ([]workflowRunResponse, error) {
	workflow := url.PathEscape(TransformationPath)
	query := url.Values{"head_sha": {predecessorSHA}, "status": {"completed"}, "per_page": {"100"}}
	path := fmt.Sprintf("%s/actions/workflows/%s/runs?%s", repositoryPath, workflow, query.Encode())
	var response workflowRunsResponse
	if err := collector.getJSON(ctx, path, &response); err != nil {
		return nil, err
	}
	valid := make([]workflowRunResponse, 0, len(response.WorkflowRuns))
	for _, run := range response.WorkflowRuns {
		if run.HeadSHA == predecessorSHA && run.Status == "completed" && run.Conclusion == "success" {
			valid = append(valid, run)
		}
	}
	return valid, nil
}
