package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementCausalFrontierLSPObservationSchema = "gooo.lsp.self-improvement.causal-frontier.v1"

type SelfImprovementCausalFrontierLSPObservation struct {
	Schema         string `json:"schema"`
	URI            string `json:"uri"`
	Version        int    `json:"version"`
	Declaration    string `json:"declaration"`
	Status         string `json:"status"`
	SelectedID     string `json:"selectedId"`
	SelectedState  string `json:"selectedState"`
	Visible        bool   `json:"visible"`
	Reason         string `json:"reason"`
	OriginDigest   string `json:"originDigest"`
	NonAuthorizing bool   `json:"nonAuthorizing"`
	Digest         string `json:"digest"`
}

func ObserveSelfImprovementCausalFrontierLSP(
	uri string,
	version int,
	declaration string,
	source valueexecution.SelfImprovementCausalFrontierObservation,
) SelfImprovementCausalFrontierLSPObservation {
	observation := SelfImprovementCausalFrontierLSPObservation{
		Schema:         SelfImprovementCausalFrontierLSPObservationSchema,
		URI:            uri,
		Version:        version,
		Declaration:    declaration,
		Status:         source.Status,
		SelectedID:     source.SelectedID,
		SelectedState:  source.SelectedState,
		OriginDigest:   source.Digest,
		NonAuthorizing: true,
	}

	switch {
	case strings.TrimSpace(uri) == "" || version < 0 || strings.TrimSpace(declaration) == "":
		observation.Status = valueexecution.SelfImprovementCausalFrontierUnknown
		observation.Reason = "LSP_CONTEXT_INVALID"
	case source.Status == valueexecution.SelfImprovementCausalFrontierUnknown || source.Digest == "":
		observation.Status = valueexecution.SelfImprovementCausalFrontierUnknown
		observation.Reason = "SOURCE_OBSERVATION_UNKNOWN"
	default:
		observation.Visible = true
		observation.Reason = "SOURCE_OBSERVATION_PROJECTED"
	}

	observation.Digest = digestSelfImprovementCausalFrontierLSP(observation)
	return observation
}

func digestSelfImprovementCausalFrontierLSP(
	observation SelfImprovementCausalFrontierLSPObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}