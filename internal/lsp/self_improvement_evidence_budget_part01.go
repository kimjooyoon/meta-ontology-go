package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementEvidenceBudgetLSPObservationSchema = "gooo.lsp.self-improvement.evidence-budget.v1"

type SelfImprovementEvidenceBudgetLSPObservation struct {
	Schema         string                                          `json:"schema"`
	URI            string                                          `json:"uri"`
	Version        int                                             `json:"version"`
	Declaration    string                                          `json:"declaration"`
	Status         valueexecution.SelfImprovementEvidenceBudgetStatus `json:"status"`
	Visible        bool                                            `json:"visible"`
	Reason         string                                          `json:"reason"`
	EvidenceCount  int                                             `json:"evidenceCount"`
	OriginDigest   string                                          `json:"originDigest"`
	NonAuthorizing bool                                            `json:"nonAuthorizing"`
	Digest         string                                          `json:"digest"`
}

func ObserveSelfImprovementEvidenceBudgetLSP(
	uri string,
	version int,
	declaration string,
	source valueexecution.SelfImprovementEvidenceBudgetObservation,
) SelfImprovementEvidenceBudgetLSPObservation {
	observation := SelfImprovementEvidenceBudgetLSPObservation{
		Schema:         SelfImprovementEvidenceBudgetLSPObservationSchema,
		URI:            uri,
		Version:        version,
		Declaration:    declaration,
		Status:         source.Status,
		EvidenceCount:  source.ForwardEvidenceCount + source.ReverseEvidenceCount,
		OriginDigest:   source.Digest,
		NonAuthorizing: true,
	}

	switch {
	case strings.TrimSpace(uri) == "" || version < 0 || strings.TrimSpace(declaration) == "":
		observation.Status = valueexecution.SelfImprovementEvidenceBudgetUnknown
		observation.Reason = "LSP_CONTEXT_INVALID"
	case source.Status == valueexecution.SelfImprovementEvidenceBudgetUnknown || source.Digest == "":
		observation.Status = valueexecution.SelfImprovementEvidenceBudgetUnknown
		observation.Reason = "SOURCE_OBSERVATION_UNKNOWN"
	default:
		observation.Visible = true
		observation.Reason = "SOURCE_OBSERVATION_PROJECTED"
	}

	observation.Digest = digestSelfImprovementEvidenceBudgetLSP(observation)
	return observation
}

func digestSelfImprovementEvidenceBudgetLSP(observation SelfImprovementEvidenceBudgetLSPObservation) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}