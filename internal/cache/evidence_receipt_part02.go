package cache

import (
	"fmt"
)

// Validate rejects missing, zero, duplicated, or malformed evidence.
func (e EvidenceFreshness) Validate() error {
	for _, item := range []struct {
		label string
		value Digest
	}{
		{"base", e.BaseDigest}, {"head", e.HeadDigest}, {"source", e.SourceDigest},
		{"IR", e.IRDigest}, {"policy", e.PolicyDigest}, {"toolchain", e.ToolchainDigest},
		{"target", e.TargetDigest}, {"bundle", e.BundleDigest},
	} {
		if !item.value.Known() {
			return fmt.Errorf("%w: unknown %s digest", ErrInvalidReceipt, item.label)
		}
	}
	if !validCommitSHA(e.BaseSHA) || !validCommitSHA(e.HeadSHA) {
		return fmt.Errorf("%w: missing immutable base/head SHA", ErrInvalidReceipt)
	}
	if err := validateFreshnessRefs(e.EventRef, e.CheckoutRef, e.HeadSHA); err != nil {
		return err
	}
	if e.RunID == "" || e.Event != "pull_request" || e.EventID == "" || e.Attempt == 0 {
		return fmt.Errorf("%w: missing immutable event attempt", ErrInvalidReceipt)
	}
	if err := validateFreshnessJobs(e.Jobs, e.RunID, e.Attempt, e.HeadSHA); err != nil {
		return err
	}
	if e.PredecessorDigests == nil || hasDuplicateDigests(e.PredecessorDigests) {
		return fmt.Errorf("%w: missing or replayed predecessors", ErrInvalidReceipt)
	}
	if len(e.EvidenceRefs) == 0 || hasDuplicateEvidenceRefs(e.EvidenceRefs) {
		return fmt.Errorf("%w: missing or duplicated evidence refs", ErrInvalidReceipt)
	}
	for _, digest := range e.PredecessorDigests {
		if !digest.Known() {
			return fmt.Errorf("%w: unknown predecessor digest", ErrInvalidReceipt)
		}
	}
	for _, ref := range e.EvidenceRefs {
		if err := validateKeyComponent("evidence ref", ref.Name, true); err != nil ||
			!knownEvidenceRef(ref.Name) || !ref.Digest.Known() {
			return fmt.Errorf("%w: malformed evidence ref %q", ErrInvalidReceipt, ref.Name)
		}
	}
	if err := validateEvidenceBindings(e); err != nil {
		return err
	}
	return nil
}

// Equal reports whether two evidence bundles identify the same immutable run.
func (e EvidenceFreshness) Equal(other EvidenceFreshness) bool {
	left, right := canonicalEvidence(e), canonicalEvidence(other)
	return left.BaseDigest == right.BaseDigest && left.HeadDigest == right.HeadDigest &&
		left.BaseSHA == right.BaseSHA && left.HeadSHA == right.HeadSHA && left.CheckoutRef == right.CheckoutRef &&
		left.Event == right.Event && left.EventRef == right.EventRef &&
		left.RunID == right.RunID && left.EventID == right.EventID && left.Attempt == right.Attempt &&
		freshnessJobsEqual(left.Jobs, right.Jobs) &&
		left.SourceDigest == right.SourceDigest &&
		left.IRDigest == right.IRDigest && left.PolicyDigest == right.PolicyDigest &&
		left.ToolchainDigest == right.ToolchainDigest && left.TargetDigest == right.TargetDigest &&
		left.BundleDigest == right.BundleDigest && digestSliceEqual(left.PredecessorDigests, right.PredecessorDigests) &&
		evidenceRefsEqual(left.EvidenceRefs, right.EvidenceRefs)
}
