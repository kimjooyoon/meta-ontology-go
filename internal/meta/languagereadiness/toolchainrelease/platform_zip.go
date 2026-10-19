package toolchainrelease

import (
	"archive/zip"
	"fmt"
	"os"
	"strconv"
	"time"
)

func writeZip(path string, members []archiveMember) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	for _, member := range members {
		mode, err := strconv.ParseUint(member.evidence.Mode, 8, 32)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_MODE_INVALID: %w", err)
		}
		header := &zip.FileHeader{Name: member.evidence.Name, Method: zip.Deflate}
		header.SetMode(os.FileMode(mode))
		header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := entry.Write(member.data); err != nil {
			return err
		}
	}
	return writer.Close()
}
