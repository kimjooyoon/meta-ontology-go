package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const failureOwnerRegistryPath = ".github/ci-governance.json"
const promotionOwnerBindingCode = "CI-PROMOTION-OWNER-BINDING-001"

type catalogDocumentEntry struct {
	Code            string
	Class           string
	Severity        string
	BlockingScope   string
	Parallelizable  bool
	HandoffRequired bool
	Owner           string
}
type failureOwnerRegistryEntry struct {
	Branch string   `json:"branch"`
	Paths  []string `json:"paths"`
}
type failureOwnerRegistry struct {
	Schema                string                      `json:"schema"`
	ProtectedPushBranches []string                    `json:"protected_push_branches"`
	Ownership             []failureOwnerRegistryEntry `json:"ownership"`
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
