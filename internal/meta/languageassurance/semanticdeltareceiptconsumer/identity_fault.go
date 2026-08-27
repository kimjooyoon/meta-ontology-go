package semanticdeltareceiptconsumer

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
)

const (
	identityFaultSchema                  = "gooo/semantic-delta-claim-identity-fault/v2"
	identityFaultID                      = "raw-only-stable-id-recreation"
	identityFaultTarget                  = "alternate-observation"
	identityFaultMutation                = "stable_identity_rekey_with_reference_closure"
	identityFaultRule                    = "rekey each alternate stable_id bijectively to gooo://semantic-delta/identity-fault/claim/<sha256(rule|alternate-before-raw-digest|alternate-after-raw-digest|original-stable-id)> and rewrite every internal identity reference through the same mapping"
	identityFaultAlgorithm               = "canonical-ordinal-edge-join/v1"
	identityFaultSource                  = "internal/meta/languageassurance/semanticdeltareceiptconsumer/identity_fault.go"
	identityFaultSemanticSlotDenominator = 7
)

// IdentityFaultInput is the consumer's copied input boundary. It reads the
// raw source pairs and diagnostic artifact without importing the producer.
type IdentityFaultInput struct {
	Baseline     Input
	Alternate    Input
	ArtifactPath string
}

type IdentityFaultSourcePair struct {
	BeforePath           string `json:"before_path"`
	AfterPath            string `json:"after_path"`
	BeforeRawDigest      string `json:"before_raw_digest"`
	AfterRawDigest       string `json:"after_raw_digest"`
	BeforeSemanticDigest string `json:"before_semantic_digest"`
	AfterSemanticDigest  string `json:"after_semantic_digest"`
}

type identityFaultArtifact struct {
	Schema   string `json:"schema"`
	FaultID  string `json:"fault_id"`
	Target   string `json:"target"`
	Mutation string `json:"mutation"`
	Rule     string `json:"rule"`
}

// IdentityFaultArtifactEvidence is a copied wire model, not a producer type.
type IdentityFaultArtifactEvidence struct {
	ArtifactPath   string `json:"artifact_path"`
	ArtifactBytes  int    `json:"artifact_bytes"`
	ArtifactDigest string `json:"artifact_digest"`
	FaultID        string `json:"fault_id"`
	Target         string `json:"target"`
	Mutation       string `json:"mutation"`
	Rule           string `json:"rule"`
}

type IdentityFaultMappingRow struct {
	OldStableID string `json:"old_stable_id"`
	NewStableID string `json:"new_stable_id"`
	Ordinal     int    `json:"ordinal"`
}

type IdentityFaultGraphEvidence struct {
	OldStableIDs                  []string                  `json:"old_stable_ids"`
	NewStableIDs                  []string                  `json:"new_stable_ids"`
	Mapping                       []IdentityFaultMappingRow `json:"mapping"`
	MappingDigest                 string                    `json:"mapping_digest"`
	MappingTotal                  int                       `json:"mapping_total"`
	Bijection                     bool                      `json:"bijection"`
	ReferenceDenominator          int                       `json:"reference_denominator"`
	RewrittenReferenceCount       int                       `json:"rewritten_reference_count"`
	DanglingReferenceCount        int                       `json:"dangling_reference_count"`
	OriginalSemanticGraphDigest   string                    `json:"original_semantic_graph_digest"`
	NormalizedSemanticGraphDigest string                    `json:"normalized_semantic_graph_digest"`
	AlphaEquivalentSemanticGraph  bool                      `json:"alpha_equivalent_semantic_graph"`
	RawEvidenceChanged            int                       `json:"raw_evidence_changed"`
	RawEvidenceTotal              int                       `json:"raw_evidence_total"`
	SemanticSlotDenominator       int                       `json:"semantic_slot_denominator"`
	SemanticSlotUnique            int                       `json:"semantic_slot_unique"`
	SemanticSlotTotal             int                       `json:"semantic_slot_total"`
	Decision                      string                    `json:"decision"`
	Resolution                    string                    `json:"resolution"`
	Stage                         string                    `json:"stage"`
	Step                          string                    `json:"step"`
	Reason                        string                    `json:"reason"`
}

type identityFaultObservation struct {
	SourcePair IdentityFaultSourcePair `json:"source_pair"`
	Records    []ClaimIdentityRecord   `json:"records"`
}

// IdentityFaultReceipt is deliberately shaped like the producer receipt but
// is computed by this package's separate lowering and graph implementation.
type IdentityFaultReceipt struct {
	Schema                string                        `json:"schema"`
	AlgorithmID           string                        `json:"algorithm_id"`
	AlgorithmSourcePath   string                        `json:"algorithm_source_path"`
	AlgorithmSourceBytes  int                           `json:"algorithm_source_bytes"`
	AlgorithmSourceDigest string                        `json:"algorithm_source_digest"`
	Artifact              IdentityFaultArtifactEvidence `json:"artifact"`
	Baseline              identityFaultObservation      `json:"baseline"`
	Alternate             identityFaultObservation      `json:"alternate"`
	FaultedAlternate      identityFaultObservation      `json:"faulted_alternate"`
	Graph                 IdentityFaultGraphEvidence    `json:"graph"`
	Persistence           ClaimIdentityPairComparison   `json:"persistence"`
	RawSemanticPreserved  bool                          `json:"raw_semantic_preserved"`
	ReconstructionExact   bool                          `json:"reconstruction_exact"`
	FaultGraphClosed      bool                          `json:"fault_graph_closed"`
	Decision              string                        `json:"decision"`
	Resolution            string                        `json:"resolution"`
	Stage                 string                        `json:"stage"`
	Step                  string                        `json:"step"`
	Reason                string                        `json:"reason"`
}

// IdentityFaultReceiptBundle exposes a full implementation receipt and an
// opaque common evidence wire. The witness compares the latter without
// interpreting either implementation's algorithm.
type IdentityFaultReceiptBundle struct {
	Receipt         IdentityFaultReceipt
	ComparisonBytes []byte
}

func IdentityFaultReceiptFromFiles(input IdentityFaultInput) IdentityFaultReceiptBundle {
	algorithmID, algorithmPath, algorithmBytes, algorithmDigest := identityFaultAlgorithmBinding()
	receipt := IdentityFaultReceipt{
		Schema: identityFaultSchema, AlgorithmID: algorithmID, AlgorithmSourcePath: algorithmPath, AlgorithmSourceBytes: algorithmBytes, AlgorithmSourceDigest: algorithmDigest,
		Baseline:         identityFaultObservation{SourcePair: IdentityFaultSourcePair{BeforePath: input.Baseline.BeforePath, AfterPath: input.Baseline.AfterPath}},
		Alternate:        identityFaultObservation{SourcePair: IdentityFaultSourcePair{BeforePath: input.Alternate.BeforePath, AfterPath: input.Alternate.AfterPath}},
		FaultedAlternate: identityFaultObservation{SourcePair: IdentityFaultSourcePair{BeforePath: input.Alternate.BeforePath, AfterPath: input.Alternate.AfterPath}},
		Decision:         decisionFailClosed, Resolution: resolutionLower, Stage: "identity-fault", Step: "read-artifact", Reason: "IDENTITY_FAULT_ARTIFACT_UNAVAILABLE",
	}
	if !validIdentityFaultAlgorithmBinding(algorithmPath, algorithmBytes, algorithmDigest) {
		receipt.Stage, receipt.Step, receipt.Reason = "identity-fault", "bind-algorithm-source", "ALGORITHM_SOURCE_UNAVAILABLE"
		return identityFaultReceiptBundle(receipt)
	}
	artifact, evidence, ok := readIdentityFaultArtifact(input.ArtifactPath)
	if !ok {
		receipt.Artifact = evidence
		return identityFaultReceiptBundle(receipt)
	}
	receipt.Artifact = evidence
	baseline, baselineOK := identityFaultObservationFromFiles(input.Baseline)
	if !baselineOK {
		receipt.Stage, receipt.Step, receipt.Reason = "identity-fault", "read-baseline-source-pair", "SOURCE_PAIR_UNAVAILABLE"
		return identityFaultReceiptBundle(receipt)
	}
	alternate, alternateOK := identityFaultObservationFromFiles(input.Alternate)
	if !alternateOK {
		receipt.Stage, receipt.Step, receipt.Reason = "identity-fault", "read-alternate-source-pair", "SOURCE_PAIR_UNAVAILABLE"
		return identityFaultReceiptBundle(receipt)
	}
	faultedRecords, graph, faultReason := rekeyIdentityFault(alternate.Records, alternate.SourcePair, artifact)
	receipt.Baseline, receipt.Alternate = baseline, alternate
	receipt.FaultedAlternate = identityFaultObservation{SourcePair: alternate.SourcePair, Records: faultedRecords}
	receipt.Graph = graph
	receipt.RawSemanticPreserved = sourcePairSemanticPreserved(baseline.SourcePair, alternate.SourcePair)
	receipt.ReconstructionExact = true
	receipt.FaultGraphClosed = faultReason == ""
	receipt.Persistence = CompareClaimIdentityRecords(baseline.Records, consumerOrdinalJoinForPersistence(alternate.Records, faultedRecords))
	receipt.Graph.RawEvidenceChanged = receipt.Persistence.RawEvidenceChanged
	receipt.Graph.RawEvidenceTotal = receipt.Persistence.RawEvidenceTotal
	if faultReason != "" {
		receipt.Decision, receipt.Resolution, receipt.Stage, receipt.Step, receipt.Reason = decisionFailClosed, resolutionLower, "identity-fault", "rekey-graph", faultReason
		return identityFaultReceiptBundle(receipt)
	}
	receipt.Decision, receipt.Resolution, receipt.Stage, receipt.Step, receipt.Reason = decisionFailClosed, resolutionLower, receipt.Persistence.Stage, receipt.Persistence.Step, receipt.Persistence.Reason
	return identityFaultReceiptBundle(receipt)
}

func identityFaultReceiptBundle(receipt IdentityFaultReceipt) IdentityFaultReceiptBundle {
	return IdentityFaultReceiptBundle{Receipt: receipt, ComparisonBytes: MarshalIdentityFaultReceiptEvidence(receipt)}
}

// MarshalIdentityFaultReceiptEvidence serializes the receipt's common evidence
// wire. Implementation-specific algorithm provenance is intentionally omitted
// so an opaque witness can compare producer and consumer evidence without
// importing either implementation's logic.
func MarshalIdentityFaultReceiptEvidence(receipt IdentityFaultReceipt) []byte {
	comparison := receipt
	comparison.AlgorithmID = ""
	comparison.AlgorithmSourcePath = ""
	comparison.AlgorithmSourceBytes = 0
	comparison.AlgorithmSourceDigest = ""
	return identityFaultJSON(comparison)
}

func identityFaultAlgorithmBinding() (string, string, int, string) {
	path, _, _, ok := runtime.Caller(0)
	if !ok {
		return identityFaultAlgorithm, identityFaultSource, 0, ""
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return identityFaultAlgorithm, identityFaultSource, 0, ""
	}
	return identityFaultAlgorithm, identityFaultSource, len(raw), digestBytes(raw)
}

func validIdentityFaultAlgorithmBinding(path string, bytes int, digest string) bool {
	return path == identityFaultSource && bytes > 0 && strings.HasPrefix(digest, "sha256:") && len(digest) == len("sha256:")+64
}

func identityFaultObservationFromFiles(input Input) (identityFaultObservation, bool) {
	records, sourcePair, err := ClaimIdentityRecordsFromFiles(input)
	if err != nil {
		return identityFaultObservation{}, false
	}
	return identityFaultObservation{SourcePair: IdentityFaultSourcePair{BeforePath: sourcePair.BeforePath, AfterPath: sourcePair.AfterPath, BeforeRawDigest: sourcePair.BeforeRawDigest, AfterRawDigest: sourcePair.AfterRawDigest, BeforeSemanticDigest: sourcePair.BeforeSemanticDigest, AfterSemanticDigest: sourcePair.AfterSemanticDigest}, Records: records}, true
}

func readIdentityFaultArtifact(path string) (identityFaultArtifact, IdentityFaultArtifactEvidence, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return identityFaultArtifact{}, IdentityFaultArtifactEvidence{ArtifactPath: path}, false
	}
	var artifact identityFaultArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&artifact) != nil {
		return identityFaultArtifact{}, IdentityFaultArtifactEvidence{ArtifactPath: path, ArtifactBytes: len(raw), ArtifactDigest: digestBytes(raw)}, false
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF || artifact.Schema != identityFaultSchema || artifact.FaultID != identityFaultID || artifact.Target != identityFaultTarget || artifact.Mutation != identityFaultMutation || artifact.Rule != identityFaultRule {
		return identityFaultArtifact{}, IdentityFaultArtifactEvidence{ArtifactPath: path, ArtifactBytes: len(raw), ArtifactDigest: digestBytes(raw)}, false
	}
	return artifact, IdentityFaultArtifactEvidence{ArtifactPath: path, ArtifactBytes: len(raw), ArtifactDigest: digestBytes(raw), FaultID: artifact.FaultID, Target: artifact.Target, Mutation: artifact.Mutation, Rule: artifact.Rule}, true
}

type consumerOrdinalRecord struct {
	record     ClaimIdentityRecord
	sourceSlot int
	ordinal    int
}

type consumerOrdinalEdge struct {
	ordinal int
	oldID   string
	newID   string
}

// rekeyIdentityFault constructs the graph by semantic ordinal. It never
// builds a producer-style old->new map: references are rewritten by joining
// the target's old ordinal to the target's new ordinal.
func rekeyIdentityFault(records []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact) ([]ClaimIdentityRecord, IdentityFaultGraphEvidence, string) {
	ordered := consumerOrdinalInventory(records)
	if reason := consumerOrdinalInventoryReason(ordered); reason != "" {
		return nil, identityFaultGraphFailure(consumerGraphForRecords(records, nil), reason), reason
	}
	faulted := make([]ClaimIdentityRecord, len(ordered))
	rows := make([]IdentityFaultMappingRow, 0, len(ordered))
	for ordinal, entry := range ordered {
		value := entry.record
		value.StableID = faultStableID(artifact.Rule, observation, entry.record.StableID)
		if entry.record.PreservationOf != "" {
			targetOrdinal := consumerFindOrdinal(ordered, entry.record.PreservationOf)
			if targetOrdinal < 0 {
				return nil, identityFaultGraphFailure(consumerGraphForRecords(records, nil), "IDENTITY_REFERENCE_CLOSURE_BROKEN"), "IDENTITY_REFERENCE_CLOSURE_BROKEN"
			}
			value.PreservationOf = faultStableID(artifact.Rule, observation, ordered[targetOrdinal].record.StableID)
		}
		faulted[ordinal] = value
		rows = append(rows, IdentityFaultMappingRow{OldStableID: entry.record.StableID, NewStableID: value.StableID, Ordinal: ordinal})
	}
	sort.Slice(faulted, func(i, j int) bool { return faulted[i].StableID < faulted[j].StableID })
	graph, reason := validateIdentityFaultGraph(records, faulted, observation, artifact, rows)
	if reason != "" {
		return faulted, graph, reason
	}
	return faulted, graph, ""
}

func validateIdentityFaultGraph(original, faulted []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact, rows []IdentityFaultMappingRow) (IdentityFaultGraphEvidence, string) {
	oldIDs := consumerSortedIDs(original)
	newIDs := consumerSortedIDs(faulted)
	graph := consumerGraphForRecords(original, newIDs)
	if graph.SemanticSlotUnique != graph.SemanticSlotTotal {
		reason := "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if graph.SemanticSlotUnique != identityFaultSemanticSlotDenominator || graph.SemanticSlotTotal != identityFaultSemanticSlotDenominator {
		reason := "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	expectedEdges, reason := consumerExpectedOrdinalEdges(original, observation, artifact)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	observedEdges, reason := consumerObservedOrdinalEdges(original, faulted, rows)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	if reason = validateConsumerOrdinalEdges(observedEdges, expectedEdges); reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	expectedRows := consumerRowsFromEdges(expectedEdges)
	graph, reason = consumerValidateMappingRows(rows, expectedRows, oldIDs, newIDs)
	graph.SemanticSlotDenominator = identityFaultSemanticSlotDenominator
	graph.SemanticSlotUnique, graph.SemanticSlotTotal = consumerSemanticSlotCoverage(original)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !consumerRecordsMatchByOrdinal(original, faulted, observedEdges) {
		reason = "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.ReferenceDenominator, graph.RewrittenReferenceCount, graph.DanglingReferenceCount = consumerReferenceClosure(original, faulted)
	graph.OriginalSemanticGraphDigest = consumerAlphaGraphDigest(original)
	graph.NormalizedSemanticGraphDigest = consumerAlphaGraphDigest(faulted)
	graph.AlphaEquivalentSemanticGraph = graph.OriginalSemanticGraphDigest == graph.NormalizedSemanticGraphDigest
	if graph.DanglingReferenceCount != 0 || graph.RewrittenReferenceCount != graph.ReferenceDenominator {
		reason = "IDENTITY_REFERENCE_CLOSURE_BROKEN"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !graph.AlphaEquivalentSemanticGraph {
		reason = "IDENTITY_SEMANTIC_GRAPH_NOT_ALPHA_EQUIVALENT"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = "PASS", resolutionExact, "identity-fault", "rekey-graph", "IDENTITY_FAULT_GRAPH_EXACT"
	return graph, ""
}

// ValidateIdentityFaultGraph is the consumer's production validation entry
// point for raw observations and wire mapping rows. It is intentionally
// public so fixtures can exercise the same observed-edge reconstruction used
// by receipt production.
func ValidateIdentityFaultGraph(original, faulted []ClaimIdentityRecord, observation IdentityFaultSourcePair, rule string, rows []IdentityFaultMappingRow) (IdentityFaultGraphEvidence, string) {
	return validateIdentityFaultGraph(original, faulted, observation, identityFaultArtifact{Rule: rule}, rows)
}

func identityFaultGraphFailure(graph IdentityFaultGraphEvidence, reason string) IdentityFaultGraphEvidence {
	graph.SemanticSlotDenominator = identityFaultSemanticSlotDenominator
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = decisionFailClosed, resolutionLower, "identity-fault", "rekey-graph", reason
	return graph
}

func consumerGraphForRecords(original []ClaimIdentityRecord, newIDs []string) IdentityFaultGraphEvidence {
	oldIDs := consumerSortedIDs(original)
	if newIDs == nil {
		newIDs = []string{}
	}
	rows := make([]IdentityFaultMappingRow, 0, len(oldIDs))
	unique, total := consumerSemanticSlotCoverage(original)
	return IdentityFaultGraphEvidence{OldStableIDs: oldIDs, NewStableIDs: append([]string(nil), newIDs...), Mapping: rows, MappingTotal: 0, MappingDigest: digestBytes(identityFaultJSON(rows)), SemanticSlotDenominator: identityFaultSemanticSlotDenominator, SemanticSlotUnique: unique, SemanticSlotTotal: total}
}

func consumerSemanticSlotCoverage(records []ClaimIdentityRecord) (int, int) {
	keys := make([]string, 0, len(records))
	for _, record := range records {
		keys = append(keys, consumerSemanticSlotKey(record))
	}
	sort.Strings(keys)
	unique := 0
	for index, key := range keys {
		if index == 0 || key != keys[index-1] {
			unique++
		}
	}
	return unique, len(keys)
}

func consumerOrdinalInventory(records []ClaimIdentityRecord) []consumerOrdinalRecord {
	ordered := make([]consumerOrdinalRecord, len(records))
	for slot, record := range records {
		ordered[slot] = consumerOrdinalRecord{record: record, sourceSlot: slot}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := consumerSemanticSlotKey(ordered[i].record), consumerSemanticSlotKey(ordered[j].record)
		if left != right {
			return left < right
		}
		return ordered[i].record.StableID < ordered[j].record.StableID
	})
	for ordinal := range ordered {
		ordered[ordinal].ordinal = ordinal
	}
	return ordered
}

func consumerSemanticSlotKey(record ClaimIdentityRecord) string {
	return strings.Join([]string{record.Kind, record.RelationRole, record.NormalizedProposition, record.PropositionDigest, record.TargetAddress, record.TargetAddressDigest}, "\x00")
}

func consumerOrdinalInventoryReason(ordered []consumerOrdinalRecord) string {
	if len(ordered) == 0 {
		return "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
	}
	ids := make([]string, len(ordered))
	for index, entry := range ordered {
		if entry.record.StableID == "" || entry.ordinal != index {
			return "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
		}
		ids[index] = entry.record.StableID
	}
	sort.Strings(ids)
	for index := 1; index < len(ids); index++ {
		if ids[index] == ids[index-1] {
			return "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE"
		}
	}
	for index := 1; index < len(ordered); index++ {
		if consumerSemanticSlotKey(ordered[index-1].record) == consumerSemanticSlotKey(ordered[index].record) {
			return "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS"
		}
	}
	return ""
}

func consumerSortedIDs(records []ClaimIdentityRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.StableID)
	}
	sort.Strings(ids)
	return ids
}

func consumerFindOrdinal(ordered []consumerOrdinalRecord, stableID string) int {
	for _, entry := range ordered {
		if entry.record.StableID == stableID {
			return entry.ordinal
		}
	}
	return -1
}

func consumerExpectedOrdinalEdges(original []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact) ([]consumerOrdinalEdge, string) {
	oldInventory := consumerOrdinalInventory(original)
	if reason := consumerOrdinalInventoryReason(oldInventory); reason != "" {
		return nil, reason
	}
	edges := make([]consumerOrdinalEdge, len(oldInventory))
	for ordinal := range oldInventory {
		want := faultStableID(artifact.Rule, observation, oldInventory[ordinal].record.StableID)
		edges[ordinal] = consumerOrdinalEdge{ordinal: ordinal, oldID: oldInventory[ordinal].record.StableID, newID: want}
	}
	return edges, ""
}

func consumerObservedOrdinalEdges(original, faulted []ClaimIdentityRecord, rows []IdentityFaultMappingRow) ([]consumerOrdinalEdge, string) {
	oldInventory, newInventory := consumerOrdinalInventory(original), consumerOrdinalInventory(faulted)
	if reason := consumerOrdinalInventoryReason(oldInventory); reason != "" {
		return nil, reason
	}
	if reason := consumerOrdinalInventoryReason(newInventory); reason != "" {
		return nil, reason
	}
	if len(oldInventory) != len(newInventory) {
		return nil, "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
	}
	canonical := append([]IdentityFaultMappingRow(nil), rows...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].OldStableID != canonical[j].OldStableID {
			return canonical[i].OldStableID < canonical[j].OldStableID
		}
		return canonical[i].NewStableID < canonical[j].NewStableID
	})
	if len(canonical) != len(oldInventory) || !consumerUniqueMappingRows(canonical) {
		return nil, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE"
	}
	edges := make([]consumerOrdinalEdge, len(oldInventory))
	for ordinal, entry := range oldInventory {
		rowIndex := -1
		for index, row := range canonical {
			if row.OldStableID == entry.record.StableID {
				rowIndex = index
				break
			}
		}
		if rowIndex < 0 || canonical[rowIndex].NewStableID == "" {
			return nil, "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
		}
		newID := canonical[rowIndex].NewStableID
		if consumerFindOrdinal(newInventory, newID) < 0 {
			return nil, "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
		}
		edges[ordinal] = consumerOrdinalEdge{ordinal: canonical[rowIndex].Ordinal, oldID: entry.record.StableID, newID: newID}
	}
	return edges, ""
}

func validateConsumerOrdinalEdges(observed, expected []consumerOrdinalEdge) string {
	if len(observed) != len(expected) {
		return "IDENTITY_FAULT_ORDINAL_EDGE_MISMATCH"
	}
	ordered := append([]consumerOrdinalEdge(nil), observed...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ordinal != ordered[j].ordinal {
			return ordered[i].ordinal < ordered[j].ordinal
		}
		if ordered[i].oldID != ordered[j].oldID {
			return ordered[i].oldID < ordered[j].oldID
		}
		return ordered[i].newID < ordered[j].newID
	})
	for ordinal, edge := range ordered {
		if edge.ordinal != ordinal || edge.oldID != expected[ordinal].oldID {
			return "IDENTITY_FAULT_ORDINAL_EDGE_MISMATCH"
		}
	}
	return ""
}

func consumerRowsFromEdges(edges []consumerOrdinalEdge) []IdentityFaultMappingRow {
	rows := make([]IdentityFaultMappingRow, 0, len(edges))
	for _, edge := range edges {
		rows = append(rows, IdentityFaultMappingRow{OldStableID: edge.oldID, NewStableID: edge.newID, Ordinal: edge.ordinal})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].OldStableID < rows[j].OldStableID })
	return rows
}

func consumerValidateMappingRows(rows, expected []IdentityFaultMappingRow, oldIDs, newIDs []string) (IdentityFaultGraphEvidence, string) {
	canonical := append([]IdentityFaultMappingRow(nil), rows...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].OldStableID != canonical[j].OldStableID {
			return canonical[i].OldStableID < canonical[j].OldStableID
		}
		return canonical[i].NewStableID < canonical[j].NewStableID
	})
	graph := IdentityFaultGraphEvidence{OldStableIDs: append([]string(nil), oldIDs...), NewStableIDs: append([]string(nil), newIDs...), Mapping: canonical, MappingTotal: len(canonical), MappingDigest: digestBytes(identityFaultJSON(canonical))}
	if len(canonical) == 0 || !consumerUniqueMappingRows(canonical) {
		return graph, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE"
	}
	observedOld, observedNew := make([]string, len(canonical)), make([]string, len(canonical))
	for index, row := range canonical {
		if row.OldStableID == "" || row.NewStableID == "" {
			return graph, "IDENTITY_FAULT_MAPPING_NOT_BIJECTIVE"
		}
		observedOld[index], observedNew[index] = row.OldStableID, row.NewStableID
	}
	sort.Strings(observedOld)
	sort.Strings(observedNew)
	expectedCanonical := append([]IdentityFaultMappingRow(nil), expected...)
	sort.Slice(expectedCanonical, func(i, j int) bool { return expectedCanonical[i].OldStableID < expectedCanonical[j].OldStableID })
	expectedOld, expectedNew := make([]string, len(expectedCanonical)), make([]string, len(expectedCanonical))
	for index, row := range expectedCanonical {
		expectedOld[index], expectedNew[index] = row.OldStableID, row.NewStableID
	}
	sort.Strings(expectedOld)
	sort.Strings(expectedNew)
	if !consumerStringSliceEqual(observedOld, oldIDs) || !consumerStringSliceEqual(observedNew, newIDs) || !consumerStringSliceEqual(expectedOld, oldIDs) || !consumerStringSliceEqual(expectedNew, newIDs) {
		return graph, "IDENTITY_FAULT_MAPPING_INVENTORY_MISMATCH"
	}
	if len(canonical) != len(expectedCanonical) {
		return graph, "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
	}
	for index := range canonical {
		if canonical[index] != expectedCanonical[index] {
			return graph, "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
		}
	}
	graph.Bijection = true
	return graph, ""
}

func consumerUniqueMappingRows(rows []IdentityFaultMappingRow) bool {
	byOld := append([]IdentityFaultMappingRow(nil), rows...)
	byNew := append([]IdentityFaultMappingRow(nil), rows...)
	sort.Slice(byOld, func(i, j int) bool { return byOld[i].OldStableID < byOld[j].OldStableID })
	sort.Slice(byNew, func(i, j int) bool { return byNew[i].NewStableID < byNew[j].NewStableID })
	for index := 1; index < len(byOld); index++ {
		if byOld[index].OldStableID == byOld[index-1].OldStableID || byNew[index].NewStableID == byNew[index-1].NewStableID {
			return false
		}
	}
	return byOld[0].OldStableID != "" && byNew[0].NewStableID != ""
}

func consumerStringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func consumerRecordsMatchByOrdinal(original, faulted []ClaimIdentityRecord, edges []consumerOrdinalEdge) bool {
	oldInventory, newInventory := consumerOrdinalInventory(original), consumerOrdinalInventory(faulted)
	if len(oldInventory) != len(newInventory) || len(edges) != len(oldInventory) {
		return false
	}
	for ordinal, edge := range edges {
		if oldInventory[ordinal].record.StableID != edge.oldID || newInventory[ordinal].record.StableID != edge.newID || !identityRecordEqualExceptReferences(oldInventory[ordinal].record, newInventory[ordinal].record) {
			return false
		}
	}
	return true
}

func identityRecordEqualExceptReferences(left, right ClaimIdentityRecord) bool {
	return left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func consumerReferenceClosure(original, faulted []ClaimIdentityRecord) (int, int, int) {
	oldInventory, newInventory := consumerOrdinalInventory(original), consumerOrdinalInventory(faulted)
	if len(oldInventory) != len(newInventory) {
		return 0, 0, len(original) + len(faulted)
	}
	denominator, rewritten, dangling := 0, 0, 0
	for ordinal, oldEntry := range oldInventory {
		oldRecord, newRecord := oldEntry.record, newInventory[ordinal].record
		if oldRecord.PreservationOf == "" {
			if newRecord.PreservationOf != "" {
				dangling++
			}
			continue
		}
		denominator++
		targetOrdinal := consumerFindOrdinal(oldInventory, oldRecord.PreservationOf)
		if targetOrdinal < 0 || newRecord.PreservationOf != newInventory[targetOrdinal].record.StableID {
			dangling++
			continue
		}
		rewritten++
	}
	return denominator, rewritten, dangling
}

// consumerOrdinalJoinForPersistence restores only the comparison slot's
// relation reference by ordinal. It is not a reverse map and is never used to
// validate the fault graph itself.
func consumerOrdinalJoinForPersistence(original, faulted []ClaimIdentityRecord) []ClaimIdentityRecord {
	joined := append([]ClaimIdentityRecord(nil), faulted...)
	oldInventory, newInventory := consumerOrdinalInventory(original), consumerOrdinalInventory(faulted)
	if len(oldInventory) != len(newInventory) {
		return joined
	}
	for ordinal, oldEntry := range oldInventory {
		slot := newInventory[ordinal].sourceSlot
		joined[slot].PreservationOf = oldEntry.record.PreservationOf
	}
	return joined
}

type consumerAlphaVertex struct {
	Ordinal                      int    `json:"ordinal"`
	Kind                         string `json:"kind"`
	RelationRole                 string `json:"relation_role"`
	NormalizedProposition        string `json:"normalized_proposition"`
	PropositionDigest            string `json:"proposition_digest"`
	TargetAddress                string `json:"target_address"`
	TargetAddressDigest          string `json:"target_address_digest"`
	PreservationOrdinal          int    `json:"preservation_ordinal"`
	EvidenceBeforeSemanticDigest string `json:"evidence_before_semantic_digest,omitempty"`
	EvidenceAfterSemanticDigest  string `json:"evidence_after_semantic_digest,omitempty"`
}

func consumerAlphaGraphDigest(records []ClaimIdentityRecord) string {
	ordered := consumerOrdinalInventory(records)
	vertices := make([]consumerAlphaVertex, 0, len(ordered))
	for ordinal, entry := range ordered {
		preservationOrdinal := -1
		if entry.record.PreservationOf != "" {
			preservationOrdinal = consumerFindOrdinal(ordered, entry.record.PreservationOf)
			if preservationOrdinal < 0 {
				preservationOrdinal = -2
			}
		}
		vertices = append(vertices, consumerAlphaVertex{Ordinal: ordinal, Kind: entry.record.Kind, RelationRole: entry.record.RelationRole, NormalizedProposition: entry.record.NormalizedProposition, PropositionDigest: entry.record.PropositionDigest, TargetAddress: entry.record.TargetAddress, TargetAddressDigest: entry.record.TargetAddressDigest, PreservationOrdinal: preservationOrdinal, EvidenceBeforeSemanticDigest: entry.record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: entry.record.EvidenceAfterSemanticDigest})
	}
	return digestBytes(identityFaultJSON(vertices))
}

func faultStableID(rule string, observation IdentityFaultSourcePair, original string) string {
	material := strings.Join([]string{rule, observation.BeforeRawDigest, observation.AfterRawDigest, original}, "|")
	return "gooo://semantic-delta/identity-fault/claim/" + strings.TrimPrefix(digestBytes([]byte(material)), "sha256:")
}

func sourcePairSemanticPreserved(left, right IdentityFaultSourcePair) bool {
	return left.BeforeSemanticDigest != "" && left.AfterSemanticDigest != "" && left.BeforeSemanticDigest == right.BeforeSemanticDigest && left.AfterSemanticDigest == right.AfterSemanticDigest
}

func identityFaultJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
