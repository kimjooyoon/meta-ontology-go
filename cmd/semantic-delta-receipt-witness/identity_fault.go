package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

const (
	identityFaultSchema   = "gooo/semantic-delta-claim-identity-fault/v2"
	identityFaultID       = "raw-only-stable-id-recreation"
	identityFaultTarget   = "alternate-observation"
	identityFaultMutation = "stable_identity_rekey_with_reference_closure"
	identityFaultRule     = "rekey each alternate stable_id bijectively to gooo://semantic-delta/identity-fault/claim/<sha256(rule|alternate-before-raw-digest|alternate-after-raw-digest|original-stable-id)> and rewrite every internal identity reference through the same mapping"
)

type identityFaultArtifact struct {
	Schema   string `json:"schema"`
	FaultID  string `json:"fault_id"`
	Target   string `json:"target"`
	Mutation string `json:"mutation"`
	Rule     string `json:"rule"`
}

type identityFaultEvidence struct {
	ArtifactPath   string `json:"artifact_path"`
	ArtifactBytes  int    `json:"artifact_bytes"`
	ArtifactDigest string `json:"artifact_digest"`
	FaultID        string `json:"fault_id"`
	Target         string `json:"target"`
	Mutation       string `json:"mutation"`
	Rule           string `json:"rule"`
}

type identityFaultMappingRow struct {
	OldStableID string `json:"old_stable_id"`
	NewStableID string `json:"new_stable_id"`
}

type identityFaultGraphEvidence struct {
	OldStableIDs                 []string                  `json:"old_stable_ids"`
	NewStableIDs                 []string                  `json:"new_stable_ids"`
	Mapping                      []identityFaultMappingRow `json:"mapping"`
	MappingDigest                string                    `json:"mapping_digest"`
	MappingTotal                 int                       `json:"mapping_total"`
	Bijection                    bool                      `json:"bijection"`
	ReferenceDenominator         int                       `json:"reference_denominator"`
	RewrittenReferenceCount      int                       `json:"rewritten_reference_count"`
	DanglingReferenceCount       int                       `json:"dangling_reference_count"`
	AlphaEquivalentSemanticGraph bool                      `json:"alpha_equivalent_semantic_graph"`
	Decision                     string                    `json:"decision"`
	Resolution                   string                    `json:"resolution"`
	Stage                        string                    `json:"stage"`
	Step                         string                    `json:"step"`
	Reason                       string                    `json:"reason"`
}

type producerIdentityFaultResult struct {
	Records    []producer.ClaimIdentityRecord
	Graph      identityFaultGraphEvidence
	OldToNew   map[string]string
	NewToOld   map[string]string
	Valid      bool
	FailReason string
}

type consumerIdentityFaultResult struct {
	Records    []consumer.ClaimIdentityRecord
	Graph      identityFaultGraphEvidence
	OldToNew   map[string]string
	NewToOld   map[string]string
	Valid      bool
	FailReason string
}

func readIdentityFaultArtifact(path string) (identityFaultArtifact, identityFaultEvidence, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("read identity fault artifact: %w", err)
	}
	var artifact identityFaultArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("decode identity fault artifact: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("identity fault artifact has trailing data")
	}
	if artifact.Schema != identityFaultSchema || artifact.FaultID != identityFaultID || artifact.Target != identityFaultTarget || artifact.Mutation != identityFaultMutation || artifact.Rule != identityFaultRule {
		return identityFaultArtifact{}, identityFaultEvidence{}, fmt.Errorf("identity fault artifact contract mismatch")
	}
	return artifact, identityFaultEvidence{ArtifactPath: path, ArtifactBytes: len(raw), ArtifactDigest: bytesDigest(raw), FaultID: artifact.FaultID, Target: artifact.Target, Mutation: artifact.Mutation, Rule: artifact.Rule}, nil
}

func mutateProducerIdentityFault(records []producer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact) producerIdentityFaultResult {
	rows := make([]identityFaultMappingRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, identityFaultMappingRow{OldStableID: record.StableID, NewStableID: producerFaultStableID(artifact.Rule, observation, record.StableID)})
	}
	result := producerIdentityFaultResult{OldToNew: map[string]string{}, NewToOld: map[string]string{}}
	result.Graph, result.OldToNew, result.NewToOld, result.FailReason = validateIdentityFaultMapping(rows, identityFaultIDsFromProducer(records), identityFaultNewIDs(rows), nil)
	if result.FailReason != "" {
		result.Graph = identityFaultGraphFailure(result.Graph, result.FailReason)
		return result
	}
	faulted := append([]producer.ClaimIdentityRecord(nil), records...)
	for index := range faulted {
		faulted[index].StableID = result.OldToNew[faulted[index].StableID]
		if faulted[index].PreservationOf != "" {
			if rewritten, ok := result.OldToNew[faulted[index].PreservationOf]; ok {
				faulted[index].PreservationOf = rewritten
			}
		}
	}
	result.Records = faulted
	result.Graph, result.FailReason = validateProducerIdentityFaultGraph(records, result.Records, observation, artifact, result.Graph.Mapping, result.OldToNew, result.NewToOld)
	result.Valid = result.FailReason == ""
	if !result.Valid {
		result.Graph = identityFaultGraphFailure(result.Graph, result.FailReason)
	}
	sort.Slice(result.Records, func(i, j int) bool { return result.Records[i].StableID < result.Records[j].StableID })
	return result
}

func mutateConsumerIdentityFault(records []consumer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact) consumerIdentityFaultResult {
	rows := make([]identityFaultMappingRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, identityFaultMappingRow{OldStableID: record.StableID, NewStableID: consumerFaultStableID(artifact.Rule, observation, record.StableID)})
	}
	result := consumerIdentityFaultResult{OldToNew: map[string]string{}, NewToOld: map[string]string{}}
	result.Graph, result.OldToNew, result.NewToOld, result.FailReason = validateIdentityFaultMapping(rows, identityFaultIDsFromConsumer(records), identityFaultNewIDs(rows), nil)
	if result.FailReason != "" {
		result.Graph = identityFaultGraphFailure(result.Graph, result.FailReason)
		return result
	}
	faulted := append([]consumer.ClaimIdentityRecord(nil), records...)
	for index := range faulted {
		faulted[index].StableID = result.OldToNew[faulted[index].StableID]
		if faulted[index].PreservationOf != "" {
			if rewritten, ok := result.OldToNew[faulted[index].PreservationOf]; ok {
				faulted[index].PreservationOf = rewritten
			}
		}
	}
	result.Records = faulted
	result.Graph, result.FailReason = validateConsumerIdentityFaultGraph(records, result.Records, observation, artifact, result.Graph.Mapping, result.OldToNew, result.NewToOld)
	result.Valid = result.FailReason == ""
	if !result.Valid {
		result.Graph = identityFaultGraphFailure(result.Graph, result.FailReason)
	}
	sort.Slice(result.Records, func(i, j int) bool { return result.Records[i].StableID < result.Records[j].StableID })
	return result
}

func validateIdentityFaultMapping(rows []identityFaultMappingRow, oldIDs, newIDs []string, expected map[string]string) (identityFaultGraphEvidence, map[string]string, map[string]string, string) {
	canonical := append([]identityFaultMappingRow(nil), rows...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].OldStableID != canonical[j].OldStableID {
			return canonical[i].OldStableID < canonical[j].OldStableID
		}
		return canonical[i].NewStableID < canonical[j].NewStableID
	})
	graph := identityFaultGraphEvidence{OldStableIDs: append([]string(nil), oldIDs...), NewStableIDs: append([]string(nil), newIDs...), Mapping: canonical, MappingTotal: len(canonical), MappingDigest: bytesDigest(mustJSON(canonical))}
	sort.Strings(graph.OldStableIDs)
	sort.Strings(graph.NewStableIDs)
	oldToNew := make(map[string]string, len(rows))
	newToOld := make(map[string]string, len(rows))
	for _, row := range rows {
		if row.OldStableID == "" || row.NewStableID == "" {
			return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_NOT_BIJECTIVE"
		}
		if _, exists := oldToNew[row.OldStableID]; exists {
			return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE"
		}
		if _, exists := newToOld[row.NewStableID]; exists {
			return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE"
		}
		oldToNew[row.OldStableID] = row.NewStableID
		newToOld[row.NewStableID] = row.OldStableID
	}
	if len(oldToNew) == 0 || len(oldToNew) != len(oldIDs) || len(newToOld) != len(newIDs) || len(oldToNew) != len(newToOld) || !identityIDsUnique(oldIDs) || !identityIDsUnique(newIDs) {
		return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_NOT_BIJECTIVE"
	}
	if !identityStringSetEqual(oldIDs, mapKeys(oldToNew)) || !identityStringSetEqual(newIDs, mapValues(oldToNew)) {
		return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
	}
	if expected != nil {
		for oldID, newID := range expected {
			if oldToNew[oldID] != newID {
				return graph, oldToNew, newToOld, "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
			}
		}
	}
	graph.Bijection = true
	return graph, oldToNew, newToOld, ""
}

func identityFaultNewIDs(rows []identityFaultMappingRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.NewStableID)
	}
	return ids
}

func validateProducerIdentityFaultGraph(original, faulted []producer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact, rows []identityFaultMappingRow, oldToNew, newToOld map[string]string) (identityFaultGraphEvidence, string) {
	graph, mappedOldToNew, mappedNewToOld, reason := validateIdentityFaultMapping(rows, identityFaultIDsFromProducer(original), identityFaultIDsFromProducer(faulted), oldToNew)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !identityFaultProducerRecordsMatch(original, faulted, observation, artifact, mappedOldToNew, mappedNewToOld) {
		reason := "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.ReferenceDenominator, graph.RewrittenReferenceCount, graph.DanglingReferenceCount = producerReferenceCounts(original, faulted, mappedOldToNew, mappedNewToOld)
	graph.AlphaEquivalentSemanticGraph = producerSemanticGraphDigest(original) == producerSemanticGraphDigest(normalizeProducerFaultedReferences(faulted, mappedNewToOld))
	if graph.DanglingReferenceCount != 0 || graph.RewrittenReferenceCount != graph.ReferenceDenominator {
		reason := "IDENTITY_REFERENCE_CLOSURE_BROKEN"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !graph.AlphaEquivalentSemanticGraph {
		reason := "IDENTITY_SEMANTIC_GRAPH_NOT_ALPHA_EQUIVALENT"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = "PASS", "EXACT", "identity-fault", "rekey-graph", "IDENTITY_FAULT_GRAPH_EXACT"
	return graph, ""
}

func validateConsumerIdentityFaultGraph(original, faulted []consumer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact, rows []identityFaultMappingRow, oldToNew, newToOld map[string]string) (identityFaultGraphEvidence, string) {
	graph, mappedOldToNew, mappedNewToOld, reason := validateIdentityFaultMapping(rows, identityFaultIDsFromConsumer(original), identityFaultIDsFromConsumer(faulted), oldToNew)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !identityFaultConsumerRecordsMatch(original, faulted, observation, artifact, mappedOldToNew, mappedNewToOld) {
		reason := "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.ReferenceDenominator, graph.RewrittenReferenceCount, graph.DanglingReferenceCount = consumerReferenceCounts(original, faulted, mappedOldToNew, mappedNewToOld)
	graph.AlphaEquivalentSemanticGraph = consumerSemanticGraphDigest(original) == consumerSemanticGraphDigest(normalizeConsumerFaultedReferences(faulted, mappedNewToOld))
	if graph.DanglingReferenceCount != 0 || graph.RewrittenReferenceCount != graph.ReferenceDenominator {
		reason := "IDENTITY_REFERENCE_CLOSURE_BROKEN"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !graph.AlphaEquivalentSemanticGraph {
		reason := "IDENTITY_SEMANTIC_GRAPH_NOT_ALPHA_EQUIVALENT"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = "PASS", "EXACT", "identity-fault", "rekey-graph", "IDENTITY_FAULT_GRAPH_EXACT"
	return graph, ""
}

func identityFaultGraphFailure(graph identityFaultGraphEvidence, reason string) identityFaultGraphEvidence {
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "identity-fault", "rekey-graph", reason
	return graph
}

func producerFaultStableID(rule string, observation evolutionSourcePair, original string) string {
	material := strings.Join([]string{rule, observation.BeforeRawDigest, observation.AfterRawDigest, original}, "|")
	digest := strings.TrimPrefix(bytesDigest([]byte(material)), "sha256:")
	return "gooo://semantic-delta/identity-fault/claim/" + digest
}

func consumerFaultStableID(rule string, observation evolutionSourcePair, original string) string {
	material := strings.Join([]string{rule, observation.BeforeRawDigest, observation.AfterRawDigest, original}, "|")
	digest := strings.TrimPrefix(bytesDigest([]byte(material)), "sha256:")
	return "gooo://semantic-delta/identity-fault/claim/" + digest
}

func identityFaultIDsFromProducer(records []producer.ClaimIdentityRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.StableID)
	}
	return ids
}

func identityFaultIDsFromConsumer(records []consumer.ClaimIdentityRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.StableID)
	}
	return ids
}

func identityIDsUnique(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

func identityStringSetEqual(left, right []string) bool {
	if !identityIDsUnique(left) || !identityIDsUnique(right) || len(left) != len(right) {
		return false
	}
	leftSet := make(map[string]bool, len(left))
	for _, id := range left {
		leftSet[id] = true
	}
	for _, id := range right {
		if !leftSet[id] {
			return false
		}
	}
	return true
}

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func mapValues(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func identityFaultProducerRecordsMatch(original, faulted []producer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact, oldToNew, newToOld map[string]string) bool {
	byID := make(map[string]producer.ClaimIdentityRecord, len(faulted))
	for _, record := range faulted {
		if _, exists := byID[record.StableID]; exists {
			return false
		}
		byID[record.StableID] = record
	}
	for _, before := range original {
		newID, ok := oldToNew[before.StableID]
		if !ok || newID != producerFaultStableID(artifact.Rule, observation, before.StableID) {
			return false
		}
		after, ok := byID[newID]
		if !ok || !producerRecordEqualExceptIdentityReferences(before, after) {
			return false
		}
	}
	for _, record := range faulted {
		if _, ok := newToOld[record.StableID]; !ok {
			return false
		}
	}
	return true
}

func identityFaultConsumerRecordsMatch(original, faulted []consumer.ClaimIdentityRecord, observation evolutionSourcePair, artifact identityFaultArtifact, oldToNew, newToOld map[string]string) bool {
	byID := make(map[string]consumer.ClaimIdentityRecord, len(faulted))
	for _, record := range faulted {
		if _, exists := byID[record.StableID]; exists {
			return false
		}
		byID[record.StableID] = record
	}
	for _, before := range original {
		newID, ok := oldToNew[before.StableID]
		if !ok || newID != consumerFaultStableID(artifact.Rule, observation, before.StableID) {
			return false
		}
		after, ok := byID[newID]
		if !ok || !consumerRecordEqualExceptIdentityReferences(before, after) {
			return false
		}
	}
	for _, record := range faulted {
		if _, ok := newToOld[record.StableID]; !ok {
			return false
		}
	}
	return true
}

func producerRecordEqualExceptIdentityReferences(left, right producer.ClaimIdentityRecord) bool {
	return left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func consumerRecordEqualExceptIdentityReferences(left, right consumer.ClaimIdentityRecord) bool {
	return left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func producerReferenceCounts(original, faulted []producer.ClaimIdentityRecord, oldToNew, newToOld map[string]string) (int, int, int) {
	byID := make(map[string]producer.ClaimIdentityRecord, len(faulted))
	for _, record := range faulted {
		byID[record.StableID] = record
	}
	denominator, rewritten, dangling := 0, 0, 0
	for _, before := range original {
		after, ok := byID[oldToNew[before.StableID]]
		if before.PreservationOf == "" {
			if !ok || after.PreservationOf != "" {
				dangling++
			}
			continue
		}
		denominator++
		want := oldToNew[before.PreservationOf]
		if !ok || want == "" || after.PreservationOf != want {
			dangling++
			continue
		}
		if _, ok := newToOld[after.PreservationOf]; !ok {
			dangling++
			continue
		}
		rewritten++
	}
	return denominator, rewritten, dangling
}

func consumerReferenceCounts(original, faulted []consumer.ClaimIdentityRecord, oldToNew, newToOld map[string]string) (int, int, int) {
	byID := make(map[string]consumer.ClaimIdentityRecord, len(faulted))
	for _, record := range faulted {
		byID[record.StableID] = record
	}
	denominator, rewritten, dangling := 0, 0, 0
	for _, before := range original {
		after, ok := byID[oldToNew[before.StableID]]
		if before.PreservationOf == "" {
			if !ok || after.PreservationOf != "" {
				dangling++
			}
			continue
		}
		denominator++
		want := oldToNew[before.PreservationOf]
		if !ok || want == "" || after.PreservationOf != want {
			dangling++
			continue
		}
		if _, ok := newToOld[after.PreservationOf]; !ok {
			dangling++
			continue
		}
		rewritten++
	}
	return denominator, rewritten, dangling
}

func normalizeProducerFaultedReferences(records []producer.ClaimIdentityRecord, newToOld map[string]string) []producer.ClaimIdentityRecord {
	result := append([]producer.ClaimIdentityRecord(nil), records...)
	for index := range result {
		if oldReference, ok := newToOld[result[index].PreservationOf]; ok {
			result[index].PreservationOf = oldReference
		}
	}
	return result
}

func normalizeConsumerFaultedReferences(records []consumer.ClaimIdentityRecord, newToOld map[string]string) []consumer.ClaimIdentityRecord {
	result := append([]consumer.ClaimIdentityRecord(nil), records...)
	for index := range result {
		if oldReference, ok := newToOld[result[index].PreservationOf]; ok {
			result[index].PreservationOf = oldReference
		}
	}
	return result
}

type identityFaultSemanticRow struct {
	Kind                         string `json:"kind"`
	RelationRole                 string `json:"relation_role"`
	NormalizedProposition        string `json:"normalized_proposition"`
	PropositionDigest            string `json:"proposition_digest"`
	TargetAddress                string `json:"target_address"`
	TargetAddressDigest          string `json:"target_address_digest"`
	PreservationOf               string `json:"preservation_of,omitempty"`
	EvidenceBeforeSemanticDigest string `json:"evidence_before_semantic_digest,omitempty"`
	EvidenceAfterSemanticDigest  string `json:"evidence_after_semantic_digest,omitempty"`
}

func producerSemanticGraphDigest(records []producer.ClaimIdentityRecord) string {
	rows := make([]identityFaultSemanticRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, identityFaultSemanticRow{Kind: record.Kind, RelationRole: record.RelationRole, NormalizedProposition: record.NormalizedProposition, PropositionDigest: record.PropositionDigest, TargetAddress: record.TargetAddress, TargetAddressDigest: record.TargetAddressDigest, PreservationOf: record.PreservationOf, EvidenceBeforeSemanticDigest: record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: record.EvidenceAfterSemanticDigest})
	}
	sort.Slice(rows, func(i, j int) bool {
		return identityFaultSemanticRowKey(rows[i]) < identityFaultSemanticRowKey(rows[j])
	})
	return bytesDigest(mustJSON(rows))
}

func consumerSemanticGraphDigest(records []consumer.ClaimIdentityRecord) string {
	rows := make([]identityFaultSemanticRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, identityFaultSemanticRow{Kind: record.Kind, RelationRole: record.RelationRole, NormalizedProposition: record.NormalizedProposition, PropositionDigest: record.PropositionDigest, TargetAddress: record.TargetAddress, TargetAddressDigest: record.TargetAddressDigest, PreservationOf: record.PreservationOf, EvidenceBeforeSemanticDigest: record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: record.EvidenceAfterSemanticDigest})
	}
	sort.Slice(rows, func(i, j int) bool {
		return identityFaultSemanticRowKey(rows[i]) < identityFaultSemanticRowKey(rows[j])
	})
	return bytesDigest(mustJSON(rows))
}

func identityFaultSemanticRowKey(row identityFaultSemanticRow) string {
	return strings.Join([]string{row.Kind, row.RelationRole, row.NormalizedProposition, row.PropositionDigest, row.TargetAddress, row.TargetAddressDigest, row.PreservationOf, row.EvidenceBeforeSemanticDigest, row.EvidenceAfterSemanticDigest}, "\x00")
}
