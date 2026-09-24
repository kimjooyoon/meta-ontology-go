package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementAdoptionBoundaryLSPObservationSchema = "gooo.lsp.self-improvement.adoption-boundary.v1"

type SelfImprovementAdoptionBoundaryLSPObservation struct {
	Schema             string `json:"schema"`
	URI                string `json:"uri"`
	Version            int    `json:"version"`
	Declaration        string `json:"declaration"`
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	Visible            bool   `json:"visible"`
	CandidateDigest    string `json:"candidateDigest"`
	ContractStatus     string `json:"contractStatus"`
	ReverseStatus      string `json:"reverseStatus"`
	ContractPreserved  bool   `json:"contractPreserved"`
	ReverseObserved    bool   `json:"reverseObserved"`
	AdoptionAuthorized bool   `json:"adoptionAuthorized"`
	NonAuthorizing     bool   `json:"nonAuthorizing"`
	OriginReason       string `json:"originReason"`
	OriginDigest       string `json:"originDigest"`
	Digest             string `json:"digest"`
}

func ObserveSelfImprovementAdoptionBoundaryLSP(
	uri string,
	version int,
	declaration string,
	source valueexecution.SelfImprovementAdoptionBoundaryObservation,
) SelfImprovementAdoptionBoundaryLSPObservation {
	observation := SelfImprovementAdoptionBoundaryLSPObservation{
		Schema:             SelfImprovementAdoptionBoundaryLSPObservationSchema,
		URI:                uri,
		Version:            version,
		Declaration:        declaration,
		Status:             source.Status,
		CandidateDigest:    source.CandidateDigest,
		ContractStatus:     source.ContractStatus,
		ReverseStatus:      source.ReverseStatus,
		ContractPreserved:  source.ContractPreserved,
		ReverseObserved:    source.ReverseObserved,
		AdoptionAuthorized: false,
		NonAuthorizing:     true,
		OriginReason:       source.Reason,
		OriginDigest:       source.Digest,
	}

	switch {
	case strings.TrimSpace(uri) == "" || version < 0 || strings.TrimSpace(declaration) == "":
		observation.Status = valueexecution.SelfImprovementAdoptionBoundaryUnknown
		observation.Reason = "LSP_CONTEXT_INVALID"
	case source.Status == "" || source.Status == valueexecution.SelfImprovementAdoptionBoundaryUnknown || source.Digest == "":
		observation.Status = valueexecution.SelfImprovementAdoptionBoundaryUnknown
		observation.Reason = "SOURCE_OBSERVATION_UNKNOWN"
	default:
		observation.Visible = true
		observation.Reason = "SOURCE_OBSERVATION_PROJECTED"
	}

	observation.Digest = digestSelfImprovementAdoptionBoundaryLSP(observation)
	return observation
}

func digestSelfImprovementAdoptionBoundaryLSP(
	observation SelfImprovementAdoptionBoundaryLSPObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}