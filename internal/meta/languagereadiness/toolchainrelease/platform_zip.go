package toolchainrelease

import (
	"archive/zip"
	"os"
	"time"
)

func writeZip(path, name string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(0o755)
	header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
	entry, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := entry.Write(data); err != nil {
		return err
	}
	return writer.Close()
}
