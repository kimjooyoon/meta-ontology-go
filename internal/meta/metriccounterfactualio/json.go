package metriccounterfactualio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func Digest(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func WriteJSON(output string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(output, append(data, '\n'), 0o644)
}

func ReadJSON[T any](input string) (T, error) {
	var zero T
	data, err := os.ReadFile(input)
	if err != nil {
		return zero, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value T
	if err := decoder.Decode(&value); err != nil {
		return zero, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return zero, fmt.Errorf("document has trailing JSON")
	}
	return value, nil
}

func Equal(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func ValidSubject(repository, subjectSHA string) bool {
	if strings.TrimSpace(repository) == "" || len(subjectSHA) != 40 {
		return false
	}
	for _, char := range subjectSHA {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}
