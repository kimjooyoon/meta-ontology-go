package toolchainrelease

import (
	"fmt"
	"os"
)

func archiveName(target Target) string {
	return "gooo-" + target.ID + "." + target.ArchiveFormat
}

func binaryName(target Target) string {
	if target.GOOS == "windows" {
		return "gooo.exe"
	}
	return "gooo"
}

func createArchive(path string, target Target, members []archiveMember) error {
	switch target.ArchiveFormat {
	case "tar.gz":
		return writeTarGzip(path, members)
	case "zip":
		return writeZip(path, members)
	default:
		return fmt.Errorf("TOOLCHAIN_RELEASE_ARCHIVE_FORMAT_UNKNOWN")
	}
}

func archiveBinary(path, binaryPath string, target Target, payload releasePayload,
	head string) (PayloadEvidence, error) {
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		return PayloadEvidence{}, err
	}
	members, evidence := payload.withBinary(binary, head)
	return evidence, createArchive(path, target, members)
}
