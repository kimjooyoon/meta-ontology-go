package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func runLive(program contract.PolicyProgram, settings options) error {
	input := contract.ContractInput{Registry: contract.KnownRegistry()}
	if settings.requestPath != "" {
		var request selfimprovementcandidate.AuthorizationRequest
		if err := readJSON(settings.requestPath, &request); err == nil {
			input = contract.ProjectAuthorizationRequest(request, contract.KnownRegistry())
		}
	}
	resolution := contract.Evaluate(program, input)
	report := contract.LiveReport{ContractResolution: resolution, Verification: contract.Verify(program, input, resolution)}
	if settings.check {
		if err := contract.VerifyResolution(report.ContractResolution); err != nil || !report.Verification.Verified {
			return fmt.Errorf("live contract check failed: resolution=%v verification=%v", err, report.Verification)
		}
	}
	return writeJSON(settings.outputPath, report)
}
