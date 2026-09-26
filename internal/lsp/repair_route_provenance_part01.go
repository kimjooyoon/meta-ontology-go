package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const RepairRouteObservationSchemaPart01 = "gooo/lsp/repair-route-provenance/v1"

type RepairRouteDecisionPart01 string

const (
	RepairRouteSourcePart01    RepairRouteDecisionPart01 = "source"
	RepairRouteParsePart01     RepairRouteDecisionPart01 = "parse"
	RepairRouteSemanticPart01  RepairRouteDecisionPart01 = "semantic"
	RepairRouteProfilePart01   RepairRouteDecisionPart01 = "profile"
	RepairRouteToolchainPart01 RepairRouteDecisionPart01 = "toolchain"
	RepairRouteContractPart01  RepairRouteDecisionPart01 = "contract"
	RepairRouteUnknownPart01   RepairRouteDecisionPart01 = "unknown"
)

type RepairRouteStatusPart01 string

const (
	RepairRoutePassPart01          RepairRouteStatusPart01 = "PASS"
	RepairRouteUnknownStatusPart01 RepairRouteStatusPart01 = "UNKNOWN"
)

type RepairRouteEvidenceInputPart01 struct {
	URI               string                    `json:"uri"`
	SourceDigest      string                    `json:"source_digest"`
	ParseDigest       string                    `json:"parse_digest,omitempty"`
	SemanticDigest    string                    `json:"semantic_digest,omitempty"`
	ProfileDigest     string                    `json:"profile_digest,omitempty"`
	ToolchainDigest   string                    `json:"toolchain_digest,omitempty"`
	ContractDigest    string                    `json:"contract_digest,omitempty"`
	Decision          RepairRouteDecisionPart01 `json:"decision"`
	DecisionSource    string                    `json:"decision_source"`
	Status            RepairRouteStatusPart01   `json:"status"`
	Reason            string                    `json:"reason,omitempty"`
	Confidence        float64                   `json:"confidence"`
	MissingStageIndex int                       `json:"missing_stage_index"`
}

type RepairRouteObservationPart01 struct {
	Schema               string                    `json:"schema"`
	URI                  string                    `json:"uri"`
	SourceDigest         string                    `json:"source_digest"`
	ParseDigest          string                    `json:"parse_digest,omitempty"`
	SemanticDigest       string                    `json:"semantic_digest,omitempty"`
	ProfileDigest        string                    `json:"profile_digest,omitempty"`
	ToolchainDigest      string                    `json:"toolchain_digest,omitempty"`
	ContractDigest       string                    `json:"contract_digest,omitempty"`
	Decision             RepairRouteDecisionPart01 `json:"decision"`
	DecisionSource       string                    `json:"decision_source"`
	Status               RepairRouteStatusPart01   `json:"status"`
	Reason               string                    `json:"reason,omitempty"`
	Confidence           float64                   `json:"confidence"`
	MissingStageIndex    int                       `json:"missing_stage_index"`
	EvidencePrefixDigest string                    `json:"evidence_prefix_digest"`
	ObservationDigest    string                    `json:"observation_digest"`
}

func NewRepairRouteObservationPart01(input RepairRouteEvidenceInputPart01) (RepairRouteObservationPart01, error) {
	observation := RepairRouteObservationPart01{
		Schema:            RepairRouteObservationSchemaPart01,
		URI:               input.URI,
		SourceDigest:      input.SourceDigest,
		ParseDigest:       input.ParseDigest,
		SemanticDigest:    input.SemanticDigest,
		ProfileDigest:     input.ProfileDigest,
		ToolchainDigest:   input.ToolchainDigest,
		ContractDigest:    input.ContractDigest,
		Decision:          input.Decision,
		DecisionSource:    input.DecisionSource,
		Status:            input.Status,
		Reason:            input.Reason,
		Confidence:        input.Confidence,
		MissingStageIndex: input.MissingStageIndex,
	}
	if err := observation.validateShapePart01(); err != nil {
		return RepairRouteObservationPart01{}, err
	}
	observation.EvidencePrefixDigest = digestRepairRoutePart01(observation.evidencePrefixPart01())
	observation.ObservationDigest = digestRepairRoutePart01(observation.observationViewPart01())
	return observation, nil
}

func (o RepairRouteObservationPart01) ValidPart01() bool {
	if o.validateShapePart01() != nil || !validRepairRouteDigestPart01(o.EvidencePrefixDigest) || !validRepairRouteDigestPart01(o.ObservationDigest) {
		return false
	}
	if o.EvidencePrefixDigest != digestRepairRoutePart01(o.evidencePrefixPart01()) {
		return false
	}
	return o.ObservationDigest == digestRepairRoutePart01(o.observationViewPart01())
}

func (o RepairRouteObservationPart01) validateShapePart01() error {
	if o.Schema != RepairRouteObservationSchemaPart01 || o.URI == "" || o.DecisionSource == "" {
		return errors.New("invalid repair route identity")
	}
	for name, value := range map[string]string{
		"source digest":    o.SourceDigest,
		"parse digest":     o.ParseDigest,
		"semantic digest":  o.SemanticDigest,
		"profile digest":   o.ProfileDigest,
		"toolchain digest": o.ToolchainDigest,
		"contract digest":  o.ContractDigest,
	} {
		if value != "" && !validRepairRouteDigestPart01(value) {
			return fmt.Errorf("invalid %s", name)
		}
	}
	if !validRepairRouteDigestPart01(o.SourceDigest) || o.Confidence < 0 || o.Confidence > 1 {
		return errors.New("invalid repair route evidence")
	}
	switch o.Decision {
	case RepairRouteSourcePart01, RepairRouteParsePart01, RepairRouteSemanticPart01, RepairRouteProfilePart01, RepairRouteToolchainPart01, RepairRouteContractPart01, RepairRouteUnknownPart01:
	default:
		return errors.New("invalid repair route decision")
	}
	switch o.Status {
	case RepairRoutePassPart01:
		if o.Decision == RepairRouteUnknownPart01 || o.MissingStageIndex != -1 {
			return errors.New("PASS requires a concrete decision and missing stage index -1")
		}
	case RepairRouteUnknownStatusPart01:
		if o.MissingStageIndex < 0 || o.Reason == "" {
			return errors.New("UNKNOWN requires reason and missing stage index")
		}
	default:
		return errors.New("invalid repair route status")
	}
	return nil
}

func (o RepairRouteObservationPart01) evidencePrefixPart01() any {
	return struct {
		Schema            string                    `json:"schema"`
		URI               string                    `json:"uri"`
		SourceDigest      string                    `json:"source_digest"`
		ParseDigest       string                    `json:"parse_digest,omitempty"`
		SemanticDigest    string                    `json:"semantic_digest,omitempty"`
		ProfileDigest     string                    `json:"profile_digest,omitempty"`
		ToolchainDigest   string                    `json:"toolchain_digest,omitempty"`
		ContractDigest    string                    `json:"contract_digest,omitempty"`
		Decision          RepairRouteDecisionPart01 `json:"decision"`
		DecisionSource    string                    `json:"decision_source"`
		Status            RepairRouteStatusPart01   `json:"status"`
		Reason            string                    `json:"reason,omitempty"`
		Confidence        float64                   `json:"confidence"`
		MissingStageIndex int                       `json:"missing_stage_index"`
	}{o.Schema, o.URI, o.SourceDigest, o.ParseDigest, o.SemanticDigest, o.ProfileDigest, o.ToolchainDigest, o.ContractDigest, o.Decision, o.DecisionSource, o.Status, o.Reason, o.Confidence, o.MissingStageIndex}
}

func (o RepairRouteObservationPart01) observationViewPart01() any {
	return struct {
		Prefix string `json:"evidence_prefix_digest"`
		Input  any    `json:"input"`
	}{o.EvidencePrefixDigest, o.evidencePrefixPart01()}
}

func digestRepairRoutePart01(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validRepairRouteDigestPart01(value string) bool {
	if len(value) != len("sha256:")+64 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
