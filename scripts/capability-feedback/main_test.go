package main

import (
    "strings"
    "testing"
)

func TestAggregatePromotesOnlyRepeatedResolvedUsefulness(t *testing.T) {
    records := []feedback{
        {
            SourceDigest:            "source-1",
            DeclarationDigest:       "declaration-1",
            CapabilityReceiptDigest: "capability-1",
            EvidenceDigest:          "evidence-1",
            SuggestedQuery:          "Which constraints can be checked next?",
            Outcome:                 "useful",
            NonExecuting:            true,
            NonAuthorizing:          true,
        },
        {
            SourceDigest:            "source-1",
            DeclarationDigest:       "declaration-1",
            CapabilityReceiptDigest: "capability-1",
            EvidenceDigest:          "evidence-2",
            SuggestedQuery:          "Which constraints can be checked next?",
            Outcome:                 "useful",
            NonExecuting:            true,
            NonAuthorizing:          true,
        },
    }

    result, err := aggregate(records)
    if err != nil {
        t.Fatalf("aggregate: %v", err)
    }
    if len(result) != 1 || result[0].NextAction != "REVIEW_INVESTMENT" {
        t.Fatalf("unexpected investment result: %+v", result)
    }
    if len(result[0].EvidenceDigests) != 2 {
        t.Fatalf("expected two evidence digests: %+v", result[0].EvidenceDigests)
    }
}

func TestAggregateKeepsUnresolvedObservationsAsCollection(t *testing.T) {
    record := feedback{
        SourceDigest:            "source-2",
        DeclarationDigest:       "declaration-2",
        CapabilityReceiptDigest: "capability-2",
        EvidenceDigest:          "evidence-3",
        SuggestedQuery:          "What evidence is still missing?",
        Outcome:                 "unresolved",
        NonExecuting:            true,
        NonAuthorizing:          true,
    }

    result, err := aggregate([]feedback{record})
    if err != nil {
        t.Fatalf("aggregate: %v", err)
    }
    if len(result) != 1 || result[0].NextAction != "COLLECT_MORE_OBSERVATIONS" {
        t.Fatalf("unexpected unresolved result: %+v", result)
    }
}

func TestDecodeFeedbackRejectsTrailingJSON(t *testing.T) {
    input := `[{"source_digest":"source","declaration_digest":"declaration","capability_receipt_digest":"capability","evidence_digest":"evidence","suggested_query":"What can gooo inspect?","outcome":"useful","non_executing":true,"non_authorizing":true}] {}`
    if _, err := decodeFeedback(strings.NewReader(input)); err == nil {
        t.Fatal("expected trailing JSON to be rejected")
    }
}

func TestFeedbackValidationRejectsAuthorizingRecord(t *testing.T) {
    record := feedback{
        SourceDigest:            "source",
        DeclarationDigest:       "declaration",
        CapabilityReceiptDigest: "capability",
        EvidenceDigest:          "evidence",
        SuggestedQuery:          "What can gooo inspect?",
        Outcome:                 "useful",
        NonExecuting:            true,
        NonAuthorizing:          false,
    }
    if err := record.validate(); err == nil {
        t.Fatal("expected authorizing record to be rejected")
    }
}