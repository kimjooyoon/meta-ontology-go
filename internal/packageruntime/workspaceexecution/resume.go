package workspaceexecution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type ContinuationEvidence struct {
	Schema            string `json:"schema"`
	PriorResultSHA256 string `json:"prior_result_sha256"`
	BodyFillsReplayed int    `json:"body_fills_replayed"`
	SavedPolicyStages int    `json:"saved_policy_stages"`
	NewModelCalls     int    `json:"new_model_calls"`
	Scope             string `json:"scope"`
}

// ResumeWorkspace keeps the original program and attempted candidate identities,
// applies the explicitly supplied policy, then observes two new native runs.
// Earlier body fills are replayed unchanged. This API accepts no model/provider.
func ResumeWorkspace(ctx context.Context, manifest packageruntime.Manifest, prior Result,
	suite bodyexecution.CompositionCases, policyManifest *packageruntime.Manifest, goBinary string) (Result, error) {
	program, current, err := reconstructWorkspace(ctx, manifest, prior)
	if err != nil {
		return Result{}, err
	}
	translated, err := translateCases(program, suite)
	if err != nil {
		return Result{}, err
	}
	if policyManifest == nil || len(prior.PolicyHistory) >= 16 {
		return Result{}, fmt.Errorf("workspace continuation requires an explicit policy and at most 16 saved stages")
	}
	snapshot, policy, err := prepareWorkspacePolicy(ctx, policyManifest)
	if err != nil {
		return Result{}, err
	}
	composition, err := bodyexecution.ResumeComposition(ctx, "workspace.gooo", current, prior.Composition, translated, *policy)
	if err != nil {
		return Result{}, fmt.Errorf("continue workspace construction: %w", err)
	}
	result, err := resumedWorkspaceResult(program, prior, snapshot, composition)
	if err != nil {
		return Result{}, err
	}
	result.Runtime, err = bodyexecution.ExecuteComposition(ctx, "workspace.gooo", current, composition, translated, goBinary)
	if err == nil {
		result.Runtime.InputSeparation = measureWorkspaceInputs(ctx, current, result, translated)
	}
	return result, err
}

func resumedWorkspaceResult(program Program, prior Result, policy *WorkspaceAssemblyPolicy,
	composition bodyexecution.Composition) (Result, error) {
	encoded, err := json.Marshal(prior)
	if err != nil {
		return Result{}, err
	}
	// Preserve each policy's source packages, including a nil first policy for
	// construction that originally used the built-in deterministic stopping rule.
	stages := append(append([]*WorkspaceAssemblyPolicy(nil), prior.PolicyHistory...), prior.AssemblyPolicy)
	raw, err := json.Marshal(stages)
	if err != nil {
		return Result{}, err
	}
	var history []*WorkspaceAssemblyPolicy
	if err := json.Unmarshal(raw, &history); err != nil {
		return Result{}, err
	}
	return Result{Schema: prior.Schema, Program: program, BodyFills: prior.BodyFills, SourceSHA256: prior.SourceSHA256,
		Composition: composition, AssemblyPolicy: policy, PolicyHistory: history,
		Continuation: &ContinuationEvidence{Schema: "gooo/workspace-body-continuation/v1", PriorResultSHA256: sourceSHA256(encoded),
			BodyFillsReplayed: len(prior.BodyFills), SavedPolicyStages: len(history),
			Scope: "source and historical policy packages reconstructed; retained and rechecked candidates keep the original cumulative budget; zero new predictions"},
		Scope: "continued package record construction with two fresh native executions; finite expectations or input-only observations"}, nil
}
