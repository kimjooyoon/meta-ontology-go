package main

import (
	"fmt"
)

func failureArtifactRefs(binding failureBinding, artifacts []artifactInput, proofArtifact *artifactInput) []string {
	refs := make([]string, 0, len(artifacts)+1)
	for _, artifact := range artifacts {
		refs = append(refs, fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d", binding.Repository, binding.RunID, artifact.ID))
	}
	if proofArtifact != nil {
		refs = append(refs, fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d", binding.Repository, binding.RunID, proofArtifact.ID))
	}
	return refs
}
func failureArtifactInputs(artifacts []artifactInput, proofArtifact *artifactInput) []artifactInput {
	refs := append([]artifactInput(nil), artifacts...)
	if proofArtifact != nil {
		refs = append(refs, *proofArtifact)
	}
	return refs
}
func failureEvidenceRefs(manifest failureManifest, runRef, jobRef string) []string {
	refs := []string{runRef, jobRef, manifest.OwnerRef}
	refs = append(refs, manifest.ArtifactURLs...)
	return append(refs, manifest.CatalogRef, manifest.CatalogSHA256)
}
