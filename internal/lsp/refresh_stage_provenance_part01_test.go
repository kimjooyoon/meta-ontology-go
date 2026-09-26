package lsp

import (
    "strings"
    "testing"
)

func TestRefreshStageProvenancePart01PreservesPassAndUnknown(t *testing.T) {
    digest := "sha256:" + strings.Repeat("0", 64)
    pass, err := BuildRefreshStageProvenancePart01(RefreshObservationPart01{
        Schema:            refreshObservationSchemaPart01,
        URI:               "file:///workspace/example.gooo",
        Decision:          refreshObservationPassPart01,
        Reason:            "PARSE_REFRESH",
        SourceDigest:      digest,
        SemanticDigest:    digest,
        ProfileDigest:     digest,
        ToolchainDigest:   digest,
        ContractDigest:    digest,
        ParseCalls:        1,
        CacheHits:         0,
        StaleResultCount:  0,
        MissingStageIndex: -1,
    })
    if err != nil {
        t.Fatalf("pass build error = %v", err)
    }
    if !pass.ValidPart01() || pass.MissingStageName != "complete" {
        t.Fatalf("invalid pass provenance: %#v", pass)
    }

    unknown, err := BuildRefreshStageProvenancePart01(RefreshObservationPart01{
        Schema:            refreshObservationSchemaPart01,
        URI:               "file:///workspace/example.gooo",
        Decision:          refreshObservationUnknownPart01,
        Reason:            "MISSING_SEMANTIC_DIGEST",
        SourceDigest:      digest,
        ProfileDigest:     digest,
        ToolchainDigest:   digest,
        ContractDigest:    digest,
        ParseCalls:        1,
        CacheHits:         0,
        StaleResultCount:  0,
        MissingStageIndex: 1,
    })
    if err != nil {
        t.Fatalf("unknown build error = %v", err)
    }
    if !unknown.ValidPart01() || unknown.MissingStageName != "parse" {
        t.Fatalf("invalid unknown provenance: %#v", unknown)
    }

    tampered := unknown
    tampered.EvidencePrefixDigest = digest
    if tampered.ValidPart01() {
        t.Fatal("tampered prefix digest unexpectedly validated")
    }

    tampered = pass
    tampered.ParseCalls++
    if tampered.ValidPart01() {
        t.Fatal("tampered parse count unexpectedly validated")
    }
}
