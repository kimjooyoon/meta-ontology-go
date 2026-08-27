package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func collectRun(ctx context.Context, client *githubClient, cfg config, predecessor string,
	run workflowRun) (predecessorselection.Candidate, bool, []predecessorselection.PaginationPage, string) {
	if run.HeadSHA != predecessor || run.Status != "completed" {
		return predecessorselection.Candidate{}, false, nil, ""
	}
	endpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/artifacts?per_page=100",
		cfg.repository, run.ID)
	artifacts, pages, failureReason := collectArtifacts(ctx, client, endpoint)
	if failureReason != "" {
		return predecessorselection.Candidate{}, false, pages, failureReason
	}
	readinessName := "language-readiness-artifact-" + predecessor
	bindingName := "language-readiness-predecessor-binding-" + predecessor
	readiness, readinessOK := findArtifact(artifacts, readinessName)
	binding, bindingOK := findArtifact(artifacts, bindingName)
	if !readinessOK || !bindingOK {
		return predecessorselection.Candidate{}, false, pages, ""
	}
	jobsEndpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/jobs?filter=latest&per_page=100",
		cfg.repository, run.ID)
	jobs, jobPages, failureReason := collectJobs(ctx, client, jobsEndpoint)
	pages = append(pages, jobPages...)
	if failureReason != "" {
		return predecessorselection.Candidate{}, false, pages, failureReason
	}
	producer, producerMatches := findProducerJob(jobs, predecessorselection.ProducerJobName)
	readinessPayload, err := encodedPayload(ctx, client, cfg.repository, readiness, "artifact.json")
	if err != nil {
		return predecessorselection.Candidate{}, false, pages, "ARTIFACT_PAYLOAD_UNAVAILABLE"
	}
	bindingPayload, err := encodedPayload(ctx, client, cfg.repository, binding,
		"language-readiness-predecessor-binding-a.json")
	if err != nil {
		return predecessorselection.Candidate{}, false, pages, "ARTIFACT_PAYLOAD_UNAVAILABLE"
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
		BindingExpired: binding.Expired, BindingPayloadBase64: bindingPayload}, true, pages, ""
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
