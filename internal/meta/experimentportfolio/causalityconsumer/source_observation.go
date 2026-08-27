package causalityconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func ObserveSource(sourcePath string, source []byte) (SourceObservation, error) {
	value, err := computedValue(source)
	if err != nil {
		return SourceObservation{}, err
	}
	return SourceObservation{
		SourcePath:    canonicalSourcePath(sourcePath),
		SourceDigest:  sha256Digest(source),
		SemanticValue: value,
	}, nil
}

func computedValue(source []byte) (string, error) {
	text := string(source)
	marker := "computes "
	index := strings.Index(text, marker)
	if index < 0 {
		return "", fmt.Errorf("source does not declare a computes value")
	}
	quoted := text[index+len(marker):]
	if !strings.HasPrefix(quoted, "\"") {
		return "", fmt.Errorf("computes value is not a quoted string")
	}
	for end := 1; end < len(quoted); end++ {
		if quoted[end] != '"' || quoted[end-1] == '\\' {
			continue
		}
		value, err := strconv.Unquote(quoted[:end+1])
		if err != nil {
			return "", fmt.Errorf("computes value: %w", err)
		}
		if value == "" {
			return "", fmt.Errorf("computes value is empty")
		}
		return value, nil
	}
	return "", fmt.Errorf("computes value is unterminated")
}

func canonicalSourcePath(path string) string {
	path = filepath.ToSlash(path)
	const marker = "examples/experiment-portfolio/"
	if index := strings.Index(path, marker); index >= 0 {
		return path[index:]
	}
	return path
}

func sha256Digest(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
