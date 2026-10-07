package workspaceexecution

import (
	"context"
	"strings"
	"testing"
)

func TestWorkspaceContinuationRejectsUnboundPolicyHistory(t *testing.T) {
	m, suite, checkpoint, policy := continuationFixture(t)
	ctx := context.Background()
	prior, err := ExecuteWorkspaceWithOptions(ctx, m, suite, ExecuteOptions{AssemblyPolicy: &checkpoint})
	if err != nil {
		t.Fatal(err)
	}
	next, err := ResumeWorkspace(ctx, m, savedWorkspaceResult(t, prior), suite, &policy, "")
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*Result){
		"missing history": func(r *Result) { r.PolicyHistory = nil },
		"duplicate history": func(r *Result) {
			r.PolicyHistory = append(r.PolicyHistory, r.PolicyHistory[0])
			r.Continuation.SavedPolicyStages++
		},
		"different historical policy": func(r *Result) { r.PolicyHistory[0] = r.AssemblyPolicy },
		"missing historical policy":   func(r *Result) { r.PolicyHistory[0] = nil },
		"source packages": func(r *Result) {
			r.PolicyHistory[0].Manifest.Packages[0].Sources[0].Content += "\n"
		},
		"historical lowering":  func(r *Result) { r.PolicyHistory[0].Program.Source += "\n" },
		"current policy":       func(r *Result) { r.AssemblyPolicy = nil },
		"model count":          func(r *Result) { r.Continuation.NewModelCalls = 1 },
		"fill count":           func(r *Result) { r.Continuation.BodyFillsReplayed = 1 },
		"stage count":          func(r *Result) { r.Continuation.SavedPolicyStages = 3 },
		"schema":               func(r *Result) { r.Continuation.Schema = "other" },
		"parent":               func(r *Result) { r.Continuation.PriorResultSHA256 = "missing" },
		"missing continuation": func(r *Result) { r.Continuation = nil },
		"composition history":  func(r *Result) { r.Composition.Continuation = nil },
		"historical control": func(r *Result) {
			r.Composition.Preparations[0].Generation.Report.RecordAssembly.ControlHistory[0].Control.Policy.Source += "\n"
		},
		"history bound": func(r *Result) {
			for len(r.PolicyHistory) < 17 {
				r.PolicyHistory = append(r.PolicyHistory, r.AssemblyPolicy)
			}
			r.Continuation.SavedPolicyStages = 17
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			bad := savedWorkspaceResult(t, next)
			mutate(&bad)
			_, err := ReplayWorkspace(ctx, m, bad, suite, "missing-go-binary")
			if err == nil || strings.Contains(err.Error(), "missing-go-binary") {
				t.Fatal("altered continuation reached native execution", err)
			}
			_, err = ResumeWorkspace(ctx, m, bad, suite, &policy, "missing-go-binary")
			if err == nil || strings.Contains(err.Error(), "missing-go-binary") {
				t.Fatal("altered continuation reached another round", err)
			}
		})
	}
	m.Packages[0].Sources[0].Content += "\n"
	if _, err := ResumeWorkspace(ctx, m, next, suite, &policy, "missing-go-binary"); err == nil || strings.Contains(err.Error(), "missing-go-binary") {
		t.Fatal("changed source reached native execution", err)
	}
}
