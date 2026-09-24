package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

const ExecutionEvidencePrefixObservationSchema = "gooo.lsp.execution-evidence-prefix.v1"

type ExecutionEvidencePrefixStatus string

const (
	ExecutionEvidencePrefixComplete ExecutionEvidencePrefixStatus = "COMPLETE"
	ExecutionEvidencePrefixUnknown  ExecutionEvidencePrefixStatus = "UNKNOWN"
)

type ExecutionEvidencePrefixObservation struct {
	Schema               string                        `json:"schema"`
	StageDigests         []string                      `json:"stage_digests"`
	EvidencePrefixDigest string                        `json:"evidence_prefix_digest"`
	MissingStageIndex    int                           `json:"missing_stage_index"`
	Status               ExecutionEvidencePrefixStatus `json:"status"`
	Reason               string                        `json:"reason"`
	NonAuthorizing       bool                          `json:"non_authorizing"`
	Digest               string                        `json:"digest"`
}

func ObserveExecutionEvidencePrefix(stageDigests []string) ExecutionEvidencePrefixObservation {
	stages := append([]string(nil), stageDigests...)
	missingStageIndex := -1
	for index, stageDigest := range stages {
		if strings.TrimSpace(stageDigest) == "" {
			missingStageIndex = index
			break
		}
	}

	observation := ExecutionEvidencePrefixObservation{
		Schema:            ExecutionEvidencePrefixObservationSchema,
		StageDigests:      stages,
		MissingStageIndex: missingStageIndex,
		NonAuthorizing:    true,
	}
	observation.EvidencePrefixDigest = executionEvidencePrefixDigest(stages, missingStageIndex)
	if missingStageIndex < 0 && len(stages) > 0 {
		observation.Status = ExecutionEvidencePrefixComplete
		observation.Reason = "EXECUTION_EVIDENCE_COMPLETE"
	} else {
		observation.Status = ExecutionEvidencePrefixUnknown
		observation.Reason = "EXECUTION_EVIDENCE_MISSING_STAGE"
	}
	observation.Digest = executionEvidencePrefixObservationDigest(observation)
	return observation
}

func ValidateExecutionEvidencePrefixObservation(observation ExecutionEvidencePrefixObservation) bool {
	expected := ObserveExecutionEvidencePrefix(observation.StageDigests)
	return observation.Schema == expected.Schema &&
		observation.EvidencePrefixDigest == expected.EvidencePrefixDigest &&
		observation.MissingStageIndex == expected.MissingStageIndex &&
		observation.Status == expected.Status &&
		observation.Reason == expected.Reason &&
		observation.NonAuthorizing &&
		observation.Digest == expected.Digest
}

func executionEvidencePrefixDigest(stages []string, missingStageIndex int) string {
	hash := sha256.New()
	hash.Write([]byte(ExecutionEvidencePrefixObservationSchema))
	hash.Write([]byte{0})
	for index, stageDigest := range stages {
		if missingStageIndex >= 0 && index >= missingStageIndex {
			break
		}
		hash.Write([]byte(strconv.Itoa(index)))
		hash.Write([]byte{':'})
		hash.Write([]byte(stageDigest))
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func executionEvidencePrefixObservationDigest(observation ExecutionEvidencePrefixObservation) string {
	hash := sha256.New()
	hash.Write([]byte(observation.Schema))
	hash.Write([]byte{0})
	hash.Write([]byte(observation.EvidencePrefixDigest))
	hash.Write([]byte{0})
	hash.Write([]byte(strconv.Itoa(observation.MissingStageIndex)))
	hash.Write([]byte{0})
	hash.Write([]byte(observation.Status))
	hash.Write([]byte{0})
	hash.Write([]byte(observation.Reason))
	hash.Write([]byte{0})
	if observation.NonAuthorizing {
		hash.Write([]byte("true"))
	} else {
		hash.Write([]byte("false"))
	}
	for index, stageDigest := range observation.StageDigests {
		hash.Write([]byte{0})
		hash.Write([]byte(strconv.Itoa(index)))
		hash.Write([]byte{':'})
		hash.Write([]byte(stageDigest))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
