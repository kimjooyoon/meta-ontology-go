package toolchainrelease

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"time"
)

func writeTarGzip(path, name string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	gzipWriter.Header.ModTime = time.Unix(0, 0).UTC()
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	header := &tar.Header{
		Name: name, Mode: 0o755, Size: int64(len(data)),
		ModTime: time.Unix(0, 0).UTC(), Format: tar.FormatUSTAR,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tarWriter.Write(data); err != nil {
		return err
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	return gzipWriter.Close()
}
