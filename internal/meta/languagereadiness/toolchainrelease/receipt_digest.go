package toolchainrelease

func FinalizeReceipt(receipt PlatformReceipt) (PlatformReceipt, error) {
	receipt.ReceiptDigest = ""
	digest, err := digestValue(receipt)
	if err != nil {
		return PlatformReceipt{}, err
	}
	receipt.ReceiptDigest = digest
	return receipt, nil
}

func receiptDigestValid(receipt PlatformReceipt) bool {
	expected := receipt.ReceiptDigest
	finalized, err := FinalizeReceipt(receipt)
	return err == nil && expected != "" && finalized.ReceiptDigest == expected
}
