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
	if matrix.Schema != "gooo/ci-governance/v2" || matrix.Promotion.Source != "dev" || matrix.Promotion.Target != "main" || !matrix.Promotion.BranchProtectionRequired || !sameStringSet(matrix.ProofJobs, proofJobs) || !sameStringSet(matrix.RequiredContexts.Dev, append(append([]string(nil), proofJobs...), "CI guardian shadow")) || !sameStringSet(matrix.RequiredContexts.Main, append(append([]string(nil), proofJobs...), "CI guardian")) || matrix.GuardianContexts.DevShadow != "CI guardian shadow" || matrix.GuardianContexts.MainRequired != "CI guardian" {
		return governanceInput{}, fmt.Errorf("governance promotion contract is incomplete")
	}
	return governanceInput{Schema: matrix.Schema, RequiredContexts: governanceContexts{Dev: matrix.RequiredContexts.Dev, Main: matrix.RequiredContexts.Main}, GuardianContexts: guardianContexts{DevShadow: matrix.GuardianContexts.DevShadow, MainRequired: matrix.GuardianContexts.MainRequired}, ProofJobs: matrix.ProofJobs, Promotion: promotionInput{Source: matrix.Promotion.Source, Target: matrix.Promotion.Target, RequiredChecks: matrix.Promotion.RequiredChecks, BranchProtectionRequired: matrix.Promotion.BranchProtectionRequired}}, nil
}
