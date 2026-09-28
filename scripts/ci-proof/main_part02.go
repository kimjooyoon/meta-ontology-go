package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func verifyProof(filename, governance, receipt string, requirePass bool) error {
	if _, err := readGovernance(governance); err != nil {
		return err
	}
	bundle, err := readStrictJSON[proofBundle](filename)
	if err != nil {
		return err
	}
	if err := validateProof(bundle); err != nil {
		return err
	}
	if err := verifyReceipt(receipt, bundle); err != nil {
		return err
	}
	if requirePass && bundle.Decision != "PASS" {
		return fmt.Errorf("proof decision is %s", bundle.Decision)
	}
	return nil
}
func readGovernance(filename string) (governanceInput, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return governanceInput{}, err
	}
	var matrix governanceInput
	if err := json.Unmarshal(data, &matrix); err != nil {
		return governanceInput{}, err
	}
	if matrix.Schema != "gooo/ci-governance/v2" || matrix.Promotion.Source != "dev" || matrix.Promotion.Target != "main" || !matrix.Promotion.BranchProtectionRequired || !sameStringSet(matrix.ProofJobs, proofJobs) || !sameStringSet(matrix.RequiredContexts.Dev, proofJobs) || !sameStringSet(matrix.RequiredContexts.Main, proofJobs) {
		return governanceInput{}, fmt.Errorf("governance promotion contract is incomplete")
	}
	return governanceInput{Schema: matrix.Schema, RequiredContexts: governanceContexts{Dev: matrix.RequiredContexts.Dev, Main: matrix.RequiredContexts.Main}, ProofJobs: matrix.ProofJobs, Promotion: promotionInput{Source: matrix.Promotion.Source, Target: matrix.Promotion.Target, RequiredChecks: matrix.Promotion.RequiredChecks, BranchProtectionRequired: matrix.Promotion.BranchProtectionRequired}}, nil
}
