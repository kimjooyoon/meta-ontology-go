package metriccounterfactualio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func SafeNative(root, relative string) (string, error) {
	if relative == "" || strings.Contains(relative, "\\") ||
		strings.HasPrefix(relative, "/") || path.Clean(relative) != relative {
		return "", fmt.Errorf("unsafe fixture path %q", relative)
	}
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func ContentDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func CountLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}
