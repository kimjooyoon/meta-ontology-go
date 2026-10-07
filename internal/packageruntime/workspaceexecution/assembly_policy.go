package workspaceexecution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

// WorkspaceAssemblyPolicy retains the policy's source packages as well as their
// lowering. Replay can reconstruct imported helpers without opening old paths.
type WorkspaceAssemblyPolicy struct {
	Schema   string                  `json:"schema"`
	Manifest packageruntime.Manifest `json:"source_manifest"`
	Program  Program                 `json:"program"`
}

func prepareWorkspacePolicy(ctx context.Context, manifest *packageruntime.Manifest) (*WorkspaceAssemblyPolicy, *bodycodegen.RecordAssemblyPolicy, error) {
	if manifest == nil {
		return nil, nil, nil
	}
	raw, err := json.Marshal(manifest)
	if err != nil || len(raw) > 128<<10 {
		return nil, nil, fmt.Errorf("assembly policy source manifest exceeds 128 KiB or cannot be encoded")
	}
	var owned packageruntime.Manifest
	if err := json.Unmarshal(raw, &owned); err != nil {
		return nil, nil, err
	}
	program, err := Prepare(owned)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare assembly policy workspace: %w", err)
	}
	if len(program.Activities) != 1 {
		return nil, nil, fmt.Errorf("assembly policy requires one entry with pure helper calls; bound producers are not policy inputs")
	}
	policy := &bodycodegen.RecordAssemblyPolicy{Source: program.Source, Activity: program.Entry.LoweredName}
	if err := bodycodegen.ValidateRecordAssemblyPolicy(ctx, *policy); err != nil {
		return nil, nil, err
	}
	return &WorkspaceAssemblyPolicy{Schema: "gooo/workspace-assembly-policy/v1", Manifest: owned, Program: program}, policy, nil
}

func verifyWorkspacePolicy(ctx context.Context, prior Result) error {
	if err := verifyWorkspaceContinuation(prior); err != nil {
		return err
	}
	policy, err := verifiedWorkspacePolicySnapshot(ctx, prior.AssemblyPolicy)
	if err != nil {
		return err
	}
	history := make([]*bodycodegen.RecordAssemblyPolicy, len(prior.PolicyHistory))
	for i, snapshot := range prior.PolicyHistory {
		history[i], err = verifiedWorkspacePolicySnapshot(ctx, snapshot)
		if err != nil {
			return fmt.Errorf("workspace policy stage %d: %w", i, err)
		}
	}
	return verifyWorkspacePolicySteps(prior.Composition, policy, history)
}

func verifyWorkspacePolicySteps(composition bodyexecution.Composition, policy *bodycodegen.RecordAssemblyPolicy,
	history []*bodycodegen.RecordAssemblyPolicy) error {
	controlled := 0
	for _, step := range composition.ConstructionSteps() {
		record := step.Generation.Report.RecordAssembly
		if record == nil {
			continue
		}
		if len(record.ControlHistory) != len(history) || !sameWorkspaceRecordPolicy(record.Control, policy) {
			return fmt.Errorf("saved workspace record selection differs from its assembly policy")
		}
		for i, stage := range record.ControlHistory {
			if !sameWorkspaceRecordPolicy(stage.Control, history[i]) {
				return fmt.Errorf("saved workspace record differs from policy stage %d", i)
			}
		}
		controlled++
	}
	if (policy != nil || len(history) > 0) && controlled == 0 {
		return fmt.Errorf("saved assembly policy has no record-choice construction")
	}
	return nil
}
