package toolchainrelease

import "path/filepath"

func verifyReceiptArchive(directory string, receipt PlatformReceipt) (string, bool) {
	if receipt.Archive.Name == "" || filepath.Base(receipt.Archive.Name) != receipt.Archive.Name {
		return "", false
	}
	path := filepath.Join(directory, receipt.Archive.Name)
	digest, size, err := digestFile(path)
	if err != nil {
		return path, false
	}
	return path, digest == receipt.Archive.Digest && size == receipt.Archive.Bytes
}
