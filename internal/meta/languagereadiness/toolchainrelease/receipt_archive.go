package toolchainrelease

import "path/filepath"

func verifyReceiptArchive(root, directory string, receipt PlatformReceipt) (string, bool) {
	if receipt.Archive.Name == "" || filepath.Base(receipt.Archive.Name) != receipt.Archive.Name {
		return "", false
	}
	path := filepath.Join(directory, receipt.Archive.Name)
	digest, size, err := digestFile(path)
	if err != nil {
		return path, false
	}
	if digest != receipt.Archive.Digest || size != receipt.Archive.Bytes ||
		receipt.Schema != PlatformReceiptSchema || receipt.Binary.Name != binaryName(receipt.Platform) {
		return path, false
	}
	payload, err := buildReleasePayload(root, receipt.Platform, receipt.HeadSHA)
	if err != nil || !samePayloadEvidence(receipt.Payload,
		expectedPayloadEvidence(payload, receipt.HeadSHA, receipt.Binary)) {
		return path, false
	}
	return path, verifyArchiveMembers(path, receipt.Platform, receipt.Payload.Members) == nil
}
