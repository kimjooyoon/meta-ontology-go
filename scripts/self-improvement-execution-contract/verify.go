package main

import (
	"fmt"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func runVerification(settings options) error {
	if settings.resolutionPath == "" {
		return fmt.Errorf("-resolution is required for verify mode")
	}
	var resolution contract.ContractResolution
	if err := readJSON(settings.resolutionPath, &resolution); err != nil {
		return err
	}
	if err := contract.VerifyResolution(resolution); err != nil {
		return err
	}
	verification := contract.Verification{
		Schema: contract.VerificationSchema, ContractDigest: resolution.Digest,
		IndependentDecision: resolution.Decision, IndependentResolution: resolution.Resolution,
		IndependentReason: resolution.Reason, Verified: true, IndependentReplayComparisons: 1,
	}
	verification.Digest = verificationDigest(verification)
	return writeJSON(settings.outputPath, verification)
}
