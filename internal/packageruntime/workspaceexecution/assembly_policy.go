package workspaceexecution

import (
	"bytes"
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
	var policy *bodycodegen.RecordAssemblyPolicy
	if prior.AssemblyPolicy != nil {
		snapshot, prepared, err := prepareWorkspacePolicy(ctx, &prior.AssemblyPolicy.Manifest)
		if err != nil {
			return err
		}
		expected, _ := json.Marshal(snapshot)
		observed, _ := json.Marshal(prior.AssemblyPolicy)
		if !bytes.Equal(expected, observed) {
			return fmt.Errorf("saved assembly policy packages or lowering differ")
		}
		policy = prepared
	}
	return verifyWorkspacePolicySteps(prior.Composition, policy)
}

func verifyWorkspacePolicySteps(composition bodyexecution.Composition, policy *bodycodegen.RecordAssemblyPolicy) error {
	controlled := 0
	for _, step := range composition.Steps {
		record := step.Generation.Report.RecordAssembly
		if record == nil {
			continue
		}
		if policy == nil {
			if record.Control != nil || len(record.ControlHistory) != 0 {
				return fmt.Errorf("saved workspace control requires its policy source packages")
			}
			continue
		}
		if record.Control == nil || record.Control.Policy != *policy || len(record.ControlHistory) != 0 {
			return fmt.Errorf("saved workspace record selection differs from its assembly policy")
		}
		controlled++
	}
	if policy != nil && controlled == 0 {
		return fmt.Errorf("saved assembly policy has no record-choice construction")
	}
	return nil
}
