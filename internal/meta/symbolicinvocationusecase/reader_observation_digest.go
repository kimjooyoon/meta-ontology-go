package symbolicinvocationusecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func readerObservationBytesDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func readerObservationReportDigest(report ReaderObservationReport) string {
	report.Digest = ""
	data, _ := json.Marshal(report)
	return readerObservationBytesDigest(data)
}

func readerObservationValidDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
