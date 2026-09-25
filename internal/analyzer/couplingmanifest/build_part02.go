package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

// ValidateManifest exposes the detector's exact result for callers that have a
// complete detector input packet. This wrapper performs no normalization.
func ValidateManifest(manifest detector.ChangeManifest, authority detector.AuthorityContext) detector.Result {
	return detector.Evaluate(detectorInput(manifest, authority), authority)
}

// Evaluate forwards an exact detector packet and authority context unchanged.
func Evaluate(input detector.Input, authority detector.AuthorityContext) detector.Result {
	return detector.Evaluate(input, authority)
}

// BuildManifest is the descriptive spelling of Build.
func BuildManifest(input Input) (Manifest, error) { return Build(input) }

// Adapt is a vocabulary alias for Build.
func Adapt(input Input) (Manifest, error) { return Build(input) }

// New constructs a manifest from the explicit adapter input.
func New(input Input) (Manifest, error) { return Build(input) }
func detectorInput(manifest detector.ChangeManifest, authority detector.AuthorityContext) detector.Input {
	return detector.Input{
		Schema: detector.InputSchemaV1,
		Config: detector.Config{
			Schema: detector.ConfigSchemaV1, RegistryDigest: authority.Registry.Digest,
			ToolchainDigest: authority.ToolchainDigest, ProfileDigest: authority.ProfileDigest,
			SnapshotDigest: authority.SnapshotDigest, ExpectedProviderDigest: authority.ExpectedProviderDigest,
			ExpectedObserverDigest: authority.ExpectedObserverDigest, Baseline: authority.Baseline,
			ExternalReceiptRequired: authority.ExternalReceiptRequired,
		},
		Registry: authority.Registry, Manifest: manifest, Receipts: []detector.CouplingReceipt{},
	}
}
func validateSnapshots(input Input) *ConstructionError {
	if input.Before == nil || input.Head == nil {
		return unknownError(CodeMissingSnapshot, "before and head snapshots are required")
	}
	if err := input.Before.Validate(); err != nil {
		return unknownError(CodeInvalidSnapshot, "before snapshot is not valid: %s", err.Error())
	}
	if err := input.Head.Validate(); err != nil {
		return unknownError(CodeInvalidSnapshot, "head snapshot is not valid: %s", err.Error())
	}
	return nil
}
