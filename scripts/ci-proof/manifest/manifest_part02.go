package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generatedPath(root string, record map[string]json.RawMessage, lineNumber int) (string, error) {
	var generatedFile string
	rawPath, ok := record["generated_file"]
	if !ok || json.Unmarshal(rawPath, &generatedFile) != nil || generatedFile == "" {
		return "", fmt.Errorf("generated manifest line %d has no generated_file", lineNumber)
	}
	if !filepath.IsAbs(generatedFile) {
		generatedFile = filepath.Join(root, generatedFile)
	}
	candidate, err := filepath.Abs(generatedFile)
	if err != nil {
		return "", fmt.Errorf("generated manifest line %d path: %w", lineNumber, err)
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("generated manifest line %d generated_file escapes output root", lineNumber)
	}
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("generated manifest line %d generated_file is not a file", lineNumber)
	}
	return filepath.ToSlash(relative), nil
}
func validDigest(value string) bool {
	if len(value) != 64 || strings.Trim(value, "0") == "" {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
