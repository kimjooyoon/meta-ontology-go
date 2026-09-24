package provenance

import "testing"

func TestModuleIdentityBindingPreservesIdentityAndFirstUnknownStagePart01(t *testing.T) {
	identity := ModuleReleaseIdentity{
		ModulePath: "example/core",
		Release:    "v1.0.0",
		Digest:     "sha256:core",
	}
	binding := BindModuleIdentityToPlanPart01(identity, identity, "source", "semantic", "plan")
	if binding.Status != ModuleIdentityObserved {
		t.Fatalf("status = %q, want OBSERVED", binding.Status)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("validate observed binding: %v", err)
	}

	missing := BindModuleIdentityToPlanPart01(identity, identity, "", "", "")
	if missing.Status != ModuleIdentityUnknown {
		t.Fatalf("status = %q, want UNKNOWN", missing.Status)
	}
	if missing.MissingStageIndex != 1 || missing.MissingStage != "source" {
		t.Fatalf("first missing stage = %d/%q, want 1/source", missing.MissingStageIndex, missing.MissingStage)
	}
	if err := missing.Validate(); err != nil {
		t.Fatalf("validate unknown binding: %v", err)
	}

	tampered := binding
	tampered.Identity.Digest = "sha256:tampered"
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered identity unexpectedly validated")
	}

	refuted := BindModuleIdentityToPlanPart01(
		identity,
		ModuleReleaseIdentity{ModulePath: "example/core", Release: "v1.1.0", Digest: "sha256:core-v11"},
		"source",
		"semantic",
		"plan",
	)
	if refuted.Status != ModuleIdentityRefuted {
		t.Fatalf("status = %q, want REFUTED", refuted.Status)
	}
	if err := refuted.Validate(); err != nil {
		t.Fatalf("validate refuted binding: %v", err)
	}
}
