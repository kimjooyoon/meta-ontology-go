package cache

import (
	"fmt"
)

// ValidateForData binds an artifact receipt to the exact bytes it describes.
func (r CacheReceipt) ValidateForData(data []byte) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if !r.hasArtifact() || r.ContentDigest != HashBytes(data) || r.Size != int64(len(data)) {
		return fmt.Errorf("%w: receipt content differs from artifact", ErrInvalidReceipt)
	}
	return nil
}

// Seal validates and computes the immutable receipt digest.
func (r CacheReceipt) Seal() (CacheReceipt, error) {
	if err := r.Validate(); err != nil {
		return CacheReceipt{}, err
	}
	r.Evidence = canonicalEvidence(r.Evidence)
	r.EvidenceRefs = append([]EvidenceRef(nil), r.Evidence.EvidenceRefs...)
	provided := r.ReceiptDigest
	r.ReceiptDigest = ""
	digest, err := DigestOf(r)
	if err != nil {
		return CacheReceipt{}, fmt.Errorf("%w: hash receipt: %v", ErrInvalidReceipt, err)
	}
	if provided != "" && provided != digest {
		return CacheReceipt{}, fmt.Errorf("%w: receipt digest mismatch", ErrInvalidReceipt)
	}
	r.ReceiptDigest = digest
	return r, nil
}
func validReceiptStatus(status ReceiptStatus) bool {
	return status == ReceiptHit || status == ReceiptMiss || status == ReceiptRecomputed ||
		status == ReceiptStale || status == ReceiptCorrupt
}
func (r CacheReceipt) hasArtifact() bool {
	return r.Status == ReceiptHit || r.Status == ReceiptRecomputed
}
func validateEvidenceRefs(refs []EvidenceRef) error {
	if len(refs) == 0 || hasDuplicateEvidenceRefs(refs) {
		return fmt.Errorf("%w: missing or duplicated evidence refs", ErrInvalidReceipt)
	}
	for _, ref := range refs {
		if err := validateKeyComponent("evidence ref", ref.Name, true); err != nil ||
			!knownEvidenceRef(ref.Name) || !ref.Digest.Known() {
			return fmt.Errorf("%w: malformed evidence ref %q", ErrInvalidReceipt, ref.Name)
		}
	}
	return nil
}
func knownEvidenceRef(name string) bool {
	switch name {
	case "source", "ir", "policy", "toolchain", "target", "bundle":
		return true
	default:
		return false
	}
}
