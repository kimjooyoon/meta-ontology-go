package cache

import (
	"fmt"
)

// Validate checks the required cache receipt schema before it is sealed.
func (r CacheReceipt) Validate() error {
	if r.SchemaVersion != cacheReceiptSchemaVersion || !r.CacheKey.Known() ||
		!r.SemanticClosureDigest.Known() || !r.DependencyRoot.Known() || r.DirectDependencies == nil ||
		!r.PolicySchemaDigest.Known() || r.Toolchain == "" || r.Target == "" || !r.BuildTagsDigest.Known() ||
		!r.OptionsDigest.Known() ||
		r.ProducerHost == "" || !validReceiptStatus(r.Status) || r.Size < 0 {
		return fmt.Errorf("%w: required cache receipt field missing", ErrInvalidReceipt)
	}
	for _, field := range []struct{ label, value string }{
		{"domain", r.Domain}, {"key version", r.KeyVersion}, {"artifact kind", r.ArtifactKind},
		{"projection", r.Projection}, {"toolchain", r.Toolchain}, {"target", r.Target},
		{"producer host", r.ProducerHost},
	} {
		if err := validateKeyComponent(field.label, field.value, true); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidReceipt, err)
		}
	}
	if !r.HostStage.Valid() {
		return fmt.Errorf("%w: invalid host stage", ErrInvalidReceipt)
	}
	if r.hasArtifact() && (!r.ContentDigest.Known() || !r.Reconstructable) {
		return fmt.Errorf("%w: artifact receipt is not reconstructable", ErrInvalidReceipt)
	}
	if !r.hasArtifact() && (r.ContentDigest != "" || r.Size != 0 || r.Reconstructable) {
		return fmt.Errorf("%w: non-artifact receipt has content", ErrInvalidReceipt)
	}
	if err := validateEvidenceRefs(r.EvidenceRefs); err != nil {
		return err
	}
	if err := r.Evidence.Validate(); err != nil {
		return err
	}
	if !evidenceRefsEqual(r.EvidenceRefs, r.Evidence.EvidenceRefs) {
		return fmt.Errorf("%w: evidence refs disagree", ErrInvalidReceipt)
	}
	for _, digest := range r.DirectDependencies {
		if !digest.Known() {
			return fmt.Errorf("%w: unknown direct dependency", ErrInvalidReceipt)
		}
	}
	return nil
}

// ValidateForKey prevents metadata-only artifact or projection aliases from
// being accepted as a different content-addressed projection.
func (r CacheReceipt) ValidateForKey(key Key) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := validateFullKey(key); err != nil {
		return err
	}
	if r.CacheKey != key.Digest || r.Domain != key.Domain || r.KeyVersion != key.Version ||
		r.HostStage != key.HostStage || r.ArtifactKind != key.ArtifactKind || r.Projection != key.Projection ||
		r.SemanticClosureDigest != key.SemanticClosureDigest || r.DependencyRoot != key.DependencyRoot ||
		r.PolicySchemaDigest != key.PolicySchemaDigest || r.Toolchain != key.Toolchain ||
		r.OptionsDigest != key.OptionsDigest ||
		r.Target != key.Target || r.BuildTagsDigest != key.BuildTagsDigest {
		return fmt.Errorf("%w: receipt identity differs from key", ErrInvalidReceipt)
	}
	return nil
}
