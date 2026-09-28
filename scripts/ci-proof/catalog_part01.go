package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const promotionBranchBindingCode = "CI-PROMOTION-BRANCH-BINDING-001"

type catalogDocumentEntry struct {
	Code           string
	Class          string
	Severity       string
	BlockingScope  string
	Parallelizable bool
	NextOperation  string
}

var failureCatalogDigest, failureCatalogDigestErr = loadFailureCatalogDigest()

func loadFailureCatalogDigest() (string, error) {
	data, err := readFailureFile(failureCatalogPath)
	if err != nil {
		return "", fmt.Errorf("read failure catalog: %w", err)
	}
	return "sha256:" + digestBytes(data), nil
}
func readFailureFile(name string) ([]byte, error) {
	candidates := []string{name, filepath.Join("..", name), filepath.Join("..", "..", name)}
	var lastErr error
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
