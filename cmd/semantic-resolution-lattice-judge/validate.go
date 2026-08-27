package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	if got.RepositoryWrites != 0 || got.MutationAuthority || got.CaseDenominator != 4 || got.Counts.CasesTotal != 4 {
		return errors.New("effect or denominator guardrail failed")
	}
	if got.Counts.Pass != 1 || got.Counts.FailClosed != 2 || got.Counts.Unknown != 1 || len(got.Cases) != 4 {
		return errors.New("case counts are not the fixed contract")
	}
	if !hasGoooRelations(string(source)) {
		return errors.New("Gooo source relations are missing")
	}
	for _, item := range got.Cases {
		if err := validateCase(item); err != nil {
			return fmt.Errorf("case %s: %w", item.ID, err)
		}
	}
	if err := validateClaims(got.Claims, got.Cases); err != nil {
		return err
	}
	return validateMetrics(got.Metrics)
}
