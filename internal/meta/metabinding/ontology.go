package metabinding

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func readOntology(root string) (string, error) {
	path := filepath.Join(filepath.Clean(root), OntologyPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	required := [][]byte{[]byte("package metabindingcoverage"),
		[]byte("activity ResolveIndicatorOperation("), []byte("activity BindOntologyAuthority("),
		[]byte("activity BindIndicatorMetaProgram(")}
	for _, token := range required {
		if !bytes.Contains(data, token) {
			return "", fmt.Errorf("meta-binding ontology is missing %q", token)
		}
	}
	return digestBytes(data), nil
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}
