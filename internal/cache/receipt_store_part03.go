package cache

import (
	"fmt"
	"os"
)

func validateSealedReceipt(receipt CacheReceipt) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	if !receipt.ReceiptDigest.Known() {
		return fmt.Errorf("%w: missing receipt digest", ErrInvalidReceipt)
	}
	copy := receipt
	copy.Evidence = canonicalEvidence(copy.Evidence)
	copy.EvidenceRefs = append([]EvidenceRef(nil), copy.Evidence.EvidenceRefs...)
	copy.ReceiptDigest = ""
	digest, err := DigestOf(copy)
	if err != nil || digest != receipt.ReceiptDigest {
		return fmt.Errorf("%w: receipt digest mismatch", ErrInvalidReceipt)
	}
	return nil
}
func validateReceiptFile(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsafeReceiptLog
	}
	return nil
}
