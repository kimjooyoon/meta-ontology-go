package workspaceexecution

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func verifiedWorkspacePolicySnapshot(ctx context.Context, saved *WorkspaceAssemblyPolicy) (*bodycodegen.RecordAssemblyPolicy, error) {
	if saved == nil {
		return nil, nil
	}
	snapshot, policy, err := prepareWorkspacePolicy(ctx, &saved.Manifest)
	if err != nil {
		return nil, err
	}
	expected, _ := json.Marshal(snapshot)
	observed, _ := json.Marshal(saved)
	if !bytes.Equal(expected, observed) {
		return nil, fmt.Errorf("saved assembly policy packages or lowering differ")
	}
	return policy, nil
}

func sameWorkspaceRecordPolicy(control *bodycodegen.RecordAssemblyControl, policy *bodycodegen.RecordAssemblyPolicy) bool {
	if control == nil || policy == nil {
		return control == nil && policy == nil
	}
	return control.Policy == *policy
}

func verifyWorkspaceContinuation(prior Result) error {
	stages, c := len(prior.PolicyHistory), prior.Continuation
	if stages == 0 {
		if c != nil || prior.Composition.Continuation != nil {
			return fmt.Errorf("workspace continuation policy history is missing")
		}
		return nil
	}
	if stages > 16 || c == nil || c.Schema != "gooo/workspace-body-continuation/v1" ||
		c.NewModelCalls != 0 || c.BodyFillsReplayed != len(prior.BodyFills) || c.SavedPolicyStages != stages ||
		prior.Composition.Continuation == nil || prior.AssemblyPolicy == nil {
		return fmt.Errorf("workspace continuation metadata or policy history differs")
	}
	value := c.PriorResultSHA256
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") || value != strings.ToLower(value) {
		return fmt.Errorf("workspace continuation parent digest is invalid")
	}
	if _, err := hex.DecodeString(value[7:]); err != nil {
		return fmt.Errorf("workspace continuation parent digest is invalid")
	}
	return nil
}
