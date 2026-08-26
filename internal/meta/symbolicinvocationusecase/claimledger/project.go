package claimledger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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
		Schema:            ReportSchema,
		Subject:           subject,
		Metric:            contract.Metric,
		ContractDigest:    digestBytes(contractData),
		ObservationDigest: digestBytes(observationData),
		OpenClaimIDs:      []string{},
		Claims:            []Claim{},
		Evidence:          []Evidence{},
		Events:            []Event{},
	}
	for _, spec := range contract.Claims {
		projectClaim(&report, spec, observation, subject)
	}
	finalize(&report, contract.Expected)
	return report, nil
}

func projectClaim(report *Report, spec ClaimSpec, observation map[string]any, subject string) {
	claim := Claim{
		ID: spec.ID, Kind: spec.Kind, Modality: spec.Modality,
		Subject: spec.Subject, Predicate: spec.Predicate, Scope: spec.Scope,
		ProofRoute: spec.ProofRoute, Coordinate: spec.Coordinate,
		EvidenceRefs: []string{},
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

	report.Metrics.InScopeClaimTotal++
	evidenceID := "evidence:" + spec.ID
	claim.EvidenceRefs = []string{evidenceID}
	path, observed, found := lookupAny(observation, spec.Evidence.Paths)
	evidence := Evidence{
		ID: evidenceID, ClaimID: spec.ID, SourcePath: strings.Join(spec.Evidence.Paths, "|"),
		SourceDigest: report.ObservationDigest,
	}
	if !found {
		evidence.Status = "MISSING"
		claim.Status, claim.Truth, claim.Reason = "UNKNOWN", "UNDETERMINED", spec.UnknownReason
		report.Metrics.UnknownTotal++
		report.OpenClaimIDs = append(report.OpenClaimIDs, spec.ID)
		report.Evidence = append(report.Evidence, evidence)
		addEvent(report, spec, "EVIDENCE_MISSING", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_UNKNOWN", claim.Status, claim.Reason, evidenceID)
		report.Claims = append(report.Claims, claim)
		return
	}

	evidence.SourcePath = path
	matched, expectedDigest := evidenceMatches(*spec.Evidence, observed, subject)
	evidence.ObservedValueDigest = digestValue(observed)
	evidence.ExpectedValueDigest = expectedDigest
	if matched {
		evidence.Status = "VERIFIED"
		claim.Status, claim.Truth, claim.Reason = "DISCHARGED", "SATISFIED", "EVIDENCE_VERIFIED"
		report.Metrics.DischargedTotal++
		addEvent(report, spec, "EVIDENCE_VERIFIED", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_DISCHARGED", claim.Status, claim.Reason, evidenceID)
	} else {
		evidence.Status = "FAILED"
		claim.Status, claim.Truth, claim.Reason = "REFUTED", "VIOLATED", spec.RefutedReason
		report.Metrics.RefutedTotal++
		addEvent(report, spec, "EVIDENCE_REJECTED", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_REFUTED", claim.Status, claim.Reason, evidenceID)
	}
	report.Evidence = append(report.Evidence, evidence)
	report.Claims = append(report.Claims, claim)
}

func finalize(report *Report, expected ExpectedMetrics) {
	report.Metrics.FixedClaimTotal = len(report.Claims)
	report.Metrics.OpenClaimTotal = len(report.OpenClaimIDs)
	if report.Metrics.InScopeClaimTotal > 0 {
		report.Metrics.DischargeBasisPoints = report.Metrics.DischargedTotal * 10_000 / report.Metrics.InScopeClaimTotal
	}

	switch {
	case report.Metrics.InScopeClaimTotal == 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "NONE", "NO_SCOPED_CLAIMS"}
	case report.Metrics.UnknownTotal > 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "STAGE_LOCAL", "OPEN_UNKNOWN_CLAIMS_REMAIN"}
	case report.Metrics.RefutedTotal > 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "CLAIM_LOCAL", "REFUTED_CLAIMS_PRESENT"}
	default:
		report.ClaimSet = Verdict{"PASS", "EXACT", "ALL_SCOPED_CLAIMS_DISCHARGED"}
	}
	if report.ClaimSet.Decision == "PASS" && (report.Metrics.UnknownTotal > 0 || report.Metrics.RefutedTotal > 0) {
		report.Metrics.FalsePromotionCount = 1
	}
	if metricsMatch(report, expected) {
		report.Conformance = Conformance{"PASS", "FIXED_METRIC_CONTRACT_MATCHED"}
	} else {
		report.Conformance = Conformance{"FAIL_CLOSED", "FIXED_METRIC_CONTRACT_MISMATCH"}
	}
}

func metricsMatch(report *Report, expected ExpectedMetrics) bool {
	m := report.Metrics
	return m.FixedClaimTotal == expected.FixedClaimTotal &&
		m.InScopeClaimTotal == expected.InScopeClaimTotal &&
		m.DischargedTotal == expected.DischargedTotal &&
		m.UnknownTotal == expected.UnknownTotal &&
		m.RefutedTotal == expected.RefutedTotal &&
		m.ExcludedTotal == expected.ExcludedTotal &&
		m.OpenClaimTotal == expected.OpenClaimTotal &&
		m.DischargeBasisPoints == expected.DischargeBasisPoints &&
		m.FalsePromotionCount == expected.FalsePromotionCount &&
		m.ProofRoutes == expected.ProofRoutes &&
		report.ClaimSet.Decision == expected.ClaimSetDecision &&
		report.ClaimSet.Resolution == expected.Resolution
}

func addEvent(report *Report, spec ClaimSpec, kind, status, reason, evidenceID string) {
	report.Events = append(report.Events, Event{
		Sequence: len(report.Events) + 1, Type: kind, ClaimID: spec.ID,
		EvidenceID: evidenceID, Status: status, Coordinate: spec.Coordinate, Reason: reason,
	})
}
