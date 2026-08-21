package main

import (
	"testing"
)

func TestFailureOwnerRegistryAcceptsRegisteredOwners(t *testing.T) {
	for _, branch := range []string{"agent/ci-workflow", "agent/docs"} {
		t.Run(branch, func(t *testing.T) {
			if err := validateFailureOwnerRegistry(branch); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestFailureManifestAcceptsRegisteredDocsOwner(t *testing.T) {
	binding := validFailureBinding()
	binding.OwnerBranch = "agent/docs"
	input := validFailureInput()
	input.OwnerBranch = binding.OwnerBranch
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.OwnerBranch != binding.OwnerBranch || manifest.OwnerRef == "" {
		t.Fatalf("registered docs owner was not bound: %+v", manifest)
	}
}
