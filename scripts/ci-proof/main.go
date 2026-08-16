package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	governance := flag.String("governance", ".github/ci-governance.json", "governance matrix")
	evidence := flag.String("evidence", "ci-evidence-input/ci-evidence.json", "CI evidence bundle")
	jobs := flag.String("jobs", "ci-jobs.json", "workflow jobs JSON")
	scheduler := flag.String("scheduler", "ci-final-scheduler.json", "proof-side same-run scheduler result bindings")
	context := flag.String("context", "ci-proof-context.json", "GitHub proof context")
	generated := flag.String("generated", "ci-generated/first", "generated output directory")
	output := flag.String("output", "ci-proof.json", "proof bundle output")
	receipt := flag.String("receipt", "provenance-receipt.jsonl", "append-only receipt output")
	verify := flag.String("verify", "", "verify an existing proof bundle")
	requirePass := flag.Bool("require-pass", false, "reject a fail-closed decision")
	failureInput := flag.String("failure-input", "", "write a versioned CI failure manifest from an exact GitHub tuple")
	failureOutput := flag.String("failure-output", "ci-failure.json", "CI failure manifest output")
	closureInput := flag.String("closure-input", "", "write an exact-tuple no-failure closure manifest")
	closureOutput := flag.String("closure-output", "ci-closure.json", "no-failure closure manifest output")
	flag.Parse()
	var err error
	if *failureInput != "" {
		err = writeFailureManifest(*failureInput, *failureOutput)
	} else if *closureInput != "" {
		err = writeClosureManifest(*closureInput, *closureOutput)
	} else if *verify != "" {
		err = verifyProof(*verify, *governance, *receipt, *requirePass)
	} else {
		err = generateProof(*root, *governance, *evidence, *jobs, *scheduler, *context, *generated, *output, *receipt)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generateProof(root, governance, evidence, jobs, scheduler, context, generated, output, receipt string) error {
	inputs, err := readInputs(root, governance, evidence, jobs, scheduler, context)
	if err != nil {
		return err
	}
	digests, err := computeProofDigests(root, generated, inputs.Evidence)
	if err != nil {
		return err
	}
	bundle, record, err := buildProof(inputs, digests)
	if err != nil {
		return err
	}
	if err := validateProof(bundle); err != nil {
		return err
	}
	return writeOutputs(output, receipt, bundle, record)
}

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
