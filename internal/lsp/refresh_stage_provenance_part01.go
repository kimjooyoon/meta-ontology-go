package lsp

import (
    "crypto/sha256"
    "encoding/encoding"
    "encoding/hex"
    "encoding/json"
    "errors"
    "strings"
)

const refreshStageProvenanceSchemaPart01 = "gooo/lsp-refresh-stage-provenance/v1"

var refreshStageNamesPart01 = []string{
    "source",
    "parse",
    "semantic",
    "profile",
    "toolchain",
    "contract",
}

type RefreshStageProvenancePart01 struct {
    Schema               string `json:"schema"`
    URI                  string `json:"uri"`
    Decision             string `json:"decision"`
    MissingStageIndex    int    `json:"missing_stage_index"`
    MissingStageName     string `json:"missing_stage_name"`
    EvidencePrefixDigest string `json:"evidence_prefix_digest"`
    SourceDigest         string `json:"source_digest"`
    SemanticDigest       string `json:"semantic_digest,omitempty"`
    ProfileDigest        string `json:"profile_digest"`
    ToolchainDigest      string `json:"toolchain_digest"`
    ContractDigest       string `json:"contract_digest"`
}

func BuildRefreshStageProvenancePart01(
    observation RefreshObservationPart01,
) (RefreshStageProvenancePart01, error) {
    if !observation.ValidPart01() {
        return RefreshStageProvenancePart01{}, errors.New("lsp: refresh observation is invalid")
    }
    stageName := refreshStageNamePart01(observation.MissingStageIndex)
    if observation.Decision == refreshObservationPassPart01 {
        stageName = "complete"
    }
    prefixDigest, err := refreshEvidencePrefixDigestPart01(observation)
    if err != nil {
        return RefreshStageProvenancePart01{}, err
    }
    return RefreshStageProvenancePart01{
        Schema:               refreshStageProvenanceSchemaPart01,
        URI:                  observation.URI,
        Decision:             observation.Decision,
        MissingStageIndex:    observation.MissingStageIndex,
        MissingStageName:     stageName,
        EvidencePrefixDigest: prefixDigest,
        SourceDigest:         observation.SourceDigest,
        SemanticDigest:       observation.SemanticDigest,
        ProfileDigest:        observation.ProfileDigest,
        ToolchainDigest:      observation.ToolchainDigest,
        ContractDigest:       observation.ContractDigest,
    }, nil
}

func (value RefreshStageProvenancePart01) ValidPart01() bool {
    if value.Schema != refreshStageProvenanceSchemaPart01 ||
        strings.TrimSpace(value.URI) == "" ||
        !refreshObservationDigestPart01(value.SourceDigest) ||
        !refreshObservationDigestPart01(value.ProfileDigest) ||
        !refreshObservationDigestPart01(value.ToolchainDigest) ||
        !refreshObservationDigestPart01(value.ContractDigest) ||
        !refreshObservationDigestPart01(value.EvidencePrefixDigest) {
        return false
    }
    if value.Decision == refreshObservationPassPart01 {
        return value.MissingStageIndex == -1 &&
            value.MissingStageName == "complete" &&
            strings.TrimSpace(value.SemanticDigest) != ""
    }
    return value.Decision == refreshObservationUnknownPart01 &&
        value.MissingStageIndex >= 0 &&
        strings.TrimSpace(value.MissingStageName) != ""
}

func refreshStageNamePart01(index int) string {
    if index >= 0 && index < len(refreshStageNamesPart01) {
        return refreshStageNamesPart01[index]
    }
    return "unknown-stage"
}

func refreshEvidencePrefixDigestPart01(observation RefreshObservationPart01) (string, error) {
    prefix := struct {
        Schema            string `json:"schema"`
        URI               string `json:"uri"`
        Decision          string `json:"decision"`
        Reason            string `json:"reason"`
        SourceDigest      string `json:"source_digest"`
        SemanticDigest    string `json:"semantic_digest,omitempty"`
        ProfileDigest     string `json:"profile_digest"`
        ToolchainDigest   string `json:"toolchain_digest"`
        ContractDigest    string `json:"contract_digest"`
        ParseCalls        int    `json:"parse_calls"`
        CacheHits         int    `json:"cache_hits"`
        StaleResultCount  int    `json:"stale_result_count"`
        MissingStageIndex int    `json:"missing_stage_index"`
    }{
        Schema:            observation.Schema,
        URI:               observation.URI,
        Decision:          observation.Decision,
        Reason:            observation.Reason,
        SourceDigest:      observation.SourceDigest,
        SemanticDigest:    observation.SemanticDigest,
        ProfileDigest:     observation.ProfileDigest,
        ToolchainDigest:   observation.ToolchainDigest,
        ContractDigest:    observation.ContractDigest,
        ParseCalls:        observation.ParseCalls,
        CacheHits:         observation.CacheHits,
        StaleResultCount:  observation.StaleResultCount,
        MissingStageIndex: observation.MissingStageIndex,
    }
    encoded, err := json.Marshal(prefix)
    if err != nil {
        return "", err
    }
    digest := sha256.Sum256(encoded)
    return "sha256:" + hex.EncodeToString(digest[:]), nil
}
