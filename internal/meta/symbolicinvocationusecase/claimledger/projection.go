package claimledger

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func Project(contractData, observationData []byte, subject string) (Report, error) {
	var contract Contract
	if err := json.Unmarshal(contractData, &contract); err != nil {
		return Report{}, fmt.Errorf("decode contract: %w", err)
	}
	if err := validateContract(contract, subject); err != nil {
		return Report{}, err
	}
	observation := map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(observationData))
	decoder.UseNumber()
	if err := decoder.Decode(&observation); err != nil {
		return Report{}, fmt.Errorf("decode observation: %w", err)
	}
	report := Report{
		Schema: ReportSchema, Subject: subject, Metric: contract.Metric,
		ContractDigest: digestBytes(contractData), ObservationDigest: digestBytes(observationData),
		OpenClaimIDs: []string{}, Claims: []Claim{}, Evidence: []Evidence{}, Events: []Event{},
	}
	for _, spec := range contract.Claims {
		projectClaim(&report, spec, observation, subject)
	}
	finalize(&report, contract.Expected)
	return report, nil
}

func projectClaim(report *Report, spec ClaimSpec, observation map[string]any, subject string) {
	claim := Claim{
		ID: spec.ID, Kind: spec.Kind, Modality: spec.Modality, Subject: spec.Subject,
		Predicate: spec.Predicate, Scope: spec.Scope, ProofRoute: spec.ProofRoute,
		Coordinate: spec.Coordinate, EvidenceRefs: []string{},
	}
	countProofRoute(&report.Metrics.ProofRoutes, spec.ProofRoute)
	addEvent(report, spec, "CLAIM_REGISTERED", "OPEN", "CLAIM_DECLARED", "")
	if spec.Scope == "EXCLUDED" {
		claim.Status, claim.Truth, claim.Reason = "EXCLUDED", "NOT_EVALUATED", spec.ExcludedReason
		report.Metrics.ExcludedTotal++
		report.Claims = append(report.Claims, claim)
		addEvent(report, spec, "CLAIM_EXCLUDED", claim.Status, claim.Reason, "")
		return
	}
	projectEvidence(report, claim, spec, observation, subject)
}
