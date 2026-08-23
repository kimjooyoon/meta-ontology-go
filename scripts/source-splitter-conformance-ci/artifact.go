package main

import "encoding/json"

const (
	artifactSchema             = "gooo/source-splitter-external-conformance/v1"
	contractOperationID        = "gooo/meta/generation/SplitGo"
	registryOperationID        = "split-go-declarations"
	versionedDenominator       = 6
	decisionPass               = "PASS"
	decisionBlock              = "BLOCK"
	resolutionLower            = "LOWER_RESOLUTION"
	missingEvidenceScenario    = "missing-evidence"
	productionEvidenceScenario = "production-evidence"
)

type artifact struct {
	Schema               string   `json:"schema"`
	HeadSHA              string   `json:"head_sha"`
	ContractOperationID  string   `json:"contract_operation_id"`
	RegistryOperationID  string   `json:"registry_operation_id"`
	RequiredIndicatorIDs []string `json:"required_indicator_ids"`
	Denominator          int      `json:"denominator"`
	Actual               scenario `json:"actual"`
	MissingEvidence      scenario `json:"missing_evidence"`
}

type scenario struct {
	Name         string          `json:"name"`
	Decision     string          `json:"decision"`
	Resolution   string          `json:"resolution"`
	PassCount    int             `json:"pass_count"`
	FailCount    int             `json:"fail_count"`
	UnknownCount int             `json:"unknown_count"`
	Denominator  int             `json:"denominator"`
	Evaluation   json.RawMessage `json:"evaluation"`
}
