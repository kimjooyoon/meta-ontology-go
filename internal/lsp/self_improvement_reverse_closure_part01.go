package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementReverseClosureLSPObservationSchema = "gooo.lsp.self-improvement.reverse-closure.v1"

type SelfImprovementReverseClosureLSPObservation struct {
	Schema            string `json:"schema"`
	URI               string `json:"uri"`
	Version           int    `json:"version"`
	Declaration       string `json:"declaration"`
	Status            string `json:"status"`
	Outcome           string `json:"outcome"`
	Visible           bool   `json:"visible"`
	Reason            string `json:"reason"`
	ReverseObservation bool  `json:"reverseObservation"`
	OriginDigest      string `json:"originDigest"`
	NonAuthorizing    bool   `json:"nonAuthorizing"`
	Digest            string `json:"digest"`
}

func ObserveSelfImprovementReverseClosureLSP(
	uri string,
	version int,
	declaration string,
	source valueexecution.SelfImprovementReverseClosureObservation,
) SelfImprovementReverseClosureLSPObservation {
	observation := SelfImprovementReverseClosureLSPObservation{
		Schema:             SelfImprovementReverseClosureLSPObservationSchema,
		URI:                uri,
		Version:            version,
		Declaration:        declaration,
		Status:             source.Status,
		Outcome:            source.Outcome,
		ReverseObservation: source.ReverseObservation,
		OriginDigest:       source.Digest,
		NonAuthorizing:     true,
	}

	switch {
	case strings.TrimSpace(uri) == "" || version < 0 || strings.TrimSpace(declaration) == "":
		observation.Status = valueexecution.SelfImprovementReverseClosureUnknown
		observation.Reason = "LSP_CONTEXT_INVALID"
	case source.Status == valueexecution.SelfImprovementReverseClosureUnknown || source.Digest == "":
		observation.Status = valueexecution.SelfImprovementReverseClosureUnknown
		observation.Reason = "SOURCE_OBSERVATION_UNKNOWN"
	default:
		observation.Visible = true
		observation.Reason = "SOURCE_OBSERVATION_PROJECTED"
	}

	observation.Digest = digestSelfImprovementReverseClosureLSP(observation)
	return observation
}

func digestSelfImprovementReverseClosureLSP(
	observation SelfImprovementReverseClosureLSPObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}