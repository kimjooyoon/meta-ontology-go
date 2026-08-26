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

func createArchive(path string, target Target, binary []byte) error {
	switch target.ArchiveFormat {
	case "tar.gz":
		return writeTarGzip(path, binaryName(target), binary)
	case "zip":
		return writeZip(path, binaryName(target), binary)
	default:
		return fmt.Errorf("TOOLCHAIN_RELEASE_ARCHIVE_FORMAT_UNKNOWN")
	}
}

func archiveBinary(path, binaryPath string, target Target) error {
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}
	return createArchive(path, target, binary)
}
