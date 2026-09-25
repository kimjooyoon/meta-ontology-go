package guardedpromotion

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func promotionJSON(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open promotion artifact: %w", err)
	}
	var matches [][]byte
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !strings.HasSuffix(file.Name, ".json") {
			continue
		}
		opened, openErr := file.Open()
		if openErr != nil {
			return nil, openErr
		}
		data, readErr := io.ReadAll(opened)
		opened.Close()
		if readErr != nil {
			return nil, readErr
		}
		matches = append(matches, data)
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("promotion json files = %d, want 1", len(matches))
	}
	return matches[0], nil
}

func decodePromotion(data []byte, predecessorSHA string) (promotionEnvelope, error) {
	var envelope promotionEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return envelope, err
	}
	if envelope.Schema != PromotionSchema || envelope.CurrentHeadSHA != predecessorSHA ||
		envelope.Decision != "PASS" || !validDigest(envelope.ReportDigest) ||
		envelope.Summary.Satisfied != 8 || envelope.Summary.Total != 8 ||
		envelope.Summary.Unresolved != 0 || envelope.Summary.RepositoryWrites != 0 {
		return envelope, fmt.Errorf("predecessor promotion contract is not exact")
	}
	return envelope, nil
}
