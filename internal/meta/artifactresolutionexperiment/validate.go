package artifactresolutionexperiment

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func validate(input Input) string {
	contract := input.Contract
	if len(input.SubjectSHA) != 40 || strings.Trim(input.SubjectSHA, "0123456789abcdef") != "" {
		return "SUBJECT_SHA_INVALID"
	}
	if contract.Schema != ContractSchema || contract.ID == "" ||
		contract.ManifestSchema != artifactemit.OperationManifestSchema ||
		contract.InterfaceSchema != artifactemit.OperationInterfaceSchema {
		return "CONTRACT_SCHEMA_INVALID"
	}
	if contract.ManifestDefinitions != 2 || contract.InterfaceDefinitions != 0 ||
		contract.RegisteredEmitters != 3 || contract.Indicators != ExpectedIndicators ||
		contract.NotClaimedCount != ExpectedNonClaims || len(contract.NotClaimed) != ExpectedNonClaims {
		return "CONTRACT_DENOMINATOR_INVALID"
	}
	return ""
}

func topDecisionUnknown(input Input) bool {
	decisions := []string{input.Manifest.Decision, input.ManifestReplay.Decision,
		input.ManifestGolden.Decision, input.Interface.Decision,
		input.InterfaceReplay.Decision, input.InterfaceGolden.Decision}
	for _, decision := range decisions {
		if decision != "PASS" {
			return true
		}
	}
	return false
}
