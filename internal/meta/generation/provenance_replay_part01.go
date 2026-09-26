package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const generationProvenanceReplaySchemaPart01 = "gooo/generation-provenance-replay/v1"

const (
	generationProvenancePassPart01    = "PASS"
	generationProvenanceUnknownPart01 = "UNKNOWN"
)

type GenerationProvenanceReplayPart01 struct {
	Schema               string `json:"schema"`
	Task                 string `json:"task"`
	CandidateDigest      string `json:"candidate_digest"`
	GeneratedDigest      string `json:"generated_digest,omitempty"`
	ReverseDigest        string `json:"reverse_digest,omitempty"`
	Decision             string `json:"decision"`
	Reason               string `json:"reason"`
	MissingStageIndex    int    `json:"missing_stage_index"`
	EvidencePrefixDigest string `json:"evidence_prefix_digest"`
}

func BuildGenerationProvenanceReplayPart01(
	task,
	candidateDigest,
	generatedDigest,
	reverseDigest,
	decision,
	reason string,
	missingStageIndex int,
) (GenerationProvenanceReplayPart01, error) {
	replay := GenerationProvenanceReplayPart01{
		Schema:            generationProvenanceReplaySchemaPart01,
		Task:              task,
		CandidateDigest:   candidateDigest,
		GeneratedDigest:   generatedDigest,
		ReverseDigest:     reverseDigest,
		Decision:          decision,
		Reason:            reason,
		MissingStageIndex: missingStageIndex,
	}
	if err := replay.validateShape(); err != nil {
		return GenerationProvenanceReplayPart01{}, err
	}
	digest, err := generationProvenancePrefixDigestPart01(replay)
	if err != nil {
		return GenerationProvenanceReplayPart01{}, err
	}
	replay.EvidencePrefixDigest = digest
	return replay, nil
}

func (replay GenerationProvenanceReplayPart01) ValidPart01() bool {
	if replay.Schema != generationProvenanceReplaySchemaPart01 ||
		strings.TrimSpace(replay.Task) == "" ||
		!generationProvenanceDigestPart01(replay.CandidateDigest) ||
		!generationProvenanceDigestPart01(replay.EvidencePrefixDigest) {
		return false
	}
	if replay.Decision == generationProvenancePassPart01 {
		if replay.MissingStageIndex != -1 ||
			!generationProvenanceDigestPart01(replay.GeneratedDigest) ||
			!generationProvenanceDigestPart01(replay.ReverseDigest) {
			return false
		}
	} else if replay.Decision != generationProvenanceUnknownPart01 ||
		replay.MissingStageIndex < 0 ||
		strings.TrimSpace(replay.Reason) == "" {
		return false
	}
	expected, err := generationProvenancePrefixDigestPart01(replay)
	return err == nil && expected == replay.EvidencePrefixDigest
}

func (replay GenerationProvenanceReplayPart01) validateShape() error {
	if replay.Schema != generationProvenanceReplaySchemaPart01 ||
		strings.TrimSpace(replay.Task) == "" ||
		!generationProvenanceDigestPart01(replay.CandidateDigest) {
		return errors.New("generation provenance replay is incomplete")
	}
	if replay.Decision == generationProvenancePassPart01 {
		if replay.MissingStageIndex != -1 ||
			!generationProvenanceDigestPart01(replay.GeneratedDigest) ||
			!generationProvenanceDigestPart01(replay.ReverseDigest) {
			return errors.New("generation provenance pass is incomplete")
		}
		return nil
	}
	if replay.Decision == generationProvenanceUnknownPart01 &&
		replay.MissingStageIndex >= 0 &&
		strings.TrimSpace(replay.Reason) != "" {
		return nil
	}
	return errors.New("invalid generation provenance replay decision")
}

func generationProvenanceDigestPart01(value string) bool {
	return strings.HasPrefix(value, "sha256:") &&
		len(strings.TrimPrefix(value, "sha256:")) == 64
}

func generationProvenancePrefixDigestPart01(
	replay GenerationProvenanceReplayPart01,
) (string, error) {
	prefix := struct {
		Schema            string `json:"schema"`
		Task              string `json:"task"`
		CandidateDigest   string `json:"candidate_digest"`
		GeneratedDigest   string `json:"generated_digest,omitempty"`
		ReverseDigest     string `json:"reverse_digest,omitempty"`
		Decision          string `json:"decision"`
		Reason            string `json:"reason"`
		MissingStageIndex int    `json:"missing_stage_index"`
	}{
		Schema:            replay.Schema,
		Task:              replay.Task,
		CandidateDigest:   replay.CandidateDigest,
		GeneratedDigest:   replay.GeneratedDigest,
		ReverseDigest:     replay.ReverseDigest,
		Decision:          replay.Decision,
		Reason:            replay.Reason,
		MissingStageIndex: replay.MissingStageIndex,
	}
	encoded, err := json.Marshal(prefix)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
