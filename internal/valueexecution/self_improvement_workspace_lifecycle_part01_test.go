package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementWorkspaceLifecycle(t *testing.T) {
	observation := ObserveSelfImprovementWorkspaceLifecycle(
		SelfImprovementWorkspaceLifecycleInput{
			TaskID:          "task-1",
			Operation:       "SUSPEND",
			WorkspaceDigest: "workspace-digest",
			GatewayDigest:   "gateway-digest",
			ModelDigest:     "model-digest",
			StateDigest:     "state-digest",
		},
	)
	if observation.Status != SelfImprovementWorkspaceLifecycleSuspended {
		t.Fatalf("status = %q, want SUSPENDED", observation.Status)
	}
	if !observation.NonAuthorizing {
		t.Fatal("lifecycle observation must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementWorkspaceLifecycleRequiresProvenance(t *testing.T) {
	observation := ObserveSelfImprovementWorkspaceLifecycle(
		SelfImprovementWorkspaceLifecycleInput{
			TaskID:    "task-1",
			Operation: "RESUME",
		},
	)
	if observation.Status != SelfImprovementWorkspaceLifecycleUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "LIFECYCLE_PROVENANCE_MISSING" {
		t.Fatalf("reason = %q, want LIFECYCLE_PROVENANCE_MISSING", observation.Reason)
	}
}