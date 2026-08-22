package guardedpromotion

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (collector *Collector) Collect(
	ctx context.Context, repository, currentHeadSHA string, sourceRunID int64,
) Source {
	source := Source{RequestedRepository: repository, CurrentHeadSHA: currentHeadSHA}
	if err := collector.collect(ctx, &source, sourceRunID); err != nil {
		source.CollectionError = err.Error()
		if source.UnresolvedCandidates == 0 {
			source.UnresolvedCandidates = 1
		}
	}
	return source
}

func (collector *Collector) collect(ctx context.Context, source *Source, runID int64) error {
	repositoryPath, err := repositoryAPIPath(source.RequestedRepository)
	if err != nil {
		return err
	}
	if err := collector.collectRepository(ctx, repositoryPath, source); err != nil {
		return err
	}
	if err := collector.collectWorkflow(ctx, repositoryPath, source, runID); err != nil {
		return err
	}
	if err := collector.collectPredecessor(ctx, repositoryPath, source); err != nil {
		return err
	}
	return collector.collectPromotion(ctx, repositoryPath, source)
}

func repositoryAPIPath(repository string) (string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("repository %q is not owner/name", repository)
	}
	return "/repos/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]), nil
}
