package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func loadInput(settings options, program grant.PolicyProgram) (grant.GrantInput, error) {
	v24Binding := grant.V24Binding{}
	if settings.v24RequestPath != "" && settings.v24ResolutionPath != "" {
		var request selfimprovementcandidate.AuthorizationRequest
		var resolution selfimprovementcandidate.AuthorizationResolution
		if err := readJSON(settings.v24RequestPath, &request); err == nil {
			if err := readJSON(settings.v24ResolutionPath, &resolution); err == nil {
				v24Binding = grant.ProjectV24(request, resolution)
			}
		}
	}
	v25Binding := grant.V25Binding{}
	if settings.v25ContractPath != "" {
		var report v25.LiveReport
		if err := readJSON(settings.v25ContractPath, &report); err == nil {
			v25Binding = grant.ProjectV25(report.ContractResolution)
		} else {
			var resolution v25.ContractResolution
			if err := readJSON(settings.v25ContractPath, &resolution); err == nil {
				v25Binding = grant.ProjectV25(resolution)
			}
		}
	}
	source := grant.SourceArtifact{}
	if settings.sourcePath != "" {
		_ = readJSON(settings.sourcePath, &source)
	}
	request := grant.BuildRequest(program, v24Binding, v25Binding, source)
	input := grant.GrantInput{Request: request, Live: settings.mode == "live"}
	if settings.decision != "" {
		input.DecisionInputs = []grant.GrantDecisionInput{grant.BuildDecisionInput(request, settings.decision, settings.decisionSource, grant.ActorEvidence{Repository: settings.repository, Actor: settings.actor, WorkflowRunID: settings.workflowRunID, WorkflowRunAttempt: settings.workflowRunAttempt, Event: settings.event, EvidenceLabel: grant.ActorEvidenceLabel})}
	}
	return input, nil
}
