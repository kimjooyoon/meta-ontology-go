package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

const latticeSchema = "gooo/meta-semantic-resolution-lattice/v1"

func validate(sourcePath, receiptPath string) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	var got receipt
	if err := json.Unmarshal(data, &got); err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	digest := sha256.Sum256(source)
	if got.Schema != latticeSchema || got.Source != sourcePath || got.SourceSHA256 != "sha256:"+hex.EncodeToString(digest[:]) {
		return errors.New("receipt identity is not bound to the source")
	}
	declared, err := parseGoooCases(string(source))
	if err != nil {
		return fmt.Errorf("parse source semantics: %w", err)
	}
	if got.SemanticDigest != sourceSemanticDigest(declared) {
		return errors.New("receipt semantic digest is not bound to Gooo values")
	}
	if got.RepositoryWrites != 0 || got.MutationAuthority || got.CaseDenominator != 4 || got.Counts.CasesTotal != 4 {
		return errors.New("effect or denominator guardrail failed")
	}
	if got.Counts.Pass != 1 || got.Counts.FailClosed != 2 || got.Counts.Unknown != 1 || len(got.Cases) != 4 {
		return errors.New("case counts are not the fixed contract")
	}
	expectedCases := make([]latticeCase, 0, len(declared))
	expectedClaims := make([]claim, 0, len(declared))
	for _, item := range declared {
		expected := reconstructCase(item)
		expectedCases = append(expectedCases, expected)
		expectedClaims = append(expectedClaims, reconstructClaim(item, expected.Transition))
	}
	if !reflect.DeepEqual(got.Cases, expectedCases) {
		return errors.New("receipt cases are not reconstructed from Gooo values")
	}
	for _, item := range got.Cases {
		if err := validateCase(item); err != nil {
			return fmt.Errorf("case %s: %w", item.ID, err)
		}
	}
	if !reflect.DeepEqual(got.Claims, expectedClaims) {
		return errors.New("receipt claims are not reconstructed from Gooo values")
	}
	if err := validateClaims(got.Claims, got.Cases); err != nil {
		return err
	}
	if err := validateCounterfactuals(string(source), expectedCases, expectedClaims, got.Counterfactuals); err != nil {
		return err
	}
	return validateMetrics(got.Metrics, got.Counterfactuals)
}
