package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

type domainReceipt struct {
	DomainID           string              `json:"domain_id"`
	DomainVersion      string              `json:"domain_version"`
	CorrelationID       string              `json:"correlation_id"`
	ObservedAt         string              `json:"observed_at"`
	DeclarationDigest  string              `json:"declaration_digest"`
	SourceDigest       string              `json:"source_digest"`
	CatalogDigest      string              `json:"catalog_digest"`
	SchemaDigest       string              `json:"schema_digest"`
	EvidenceDigest     string              `json:"evidence_digest"`
	ToolchainIdentity  string              `json:"toolchain_identity"`
	DeclaredIdentifiers []string           `json:"declared_identifiers"`
	Observations       []domainObservation `json:"observations"`
	Metrics            domainMetrics       `json:"metrics"`
	Provenance         domainProvenance    `json:"provenance"`
	Execution          bool                `json:"execution"`
	Authorization      bool                `json:"authorization"`
	Investment         *domainInvestment   `json:"investment,omitempty"`
}

type domainObservation struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

type boundedMetric struct {
	Value              float64  `json:"value"`
	Numerator          int      `json:"numerator"`
	Denominator        int      `json:"denominator"`
	ExcludedIdentifiers []string `json:"excluded_identifiers,omitempty"`
}

type domainMetrics struct {
	ScopeCoverage       boundedMetric `json:"scope_coverage"`
	AvailableCoverage   boundedMetric `json:"available_coverage"`
	EvidenceCompleteness boundedMetric `json:"evidence_completeness"`
	UtilityYield        boundedMetric `json:"utility_yield"`
}

type domainProvenance struct {
	FirstMissingStage int      `json:"first_missing_stage"`
	NextOperation     string   `json:"next_operation"`
	LinkedDigests     []string `json:"linked_digests,omitempty"`
}

type domainInvestment struct {
	EngineeringEffort  float64 `json:"engineering_effort,omitempty"`
	CIDurationSeconds  float64 `json:"ci_duration_seconds,omitempty"`
	CIAttempts         int     `json:"ci_attempts,omitempty"`
	ResourceCost       float64 `json:"resource_cost,omitempty"`
	DecisionReason     string  `json:"decision_reason,omitempty"`
	RollbackCondition  string  `json:"rollback_condition,omitempty"`
}

type validationReport struct {
	Valid        bool   `json:"valid"`
	ReceiptDigest string `json:"receipt_digest,omitempty"`
	Error        string `json:"error,omitempty"`
}

func main() {
	path := "-"
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else if len(os.Args) > 2 {
		fail(fmt.Errorf("usage: domain-completeness [receipt.json|-]"))
	}
	payload, err := readInput(path)
	if err != nil {
		fail(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	var receipt domainReceipt
	if err := decoder.Decode(&receipt); err != nil {
		writeReport(validationReport{Error: fmt.Sprintf("decode receipt: %v", err)})
		os.Exit(2)
	}
	if err := receipt.Validate(); err != nil {
		writeReport(validationReport{Error: err.Error()})
		os.Exit(2)
	}
	digest, err := receipt.Digest()
	if err != nil {
		fail(err)
	}
	writeReport(validationReport{Valid: true, ReceiptDigest: digest})
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func (receipt domainReceipt) Validate() error {
	for name, value := range map[string]string{
		"declaration_digest": receipt.DeclarationDigest,
		"source_digest": receipt.SourceDigest,
		"catalog_digest": receipt.CatalogDigest,
		"schema_digest": receipt.SchemaDigest,
		"evidence_digest": receipt.EvidenceDigest,
	} {
		if !validDigest(value) {
			return fmt.Errorf("invalid %s", name)
		}
	}
	for name, value := range map[string]string{
		"domain_id": receipt.DomainID,
		"domain_version": receipt.DomainVersion,
		"correlation_id": receipt.CorrelationID,
		"toolchain_identity": receipt.ToolchainIdentity,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if _, err := time.Parse(time.RFC3339, receipt.ObservedAt); err != nil {
		return fmt.Errorf("observed_at is not RFC3339: %w", err)
	}
	if receipt.Execution || receipt.Authorization {
		return fmt.Errorf("domain receipt cannot execute or authorize")
	}
	if len(receipt.DeclaredIdentifiers) == 0 || len(receipt.Observations) == 0 {
		return fmt.Errorf("domain receipt requires declarations and observations")
	}
	declared := make(map[string]struct{}, len(receipt.DeclaredIdentifiers))
	for _, id := range receipt.DeclaredIdentifiers {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("declared identifier is empty")
		}
		if _, exists := declared[id]; exists {
			return fmt.Errorf("duplicate declared identifier %q", id)
		}
		declared[id] = struct{}{}
	}
	available := 0
	observed := make(map[string]struct{}, len(receipt.Observations))
	for _, observation := range receipt.Observations {
		if strings.TrimSpace(observation.ID) == "" {
			return fmt.Errorf("observation id is empty")
		}
		if _, exists := declared[observation.ID]; !exists {
			return fmt.Errorf("observation %q is not declared", observation.ID)
		}
		if _, exists := observed[observation.ID]; exists {
			return fmt.Errorf("duplicate observation %q", observation.ID)
		}
		observed[observation.ID] = struct{}{}
		switch observation.State {
		case "AVAILABLE":
			available++
		case "DEFERRED", "UNKNOWN", "OUT_OF_SCOPE":
			if strings.TrimSpace(observation.Reason) == "" {
				return fmt.Errorf("observation %q requires a reason", observation.ID)
			}
		default:
			return fmt.Errorf("observation %q has invalid state %q", observation.ID, observation.State)
		}
	}
	if receipt.Metrics.ScopeCoverage.Denominator != len(receipt.DeclaredIdentifiers) || receipt.Metrics.ScopeCoverage.Numerator != len(receipt.Observations) {
		return fmt.Errorf("scope coverage is not bound to declarations and observations")
	}
	if receipt.Metrics.AvailableCoverage.Denominator != len(receipt.DeclaredIdentifiers) || receipt.Metrics.AvailableCoverage.Numerator != available {
		return fmt.Errorf("available coverage is not bound to AVAILABLE observations")
	}
	for name, metric := range map[string]boundedMetric{
		"scope_coverage": receipt.Metrics.ScopeCoverage,
		"available_coverage": receipt.Metrics.AvailableCoverage,
		"evidence_completeness": receipt.Metrics.EvidenceCompleteness,
		"utility_yield": receipt.Metrics.UtilityYield,
	} {
		if err := validateMetric(name, metric); err != nil {
			return err
		}
	}
	if receipt.Metrics.AvailableCoverage.Numerator > receipt.Metrics.ScopeCoverage.Numerator {
		return fmt.Errorf("available coverage exceeds observed scope")
	}
	if receipt.Provenance.FirstMissingStage < -1 {
		return fmt.Errorf("first_missing_stage must be -1 or greater")
	}
	if receipt.Provenance.FirstMissingStage == -1 && strings.TrimSpace(receipt.Provenance.NextOperation) != "" {
		return fmt.Errorf("complete provenance cannot carry next_operation")
	}
	if receipt.Provenance.FirstMissingStage >= 0 && strings.TrimSpace(receipt.Provenance.NextOperation) == "" {
		return fmt.Errorf("incomplete provenance requires next_operation")
	}
	for _, digest := range receipt.Provenance.LinkedDigests {
		if !validDigest(digest) {
			return fmt.Errorf("invalid linked provenance digest")
		}
	}
	if receipt.Investment != nil {
		if receipt.Investment.EngineeringEffort < 0 || receipt.Investment.CIDurationSeconds < 0 || receipt.Investment.CIAttempts < 0 || receipt.Investment.ResourceCost < 0 {
			return fmt.Errorf("investment cost cannot be negative")
		}
		if strings.TrimSpace(receipt.Investment.DecisionReason) == "" || strings.TrimSpace(receipt.Investment.RollbackCondition) == "" {
			return fmt.Errorf("investment requires decision_reason and rollback_condition")
		}
	}
	return nil
}

func validateMetric(name string, metric boundedMetric) error {
	if metric.Numerator < 0 || metric.Denominator < 0 || metric.Numerator > metric.Denominator {
		return fmt.Errorf("%s has invalid numerator or denominator", name)
	}
	if metric.Value < 0 || metric.Value > 1 || math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
		return fmt.Errorf("%s value is outside [0,1]", name)
	}
	expected := 0.0
	if metric.Denominator > 0 {
		expected = float64(metric.Numerator) / float64(metric.Denominator)
	}
	if math.Abs(metric.Value-expected) > 1e-9 {
		return fmt.Errorf("%s value is not numerator/denominator", name)
	}
	seen := make(map[string]struct{}, len(metric.ExcludedIdentifiers))
	for _, id := range metric.ExcludedIdentifiers {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("%s has an empty excluded identifier", name)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("%s has duplicate excluded identifier %q", name, id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func (receipt domainReceipt) Digest() (string, error) {
	payload, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func writeReport(report validationReport) {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(err)
	}
	fmt.Println(string(payload))
}

func fail(err error) {
	writeReport(validationReport{Error: err.Error()})
	os.Exit(2)
}