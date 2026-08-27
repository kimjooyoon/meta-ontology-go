package main

import (
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

func TestProducerIdentityFaultGraphRejectsReferenceTampering(t *testing.T) {
	observation := evolutionSourcePair{BeforeRawDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", AfterRawDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	original := producerIdentityFaultTestRecords()
	fault := mutateProducerIdentityFault(original, observation, identityFaultArtifact{Rule: identityFaultRule})
	if !fault.Valid || fault.Graph.ReferenceDenominator != 1 || fault.Graph.RewrittenReferenceCount != 1 || fault.Graph.DanglingReferenceCount != 0 || !fault.Graph.AlphaEquivalentSemanticGraph {
		t.Fatalf("valid graph fault was not closed: %+v", fault)
	}

	tampered := append([]producer.ClaimIdentityRecord(nil), fault.Records...)
	for index := range tampered {
		if tampered[index].PreservationOf != "" {
			tampered[index].PreservationOf = "object"
		}
	}
	graph, reason := validateProducerIdentityFaultGraph(original, tampered, observation, identityFaultArtifact{Rule: identityFaultRule}, fault.Graph.Mapping, fault.OldToNew, fault.NewToOld)
	if reason != "IDENTITY_REFERENCE_CLOSURE_BROKEN" || graph.Decision != "FAIL_CLOSED" || graph.Resolution != "LOWER_RESOLUTION" || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != reason || graph.DanglingReferenceCount == 0 {
		t.Fatalf("stale reference was not rejected: reason=%s graph=%+v", reason, graph)
	}
}

func TestConsumerIdentityFaultGraphRejectsReferenceTampering(t *testing.T) {
	observation := evolutionSourcePair{BeforeRawDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", AfterRawDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	original := consumerIdentityFaultTestRecords()
	fault := mutateConsumerIdentityFault(original, observation, identityFaultArtifact{Rule: identityFaultRule})
	if !fault.Valid || fault.Graph.ReferenceDenominator != 1 || fault.Graph.RewrittenReferenceCount != 1 || fault.Graph.DanglingReferenceCount != 0 || !fault.Graph.AlphaEquivalentSemanticGraph {
		t.Fatalf("valid graph fault was not closed: %+v", fault)
	}

	tampered := append([]consumer.ClaimIdentityRecord(nil), fault.Records...)
	for index := range tampered {
		if tampered[index].PreservationOf != "" {
			tampered[index].PreservationOf = "object"
		}
	}
	graph, reason := validateConsumerIdentityFaultGraph(original, tampered, observation, identityFaultArtifact{Rule: identityFaultRule}, fault.Graph.Mapping, fault.OldToNew, fault.NewToOld)
	if reason != "IDENTITY_REFERENCE_CLOSURE_BROKEN" || graph.Decision != "FAIL_CLOSED" || graph.Resolution != "LOWER_RESOLUTION" || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != reason || graph.DanglingReferenceCount == 0 {
		t.Fatalf("stale reference was not rejected: reason=%s graph=%+v", reason, graph)
	}
}

func TestIdentityFaultMappingRejectsSwapAndDuplicate(t *testing.T) {
	observation := evolutionSourcePair{BeforeRawDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", AfterRawDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	fault := mutateProducerIdentityFault(producerIdentityFaultTestRecords(), observation, identityFaultArtifact{Rule: identityFaultRule})
	if !fault.Valid {
		t.Fatalf("valid graph fault was not created: %+v", fault)
	}

	swapped := append([]identityFaultMappingRow(nil), fault.Graph.Mapping...)
	swapped[0].NewStableID, swapped[1].NewStableID = swapped[1].NewStableID, swapped[0].NewStableID
	_, _, _, reason := validateIdentityFaultMapping(swapped, fault.Graph.OldStableIDs, fault.Graph.NewStableIDs, fault.OldToNew)
	if reason != "IDENTITY_FAULT_MAPPING_RULE_MISMATCH" {
		t.Fatalf("swapped mapping was not rejected: %s", reason)
	}

	duplicate := append([]identityFaultMappingRow(nil), fault.Graph.Mapping...)
	duplicate[1].OldStableID = duplicate[0].OldStableID
	_, _, _, reason = validateIdentityFaultMapping(duplicate, fault.Graph.OldStableIDs, fault.Graph.NewStableIDs, fault.OldToNew)
	if reason != "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE" {
		t.Fatalf("duplicate mapping was not rejected: %s", reason)
	}
}

func producerIdentityFaultTestRecords() []producer.ClaimIdentityRecord {
	return []producer.ClaimIdentityRecord{
		{StableID: "object", Kind: "OBJECT", RelationRole: "uses|observation|after", NormalizedProposition: "OBJECT\x00order\x00uses\x00payment", PropositionDigest: "sha256:1111111111111111111111111111111111111111111111111111111111111111", TargetAddress: "order\x00uses\x00payment", TargetAddressDigest: "sha256:2222222222222222222222222222222222222222222222222222222222222222", EvidenceBeforeRawDigest: "sha256:3333333333333333333333333333333333333333333333333333333333333333", EvidenceAfterRawDigest: "sha256:4444444444444444444444444444444444444444444444444444444444444444", EvidenceBeforeSemanticDigest: "sha256:5555555555555555555555555555555555555555555555555555555555555555", EvidenceAfterSemanticDigest: "sha256:6666666666666666666666666666666666666666666666666666666666666666"},
		{StableID: "preservation", Kind: "BEFORE_CLAIM_PRESERVATION", RelationRole: "preserves", NormalizedProposition: "BEFORE_CLAIM_PRESERVATION\x00claim\x00preserves\x00object", PropositionDigest: "sha256:7777777777777777777777777777777777777777777777777777777777777777", TargetAddress: "order\x00uses\x00payment", TargetAddressDigest: "sha256:2222222222222222222222222222222222222222222222222222222222222222", PreservationOf: "object", EvidenceBeforeRawDigest: "sha256:3333333333333333333333333333333333333333333333333333333333333333", EvidenceAfterRawDigest: "sha256:4444444444444444444444444444444444444444444444444444444444444444", EvidenceBeforeSemanticDigest: "sha256:5555555555555555555555555555555555555555555555555555555555555555", EvidenceAfterSemanticDigest: "sha256:6666666666666666666666666666666666666666666666666666666666666666"},
	}
}

func consumerIdentityFaultTestRecords() []consumer.ClaimIdentityRecord {
	return []consumer.ClaimIdentityRecord{
		{StableID: "object", Kind: "OBJECT", RelationRole: "uses|observation|after", NormalizedProposition: "OBJECT\x00order\x00uses\x00payment", PropositionDigest: "sha256:1111111111111111111111111111111111111111111111111111111111111111", TargetAddress: "order\x00uses\x00payment", TargetAddressDigest: "sha256:2222222222222222222222222222222222222222222222222222222222222222", EvidenceBeforeRawDigest: "sha256:3333333333333333333333333333333333333333333333333333333333333333", EvidenceAfterRawDigest: "sha256:4444444444444444444444444444444444444444444444444444444444444444", EvidenceBeforeSemanticDigest: "sha256:5555555555555555555555555555555555555555555555555555555555555555", EvidenceAfterSemanticDigest: "sha256:6666666666666666666666666666666666666666666666666666666666666666"},
		{StableID: "preservation", Kind: "BEFORE_CLAIM_PRESERVATION", RelationRole: "preserves", NormalizedProposition: "BEFORE_CLAIM_PRESERVATION\x00claim\x00preserves\x00object", PropositionDigest: "sha256:7777777777777777777777777777777777777777777777777777777777777777", TargetAddress: "order\x00uses\x00payment", TargetAddressDigest: "sha256:2222222222222222222222222222222222222222222222222222222222222222", PreservationOf: "object", EvidenceBeforeRawDigest: "sha256:3333333333333333333333333333333333333333333333333333333333333333", EvidenceAfterRawDigest: "sha256:4444444444444444444444444444444444444444444444444444444444444444", EvidenceBeforeSemanticDigest: "sha256:5555555555555555555555555555555555555555555555555555555555555555", EvidenceAfterSemanticDigest: "sha256:6666666666666666666666666666666666666666666666666666666666666666"},
	}
}
