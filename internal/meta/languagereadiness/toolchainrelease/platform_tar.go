package toolchainrelease

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"strconv"
	"time"
)

func writeTarGzip(path string, members []archiveMember) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	gzipWriter.Header.ModTime = time.Unix(0, 0).UTC()
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	for _, member := range members {
		mode, err := strconv.ParseInt(member.evidence.Mode, 8, 64)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_MEMBER_MODE_INVALID: %w", err)
		}
		header := &tar.Header{Name: member.evidence.Name, Mode: mode, Size: int64(len(member.data)),
			ModTime: time.Unix(0, 0).UTC(), Format: tar.FormatUSTAR}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write(member.data); err != nil {
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	return gzipWriter.Close()
}
