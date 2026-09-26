package lsp

import (
    "strings"
    "testing"
)

func TestRepairRouteObservationPart01PassAndTamper(t *testing.T) {
    digest := "sha256:" + strings.Repeat("a", 64)
    observation, err := NewRepairRouteObservationPart01(RepairRouteEvidenceInputPart01{
        URI: "file:///workspace/main.gooo",
        SourceDigest: digest,
        ParseDigest: digest,
        Decision: RepairRouteParsePart01,
        DecisionSource: "deterministic",
        Status: RepairRoutePassPart01,
        Confidence: 1,
        MissingStageIndex: -1,
    })
    if err != nil {
        t.Fatalf("construct observation: %v", err)
    }
    if !observation.ValidPart01() {
        t.Fatal("expected valid repair route observation")
    }
    observation.ParseDigest = "sha256:" + strings.Repeat("b", 64)
    if observation.ValidPart01() {
        t.Fatal("tampered parse digest must be rejected")
    }
}

func TestRepairRouteObservationPart01UnknownPreservesFrontier(t *testing.T) {
    digest := "sha256:" + strings.Repeat("c", 64)
    observation, err := NewRepairRouteObservationPart01(RepairRouteEvidenceInputPart01{
        URI: "file:///workspace/main.gooo",
        SourceDigest: digest,
        Decision: RepairRouteUnknownPart01,
        DecisionSource: "jev",
        Status: RepairRouteUnknownStatusPart01,
        Reason: "semantic stage has no evidence",
        Confidence: 0.5,
        MissingStageIndex: 2,
    })
    if err != nil {
        t.Fatalf("construct UNKNOWN observation: %v", err)
    }
    if !observation.ValidPart01() || observation.MissingStageIndex != 2 {
        t.Fatal("expected UNKNOWN observation to preserve frontier")
    }
}