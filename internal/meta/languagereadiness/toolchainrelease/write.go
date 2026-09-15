package toolchainrelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func WritePlatformReceipt(path string, receipt PlatformReceipt) error {
	return writeArtifact(path, receipt, receipt.ReceiptDigest)
}

func WriteReport(path string, report Report) error {
	return writeArtifact(path, report, report.ReportDigest)
}

func writeArtifact(path string, value any, digest string) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return err
	}
	line := strings.TrimPrefix(digest, "sha256:") + "  " + filepath.Base(path) + "\n"
	return os.WriteFile(path+".sha256", []byte(line), 0o644)
}
