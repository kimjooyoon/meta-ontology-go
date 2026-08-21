package metriccounterfactual

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
	data = append(data, '\n')
	return os.WriteFile(output, data, 0o644)
}

func ReadLedger(input string) (Ledger, error) {
	data, err := os.ReadFile(input)
	if err != nil {
		return Ledger{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var ledger Ledger
	if err := decoder.Decode(&ledger); err != nil {
		return Ledger{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Ledger{}, fmt.Errorf("ledger has trailing JSON")
	}
	return ledger, nil
}

func CanonicalEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func SealManifest(value Manifest) (Manifest, error) {
	value.Digest = ""
	digest, err := Digest(value)
	value.Digest = digest
	return value, err
}

func SealPlan(value Plan) (Plan, error) {
	value.Digest = ""
	digest, err := Digest(value)
	value.Digest = digest
	return value, err
}

func SealState(value State) (State, error) {
	value.Digest = ""
	digest, err := Digest(value)
	value.Digest = digest
	return value, err
}

func SealLedger(value Ledger) (Ledger, error) {
	value.Digest = ""
	digest, err := Digest(value)
	value.Digest = digest
	return value, err
}

func ValidManifest(value Manifest) bool {
	digest := value.Digest
	sealed, err := SealManifest(value)
	return err == nil && digest == sealed.Digest
}

func ValidPlan(value Plan) bool {
	digest := value.Digest
	sealed, err := SealPlan(value)
	return err == nil && digest == sealed.Digest
}

func ValidState(value State) bool {
	digest := value.Digest
	sealed, err := SealState(value)
	return err == nil && digest == sealed.Digest
}

func ValidLedger(value Ledger) bool {
	digest := value.Digest
	sealed, err := SealLedger(value)
	return err == nil && digest == sealed.Digest
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
