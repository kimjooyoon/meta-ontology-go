package artifactcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func AuthorityDigest(root string, program Program) (string, error) {
	if err := Validate(program); err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(filepath.Clean(root), program.AuthorityPath))
	if err != nil {
		return "", err
	}
	text := string(data)
	if !strings.Contains(text, "package operationartifactcoverage") {
		return "", fmt.Errorf("operation artifact authority package is missing")
	}
	for _, operation := range program.MetaOperations {
		if !strings.Contains(text, "activity "+operation.Activity+"(") {
			return "", fmt.Errorf("operation artifact authority is missing %q", operation.Activity)
		}
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
