package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const generationProvenanceAdoptionSchemaPart01 = "gooo/generation-provenance-adoption/v1"

const (
	generationProvenanceAdoptPart01           = "ADOPT"
	generationProvenanceAdoptionUnknownPart01 = "UNKNOWN"
)

type GenerationProvenanceAdoptionPart01 struct {
	Schema                     string `json:"schema"`
	CandidateDigest            string `json:"candidate_digest"`
	GeneratedDigest            string `json:"generated_digest,omitempty"`
	ReverseDigest              string `json:"reverse_digest,omitempty"`
	ReplayEvidencePrefixDigest string `json:"replay_evidence_prefix_digest"`
	Decision                   string `json:"decision"`
	Reason                     string `json:"reason"`
	MissingStageIndex          int    `json:"missing_stage_index"`
	EvidenceDigest             string `json:"evidence_digest"`
}

func EvaluateGenerationProvenanceAdoptionPart01(
	replay GenerationProvenanceReplayPart01,
) (GenerationProvenanceAdoptionPart01, error) {
	adoption := GenerationProvenanceAdoptionPart01{
		Schema:                     generationProvenanceAdoptionSchemaPart01,
		CandidateDigest:            replay.CandidateDigest,
		GeneratedDigest:            replay.GeneratedDigest,
		ReverseDigest:              replay.ReverseDigest,
		ReplayEvidencePrefixDigest: replay.EvidencePrefixDigest,
		Decision:                   generationProvenanceAdoptionUnknownPart01,
		Reason:                     "INVALID_REPLAY_EVIDENCE",
		MissingStageIndex:          replay.MissingStageIndex,
	}
	if replay.ValidPart01() && replay.Decision == generationProvenancePassPart01 {
		adoption.Decision = generationProvenanceAdoptPart01
		adoption.Reason = "GENERATION_AND_REVERSE_OBSERVED"
		adoption.MissingStageIndex = -1
	} else if adoption.MissingStageIndex < 0 {
		adoption.MissingStageIndex = 0
	}
	if err := adoption.assignEvidenceDigestPart01(); err != nil {
		return GenerationProvenanceAdoptionPart01{}, err
	}
	if !adoption.ValidPart01() {
		return GenerationProvenanceAdoptionPart01{}, errors.New("invalid generation provenance adoption")
	}
	return adoption, nil
}

func (adoption GenerationProvenanceAdoptionPart01) ValidPart01() bool {
	if adoption.Schema != generationProvenanceAdoptionSchemaPart01 ||
		!generationProvenanceDigestPart01(adoption.CandidateDigest) ||
		!generationProvenanceDigestPart01(adoption.ReplayEvidencePrefixDigest) ||
		!generationProvenanceDigestPart01(adoption.EvidenceDigest) ||
		strings.TrimSpace(adoption.Reason) == "" {
		return false
	}
	switch adoption.Decision {
	case generationProvenanceAdoptPart01:
		return adoption.MissingStageIndex == -1 &&
			generationProvenanceDigestPart01(adoption.GeneratedDigest) &&
			generationProvenanceDigestPart01(adoption.ReverseDigest)
	case generationProvenanceAdoptionUnknownPart01:
		return adoption.MissingStageIndex >= 0
	default:
		return false
	}
}

func (adoption *GenerationProvenanceAdoptionPart01) assignEvidenceDigestPart01() error {
	payload := struct {
		Schema                     string `json:"schema"`
		CandidateDigest            string `json:"candidate_digest"`
		GeneratedDigest            string `json:"generated_digest,omitempty"`
		ReverseDigest              string `json:"reverse_digest,omitempty"`
		ReplayEvidencePrefixDigest string `json:"replay_evidence_prefix_digest"`
		Decision                   string `json:"decision"`
		Reason                     string `json:"reason"`
		MissingStageIndex          int    `json:"missing_stage_index"`
	}{
		Schema:                     adoption.Schema,
		CandidateDigest:            adoption.CandidateDigest,
		GeneratedDigest:            adoption.GeneratedDigest,
		ReverseDigest:              adoption.ReverseDigest,
		ReplayEvidencePrefixDigest: adoption.ReplayEvidencePrefixDigest,
		Decision:                   adoption.Decision,
		Reason:                     adoption.Reason,
		MissingStageIndex:          adoption.MissingStageIndex,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(encoded)
	adoption.EvidenceDigest = "sha256:" + hex.EncodeToString(digest[:])
	return nil
}
