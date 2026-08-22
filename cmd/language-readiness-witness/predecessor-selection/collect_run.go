package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func collectRun(ctx context.Context, client *githubClient, cfg config, predecessor string,
	run workflowRun) (predecessorselection.Candidate, bool, error) {
	if run.HeadSHA != predecessor || run.Status != "completed" {
		return predecessorselection.Candidate{}, false, nil
	}
	var artifacts artifactList
	endpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/artifacts?per_page=100",
		cfg.repository, run.ID)
	if err := client.getJSON(ctx, endpoint, &artifacts); err != nil {
		return predecessorselection.Candidate{}, false, err
	}
	if artifacts.TotalCount != len(artifacts.Artifacts) {
		return predecessorselection.Candidate{}, false,
			fmt.Errorf("artifact pagination incomplete")
	}
	readinessName := "language-readiness-artifact-" + predecessor
	bindingName := "language-readiness-predecessor-binding-" + predecessor
	readiness, readinessOK := findArtifact(artifacts.Artifacts, readinessName)
	binding, bindingOK := findArtifact(artifacts.Artifacts, bindingName)
	if !readinessOK || !bindingOK {
		return predecessorselection.Candidate{}, false, nil
	}
	var jobs workflowJobList
	jobsEndpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/jobs?filter=latest&per_page=100",
		cfg.repository, run.ID)
	if err := client.getJSON(ctx, jobsEndpoint, &jobs); err != nil {
		return predecessorselection.Candidate{}, false, err
	}
	if jobs.TotalCount != len(jobs.Jobs) {
		return predecessorselection.Candidate{}, false, fmt.Errorf("job pagination incomplete")
	}
	producer, producerMatches := findProducerJob(jobs.Jobs, predecessorselection.ProducerJobName)
	readinessPayload, err := encodedPayload(ctx, client, cfg.repository, readiness, "artifact.json")
	if err != nil {
		return predecessorselection.Candidate{}, false, err
	}
	bindingPayload, err := encodedPayload(ctx, client, cfg.repository, binding,
		"language-readiness-predecessor-binding-a.json")
	if err != nil {
		return predecessorselection.Candidate{}, false, err
	}
	return predecessorselection.Candidate{RunID: run.ID, RunAttempt: run.RunAttempt,
		Workflow: run.Name, HeadBranch: run.HeadBranch, HeadSHA: run.HeadSHA,
		Event: run.Event, Conclusion: run.Conclusion,
		ProducerJobID: producer.ID, ProducerJobRunAttempt: producer.RunAttempt,
		ProducerJobName: producer.Name, ProducerJobStatus: producer.Status,
		ProducerJobConclusion: producer.Conclusion, ProducerJobMatches: producerMatches,
		ReadinessArtifactID: readiness.ID, ReadinessArtifactName: readiness.Name,
		ReadinessExpired: readiness.Expired, ReadinessPayloadBase64: readinessPayload,
		BindingArtifactID: binding.ID, BindingArtifactName: binding.Name,
		BindingExpired: binding.Expired, BindingPayloadBase64: bindingPayload}, true, nil
}

func findProducerJob(values []workflowJob, name string) (workflowJob, int) {
	var result workflowJob
	count := 0
	for _, value := range values {
		if value.Name == name {
			result, count = value, count+1
		}
	}
	return result, count
}

func findArtifact(values []artifactMetadata, name string) (artifactMetadata, bool) {
	var result artifactMetadata
	count := 0
	for _, value := range values {
		if value.Name == name {
			result, count = value, count+1
		}
	}
	return result, count == 1
}
