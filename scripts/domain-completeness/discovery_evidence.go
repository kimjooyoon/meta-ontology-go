package main

import (
	"bytes"
	"encoding/json"
	"reflect"

	jev "github.com/kimjooyoon/gooo-jev/gooo"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
)

const (
	capabilityDiscoverySchema       = "gooo/capability-discovery-report/v1"
	capabilityDiscoveryProfile      = "gooo-jev-capability-discovery/v1"
	capabilityDiscoveryQuery        = "How do I generate a canonical .gooo declaration?"
	capabilityDiscoverySourcePath   = "examples/capability-discovery/current.gooo.fixture"
	capabilityDiscoveryContractPath = "examples/capability-discovery/domain-contract.gooo.fixture"
	capabilityDiscoveryReportPath   = "evidence/capability-discovery.json"
	capabilityDiscoveryReplayPath   = "evidence/capability-discovery-replay.json"
	capabilityDiscoverySourceCopy   = "evidence/capability-discovery-source.gooo"
	capabilityDiscoveryContractCopy = "evidence/capability-discovery-domain-contract.gooo"
)

type capabilityDiscoveryReport struct {
	Schema                 string                            `json:"schema"`
	Decision               string                            `json:"decision"`
	Trail                  jev.CapabilityQueryTrail          `json:"capability_trail"`
	SourcePath             string                            `json:"source_path"`
	SourceDigest           string                            `json:"source_digest"`
	SemanticDigest         string                            `json:"semantic_digest"`
	DomainContractPath     string                            `json:"domain_contract_path"`
	DomainContractDigest   string                            `json:"domain_contract_digest"`
	DomainContractSemantic string                            `json:"domain_contract_semantic_digest"`
	Receipt                *completeness.CompletenessReceipt `json:"completeness_receipt"`
}

type capabilityDiscoveryEvidence struct {
	Report capabilityDiscoveryReport
	State  string
	Reason string
	Refs   []EvidenceRef
}

func validateCapabilityDiscoveryEvidence(reportRaw, replayRaw, sourceRaw, contractRaw []byte) capabilityDiscoveryEvidence {
	result := capabilityDiscoveryEvidence{State: "UNKNOWN", Reason: "CAPABILITY_DISCOVERY_EVIDENCE_MISSING"}
	result.Refs = discoveryEvidenceRefs(reportRaw, replayRaw, sourceRaw, contractRaw)
	if len(reportRaw) == 0 || len(replayRaw) == 0 || len(sourceRaw) == 0 || len(contractRaw) == 0 {
		return result
	}
	if !bytes.Equal(reportRaw, replayRaw) {
		result.State, result.Reason = "FAIL_CLOSED", "CAPABILITY_DISCOVERY_REPLAY_MISMATCH"
		return result
	}
	var report capabilityDiscoveryReport
	if err := json.Unmarshal(reportRaw, &report); err != nil {
		result.State, result.Reason = "FAIL_CLOSED", "CAPABILITY_DISCOVERY_REPORT_INVALID"
		return result
	}
	result.Report = report
	fail := func(reason string) capabilityDiscoveryEvidence {
		result.State, result.Reason = "FAIL_CLOSED", reason
		return result
	}
	if report.Schema != capabilityDiscoverySchema || report.Decision != "PROGRESS" ||
		report.SourcePath != capabilityDiscoverySourcePath || report.DomainContractPath != capabilityDiscoveryContractPath {
		return fail("CAPABILITY_DISCOVERY_REPORT_IDENTITY_MISMATCH")
	}
	if report.SourceDigest != digestBytes(sourceRaw) || report.DomainContractDigest != digestBytes(contractRaw) {
		return fail("CAPABILITY_DISCOVERY_SOURCE_DIGEST_MISMATCH")
	}
	sourceSemantic, err := semanticHash(capabilityDiscoverySourcePath, sourceRaw)
	if err != nil || sourceSemantic != report.SemanticDigest {
		return fail("CAPABILITY_DISCOVERY_SOURCE_SEMANTIC_MISMATCH")
	}
	contractSemantic, err := semanticHash(capabilityDiscoveryContractPath, contractRaw)
	if err != nil || contractSemantic != report.DomainContractSemantic {
		return fail("CAPABILITY_DISCOVERY_CONTRACT_SEMANTIC_MISMATCH")
	}
	if err := report.Trail.Validate(); err != nil {
		return fail("CAPABILITY_DISCOVERY_TRAIL_INVALID")
	}
	if report.Trail.Response.Query != capabilityDiscoveryQuery {
		return fail("CAPABILITY_DISCOVERY_QUERY_MISMATCH")
	}
	expectedTrail := jev.DiscoverCapabilityQueryTrail(capabilityDiscoveryQuery, string(sourceRaw))
	if !reflect.DeepEqual(report.Trail, expectedTrail) {
		return fail("CAPABILITY_DISCOVERY_TRAIL_SOURCE_MISMATCH")
	}
	declaration := report.Trail.Response.Declaration
	if declaration == nil {
		result.State, result.Reason = "UNKNOWN", "CAPABILITY_DISCOVERY_DECLARATION_NOT_BOUND"
		return result
	}
	if !declaration.Bound {
		result.State, result.Reason = "UNKNOWN", "CAPABILITY_DISCOVERY_DECLARATION_NOT_BOUND"
		return result
	}
	if report.Receipt == nil || completeness.Validate(report.Receipt) != nil ||
		report.Receipt.ProfileID != capabilityDiscoveryProfile || report.Receipt.Decision != report.Decision {
		return fail("CAPABILITY_DISCOVERY_COMPLETENESS_RECEIPT_INVALID")
	}
	if !discoveryReceiptBindsReport(report) {
		return fail("CAPABILITY_DISCOVERY_RECEIPT_BINDING_MISMATCH")
	}
	if !discoveryReceiptDimensionsAgree(report) {
		return fail("CAPABILITY_DISCOVERY_RECEIPT_DIMENSION_MISMATCH")
	}
	result.State, result.Reason = "PASS", "CAPABILITY_DISCOVERY_EXACT_AND_REPLAYED"
	return result
}

func validateCapabilityDiscoveryCells(
	observation languageutility.Observation,
	reportRaw, replayRaw []byte,
) (string, string) {
	expected := make(map[string]languageutility.CellObservation, len(languageutility.CanonicalStages))
	for _, cell := range observation.Cells {
		if cell.UseCaseID != "capability-discovery" {
			continue
		}
		if _, duplicate := expected[cell.StageID]; duplicate {
			return "FAIL_CLOSED", "CAPABILITY_DISCOVERY_CELL_DUPLICATE"
		}
		expected[cell.StageID] = cell
	}
	if len(expected) != len(languageutility.CanonicalStages) {
		return "UNKNOWN", "CAPABILITY_DISCOVERY_CELLS_MISSING"
	}
	reportDigest := digestBytes(reportRaw)
	replayDigest := digestBytes(replayRaw)
	for _, stage := range languageutility.CanonicalStages {
		cell, exists := expected[stage.ID]
		if !exists {
			return "UNKNOWN", "CAPABILITY_DISCOVERY_CELLS_MISSING"
		}
		if stage.ID == "RESOURCE_OBSERVED" {
			if cell.State != languageutility.StateOpen || cell.Producer != "ci:capability-discovery" ||
				cell.Step != "COLLECT_RESOURCE_OBSERVED" || cell.Reason != "DISCOVERY_RESOURCES_NOT_MEASURED" ||
				cell.EvidencePath != "" || cell.EvidenceDigest != "" {
				return "FAIL_CLOSED", "CAPABILITY_DISCOVERY_RESOURCE_CLAIM_UNSUPPORTED"
			}
			continue
		}
		if cell.State != languageutility.StateClosed {
			if cell.State == languageutility.StateRefuted {
				return "FAIL_CLOSED", "CAPABILITY_DISCOVERY_CELL_REFUTED"
			}
			return "UNKNOWN", "CAPABILITY_DISCOVERY_CELL_NOT_CLOSED"
		}
		wantDigest, wantPath := reportDigest, capabilityDiscoveryReportPath
		if stage.ID == "DETERMINISTIC_REPLAY" {
			wantDigest, wantPath = replayDigest, capabilityDiscoveryReplayPath
		}
		if cell.Producer != "scripts/language-utility-evidence" || cell.Step != "VERIFY_"+stage.ID ||
			cell.Reason != "EVIDENCE_ACCEPTED" || cell.EvidenceKey != "capability-discovery.report" ||
			cell.EvidencePath != wantPath || cell.EvidenceDigest != wantDigest {
			return "FAIL_CLOSED", "CAPABILITY_DISCOVERY_CELL_EVIDENCE_MISMATCH"
		}
	}
	return "PASS", "CAPABILITY_DISCOVERY_CELLS_MATCH_BOUND_RECEIPT"
}

func discoveryEvidenceRefs(reportRaw, replayRaw, sourceRaw, contractRaw []byte) []EvidenceRef {
	values := []struct {
		role string
		path string
		data []byte
	}{
		{role: "capability-discovery-report", path: capabilityDiscoveryReportPath, data: reportRaw},
		{role: "capability-discovery-replay", path: capabilityDiscoveryReplayPath, data: replayRaw},
		{role: "capability-discovery-source", path: capabilityDiscoverySourceCopy, data: sourceRaw},
		{role: "capability-discovery-domain-contract", path: capabilityDiscoveryContractCopy, data: contractRaw},
	}
	refs := make([]EvidenceRef, 0, len(values))
	for _, value := range values {
		if len(value.data) > 0 {
			refs = append(refs, EvidenceRef{Role: value.role, Path: value.path, Digest: digestBytes(value.data)})
		}
	}
	return refs
}

func discoveryReceiptBindsReport(report capabilityDiscoveryReport) bool {
	if report.Receipt == nil {
		return false
	}
	scope := report.Receipt.Scope
	if scopeString(scope, "source_path") != report.SourcePath ||
		scopeString(scope, "source_digest") != report.SourceDigest ||
		scopeString(scope, "semantic_digest") != report.SemanticDigest ||
		scopeString(scope, "query_digest") != report.Trail.Response.QueryDigest ||
		scopeString(scope, "jev_evidence_digest") != report.Trail.EvidenceDigest ||
		scopeString(scope, "capability_status") != string(report.Trail.Response.Status) {
		return false
	}
	contractScope, ok := scope["domain_contract"].(map[string]any)
	return ok && contractScope["provided"] == true &&
		contractScope["path"] == report.DomainContractPath &&
		contractScope["source_digest"] == report.DomainContractDigest &&
		contractScope["semantic_digest"] == report.DomainContractSemantic
}

func discoveryReceiptDimensionsAgree(report capabilityDiscoveryReport) bool {
	if report.Receipt == nil {
		return false
	}
	dimensions := make(map[string]completeness.CompletenessDimension, len(report.Receipt.Dimensions))
	for _, dimension := range report.Receipt.Dimensions {
		dimensions[dimension.ID] = dimension
	}
	declaration := dimensions["declaration_coverage"]
	contractScope, _ := report.Receipt.Scope["domain_contract"].(map[string]any)
	expectedDeclarations, ok := contractScope["expected_declarations"].(float64)
	if !ok || declaration.Denominator != int(expectedDeclarations) || declaration.Numerator < 0 ||
		declaration.Numerator > declaration.Denominator {
		return false
	}
	if !dimensionMatches(dimensions["capability_discovery_observation"], "PASS", 1, 1) {
		return false
	}
	match := dimensions["catalog_match"]
	wantMatchStatus, wantMatchNumerator := discoveryCatalogMatch(report.Trail.Response.Status)
	if !dimensionMatches(match, wantMatchStatus, wantMatchNumerator, 1) {
		return false
	}
	for _, id := range []string{"generation_coverage", "reverse_observation_coverage", "real_use_case_coverage", "permission_boundary_coverage", "network_boundary_coverage"} {
		if !dimensionMatches(dimensions[id], "UNKNOWN", 0, 1) {
			return false
		}
	}
	boundaryStatus, boundaryNumerator := "FAIL_CLOSED", 0
	if report.Trail.Response.NonExecuting && report.Trail.Response.NonAuthorizing {
		boundaryStatus, boundaryNumerator = "PASS", 1
	}
	if !dimensionMatches(dimensions["execution_boundary_coverage"], boundaryStatus, boundaryNumerator, 1) {
		return false
	}
	return dimensionMatches(dimensions["provenance_integrity"], "PASS", 1, 1)
}

func discoveryCatalogMatch(status jev.CapabilityQueryState) (string, int) {
	switch status {
	case jev.CapabilityQueryAvailable:
		return "PASS", 1
	case jev.CapabilityQueryDeferred:
		return "PROGRESS", 0
	default:
		return "UNKNOWN", 0
	}
}

func dimensionMatches(value completeness.CompletenessDimension, status string, numerator, denominator int) bool {
	return value.Status == status && value.Numerator == numerator && value.Denominator == denominator
}

func scopeString(scope map[string]any, key string) string {
	value, _ := scope[key].(string)
	return value
}

func requireDiscoveryReceiptDimension(report capabilityDiscoveryReport, id string) (completeness.CompletenessDimension, bool) {
	if report.Receipt == nil {
		return completeness.CompletenessDimension{}, false
	}
	for _, dimension := range report.Receipt.Dimensions {
		if dimension.ID == id {
			return dimension, true
		}
	}
	return completeness.CompletenessDimension{}, false
}
