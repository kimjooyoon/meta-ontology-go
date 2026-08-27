package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func collect(ctx context.Context, client *githubClient, cfg config, predecessor string) (predecessorselection.Input, error) {
	runs, err := collectWorkflowRuns(ctx, client, cfg, predecessor)
	if err != nil {
		return predecessorselection.Input{}, err
	}
	input := predecessorselection.Input{Repository: cfg.repository,
		CurrentHeadSHA: cfg.currentHead, PredecessorSHA: predecessor,
		Branch: cfg.branch, Workflow: cfg.workflow}
	for _, run := range runs {
		candidate, ok, err := collectRun(ctx, client, cfg, predecessor, run)
		if err != nil {
			return predecessorselection.Input{}, err
		}
		if ok {
			input.Candidates = append(input.Candidates, candidate)
		}
	}
	sort.Slice(input.Candidates, func(i, j int) bool {
		return input.Candidates[i].RunID < input.Candidates[j].RunID
	})
	return input, nil
}

func encodedPayload(ctx context.Context, client *githubClient, repository string,
	metadata artifactMetadata, filename string) (string, error) {
	if metadata.Expired {
		return "", nil
	}
	archive, err := client.get(ctx, fmt.Sprintf("/repos/%s/actions/artifacts/%d/zip",
		repository, metadata.ID))
	if err != nil {
		return "", err
	}
	payload, err := verifiedPayload(archive, filename)
	return base64.StdEncoding.EncodeToString(payload), err
}
