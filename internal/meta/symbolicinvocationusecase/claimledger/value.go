package claimledger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func validateContract(contract Contract, subject string) error {
	if contract.Schema != ContractSchema {
		return fmt.Errorf("contract schema %q is not %q", contract.Schema, ContractSchema)
	}
	if contract.Metric == "" || subject == "" {
		return fmt.Errorf("metric and subject are required")
	}
	if contract.Expected.FixedClaimTotal != len(contract.Claims) {
		return fmt.Errorf("fixed claim total %d does not match %d claims", contract.Expected.FixedClaimTotal, len(contract.Claims))
	}
	seen := map[string]bool{}
	for _, spec := range contract.Claims {
		if err := validateClaim(spec, seen); err != nil {
			return err
		}
		seen[spec.ID] = true
	}
	return nil
}

func validateClaim(spec ClaimSpec, seen map[string]bool) error {
	if spec.ID == "" || seen[spec.ID] {
		return fmt.Errorf("claim id %q is empty or duplicated", spec.ID)
	}
	if spec.Kind == "" || spec.Modality == "" || spec.Subject == "" || spec.Predicate == "" {
		return fmt.Errorf("claim %q is incomplete", spec.ID)
	}
	if !validStage(spec.Coordinate.Stage) || spec.Coordinate.Step == "" {
		return fmt.Errorf("claim %q has no precise process coordinate", spec.ID)
	}
	if spec.ProofRoute != "FOUNDATION" && spec.ProofRoute != "COHERENCE" && spec.ProofRoute != "REGRESSION" {
		return fmt.Errorf("claim %q has invalid proof route %q", spec.ID, spec.ProofRoute)
	}
	if spec.Scope == "EXCLUDED" {
		if spec.ExcludedReason == "" {
			return fmt.Errorf("excluded claim %q has no reason", spec.ID)
		}
		return nil
	}
	if spec.Scope != "IN_SCOPE" || spec.Evidence == nil || len(spec.Evidence.Paths) == 0 {
		return fmt.Errorf("in-scope claim %q has no evidence selector", spec.ID)
	}
	if spec.Evidence.Operator != "EQUALS" && spec.Evidence.Operator != "NON_NULL" && spec.Evidence.Operator != "POSITIVE_INTEGER" {
		return fmt.Errorf("claim %q has invalid operator %q", spec.ID, spec.Evidence.Operator)
	}
	if spec.Evidence.Operator == "EQUALS" && len(spec.Evidence.Expected) == 0 {
		return fmt.Errorf("claim %q has no expected value", spec.ID)
	}
	if spec.UnknownReason == "" || spec.RefutedReason == "" {
		return fmt.Errorf("claim %q has incomplete failure reasons", spec.ID)
	}
	return nil
}

func validStage(stage string) bool {
	switch stage {
	case "SOURCE", "PARSE", "BIND", "COMPILE", "EMIT", "TRANSPORT", "OBSERVE", "CONFORM", "PROMOTE":
		return true
	default:
		return false
	}
}

func lookupAny(root map[string]any, paths []string) (string, any, bool) {
	for _, path := range paths {
		value, found := lookup(root, path)
		if found {
			return path, value, true
		}
	}
	return "", nil, false
}

func lookup(root map[string]any, path string) (any, bool) {
	var current any = root
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func evidenceMatches(spec EvidenceSpec, observed any, subject string) (bool, string) {
	if spec.Operator == "NON_NULL" {
		return observed != nil, ""
	}
	if spec.Operator == "POSITIVE_INTEGER" {
		number, ok := observed.(json.Number)
		if !ok {
			return false, ""
		}
		value, err := number.Int64()
		return err == nil && value > 0, ""
	}
	decoder := json.NewDecoder(bytes.NewReader(spec.Expected))
	decoder.UseNumber()
	var expected any
	if err := decoder.Decode(&expected); err != nil {
		return false, "invalid-expected-value"
	}
	if text, ok := expected.(string); ok && text == "$SUBJECT_SHA" {
		expected = subject
	}
	observedJSON, observedErr := json.Marshal(observed)
	expectedJSON, expectedErr := json.Marshal(expected)
	return observedErr == nil && expectedErr == nil && bytes.Equal(observedJSON, expectedJSON), digestValue(expected)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "unencodable"
	}
	return digestBytes(encoded)
}

func countProofRoute(counts *ProofRouteCounts, route string) {
	switch route {
	case "FOUNDATION":
		counts.Foundation++
	case "COHERENCE":
		counts.Coherence++
	case "REGRESSION":
		counts.Regression++
	}
}
