package main

import (
	"fmt"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func runVerification(program grant.PolicyProgram, settings options) error {
	if settings.grantRequestPath == "" || settings.resolutionPath == "" {
		return fmt.Errorf("-grant-request, -resolution, and -output are required for verify mode")
	}
	var request grant.GrantRequest
	var resolution grant.GrantResolution
	if err := readJSON(settings.grantRequestPath, &request); err != nil {
		return err
	}
	if err := readJSON(settings.resolutionPath, &resolution); err != nil {
		return err
	}
	input := grant.GrantInput{Request: request, DecisionInputs: resolution.DecisionInputs}
	verification := grant.Verify(program, input, resolution)
	if settings.check && !verification.Verified {
		return fmt.Errorf("grant verification failed: %#v", verification)
	}
	return writeJSON(settings.outputPath, verification)
}
