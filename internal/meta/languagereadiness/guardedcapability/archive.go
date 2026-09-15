package guardedcapability

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

func VerifyFoundationArchive(raw []byte) error {
	if digestBytes(raw) != FoundationArtifactDigest {
		return fmt.Errorf("foundation archive digest mismatch")
	}
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return fmt.Errorf("open foundation archive: %w", err)
	}
	matches := 0
	for _, file := range reader.File {
		if file.Name != "guarded-promotion-a.json" && file.Name != "guarded-promotion-b.json" {
			continue
		}
		opened, openErr := file.Open()
		if openErr != nil {
			return openErr
		}
		data, readErr := io.ReadAll(opened)
		opened.Close()
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(data, foundationRaw) || digestBytes(data) != FoundationReportFileSHA {
			return fmt.Errorf("foundation report does not match pinned receipt")
		}
		matches++
	}
	if matches != 2 {
		return fmt.Errorf("foundation reports = %d, want 2", matches)
	}
	return nil
}
