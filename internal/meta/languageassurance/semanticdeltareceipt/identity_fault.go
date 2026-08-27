package semanticdeltareceipt

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
	identityFaultAlgorithm               = "forward-map-reverse-normalize/v1"
	identityFaultSource                  = "internal/meta/languageassurance/semanticdeltareceipt/identity_fault.go"
	identityFaultSemanticSlotDenominator = 7
)

// IdentityFaultInput names the real source pairs used by the diagnostic fault.
// The producer owns reading both pairs and the fault artifact.
type IdentityFaultInput struct {
	Baseline     Input
	Alternate    Input
	ArtifactPath string
}

// IdentityFaultSourcePair is copied into the wire receipt so the fault proof
// remains tied to the raw and semantic observations that supplied it.
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

// IdentityFaultArtifactEvidence records the separately supplied fault rule;
// it is not a language/compiler input and cannot alter normal projection.
type IdentityFaultArtifactEvidence struct {
	ArtifactPath   string `json:"artifact_path"`
	ArtifactBytes  int    `json:"artifact_bytes"`
	ArtifactDigest string `json:"artifact_digest"`
	FaultID        string `json:"fault_id"`
	Target         string `json:"target"`
	Mutation       string `json:"mutation"`
	Rule           string `json:"rule"`
}

// IdentityFaultMappingRow is a canonical old-to-new StableID edge.
type IdentityFaultMappingRow struct {
	OldStableID string `json:"old_stable_id"`
	NewStableID string `json:"new_stable_id"`
	Ordinal     int    `json:"ordinal"`
}

// IdentityFaultGraphEvidence is calculated entirely by the producer. The
// consumer package has a copied wire model and an independent implementation.
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

// IdentityFaultReceipt is the producer's implementation-specific diagnostic
// receipt. Its algorithm binding is intentionally different from the
// consumer's; the witness compares the bundle's opaque common evidence wire.
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

// IdentityFaultReceiptBundle keeps the implementation-specific receipt and a
// canonical opaque evidence wire. The witness compares only ComparisonBytes;
// it never parses either receipt or recomputes the graph.
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
		Decision:         DecisionFailClosed, Resolution: ResolutionLower, Stage: "identity-fault", Step: "read-artifact", Reason: "IDENTITY_FAULT_ARTIFACT_UNAVAILABLE",
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
	receipt.Persistence = CompareClaimIdentityRecords(baseline.Records, normalizeFaultedReferences(faultedRecords, graphMappingReverse(graph.Mapping)))
	receipt.Graph.RawEvidenceChanged = receipt.Persistence.RawEvidenceChanged
	receipt.Graph.RawEvidenceTotal = receipt.Persistence.RawEvidenceTotal
	if faultReason != "" {
		receipt.Decision, receipt.Resolution, receipt.Stage, receipt.Step, receipt.Reason = DecisionFailClosed, ResolutionLower, "identity-fault", "rekey-graph", faultReason
		return identityFaultReceiptBundle(receipt)
	}
	receipt.Decision, receipt.Resolution, receipt.Stage, receipt.Step, receipt.Reason = DecisionFailClosed, ResolutionLower, receipt.Persistence.Stage, receipt.Persistence.Step, receipt.Persistence.Reason
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
	path, ok := identityFaultSourcePath()
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

func identityFaultSourcePath() (string, bool) {
	_, path, _, ok := runtime.Caller(0)
	return path, ok
}

func identityFaultObservationFromFiles(input Input) (identityFaultObservation, bool) {
	observation, err := ClaimIdentityObservationFromFiles(input)
	if err != nil {
		return identityFaultObservation{}, false
	}
	return identityFaultObservation{
		SourcePair: IdentityFaultSourcePair{BeforePath: observation.BeforePath, AfterPath: observation.AfterPath, BeforeRawDigest: observation.BeforeRawDigest, AfterRawDigest: observation.AfterRawDigest, BeforeSemanticDigest: observation.BeforeSemanticDigest, AfterSemanticDigest: observation.AfterSemanticDigest},
		Records:    observation.Records,
	}, true
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

func rekeyIdentityFault(records []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact) ([]ClaimIdentityRecord, IdentityFaultGraphEvidence, string) {
	rows := make([]IdentityFaultMappingRow, 0, len(records))
	ordinals := identitySemanticOrdinals(records)
	for _, record := range records {
		rows = append(rows, IdentityFaultMappingRow{OldStableID: record.StableID, NewStableID: faultStableID(artifact.Rule, observation, record.StableID), Ordinal: ordinals[record.StableID]})
	}
	graph, oldToNew, newToOld, reason := validateIdentityFaultMapping(rows, identityFaultIDs(records), identityFaultNewIDs(rows), nil)
	if reason != "" {
		return nil, identityFaultGraphFailure(graph, reason), reason
	}
	faulted := append([]ClaimIdentityRecord(nil), records...)
	for index := range faulted {
		oldID := faulted[index].StableID
		faulted[index].StableID = oldToNew[oldID]
		if rewritten, ok := oldToNew[faulted[index].PreservationOf]; ok {
			faulted[index].PreservationOf = rewritten
		}
	}
	sort.Slice(faulted, func(i, j int) bool { return faulted[i].StableID < faulted[j].StableID })
	graph, reason = validateIdentityFaultGraph(records, faulted, observation, artifact, graph.Mapping, oldToNew, newToOld)
	if reason != "" {
		return faulted, graph, reason
	}
	return faulted, graph, ""
}

func validateIdentityFaultMapping(rows []IdentityFaultMappingRow, oldIDs, newIDs []string, expected map[string]string) (IdentityFaultGraphEvidence, map[string]string, map[string]string, string) {
	canonical := append([]IdentityFaultMappingRow(nil), rows...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].OldStableID != canonical[j].OldStableID {
			return canonical[i].OldStableID < canonical[j].OldStableID
		}
		return canonical[i].NewStableID < canonical[j].NewStableID
	})
	graph := IdentityFaultGraphEvidence{OldStableIDs: append([]string(nil), oldIDs...), NewStableIDs: append([]string(nil), newIDs...), Mapping: canonical, MappingTotal: len(canonical), MappingDigest: digestBytes(identityFaultJSON(canonical))}
	sort.Strings(graph.OldStableIDs)
	sort.Strings(graph.NewStableIDs)
	oldToNew := map[string]string{}
	newToOld := map[string]string{}
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
	if len(oldToNew) == 0 || len(oldToNew) != len(oldIDs) || len(newToOld) != len(newIDs) || !identityIDsUnique(oldIDs) || !identityIDsUnique(newIDs) || !identityStringSetEqual(oldIDs, mapKeys(oldToNew)) || !identityStringSetEqual(newIDs, mapValues(oldToNew)) {
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

func validateIdentityFaultGraph(original, faulted []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact, rows []IdentityFaultMappingRow, oldToNew, newToOld map[string]string) (IdentityFaultGraphEvidence, string) {
	graph, mappedOldToNew, mappedNewToOld, reason := validateIdentityFaultMapping(rows, identityFaultIDs(original), identityFaultIDs(faulted), oldToNew)
	graph.SemanticSlotDenominator = identityFaultSemanticSlotDenominator
	graph.SemanticSlotUnique, graph.SemanticSlotTotal = identitySemanticSlotCoverage(original)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	if graph.SemanticSlotUnique != graph.SemanticSlotTotal {
		reason = "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if graph.SemanticSlotUnique != identityFaultSemanticSlotDenominator || graph.SemanticSlotTotal != identityFaultSemanticSlotDenominator {
		reason = "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !identityFaultRecordsMatch(original, faulted, observation, artifact, mappedOldToNew, mappedNewToOld) {
		reason = "IDENTITY_FAULT_MAPPING_RULE_MISMATCH"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.ReferenceDenominator, graph.RewrittenReferenceCount, graph.DanglingReferenceCount = identityReferenceCounts(original, faulted, mappedOldToNew, mappedNewToOld)
	graph.OriginalSemanticGraphDigest = identitySemanticGraphDigest(original)
	graph.NormalizedSemanticGraphDigest = identitySemanticGraphDigest(normalizeFaultedReferences(faulted, mappedNewToOld))
	graph.AlphaEquivalentSemanticGraph = graph.OriginalSemanticGraphDigest == graph.NormalizedSemanticGraphDigest
	if graph.DanglingReferenceCount != 0 || graph.RewrittenReferenceCount != graph.ReferenceDenominator {
		reason = "IDENTITY_REFERENCE_CLOSURE_BROKEN"
		return identityFaultGraphFailure(graph, reason), reason
	}
	if !graph.AlphaEquivalentSemanticGraph {
		reason = "IDENTITY_SEMANTIC_GRAPH_NOT_ALPHA_EQUIVALENT"
		return identityFaultGraphFailure(graph, reason), reason
	}
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = "PASS", ResolutionExact, "identity-fault", "rekey-graph", "IDENTITY_FAULT_GRAPH_EXACT"
	return graph, ""
}

// ValidateIdentityFaultGraph is the producer's public production validation
// entry point for raw observations and mapping rows. It reconstructs the
// mapping maps from the supplied rows before invoking the graph validator.
func ValidateIdentityFaultGraph(original, faulted []ClaimIdentityRecord, observation IdentityFaultSourcePair, rule string, rows []IdentityFaultMappingRow) (IdentityFaultGraphEvidence, string) {
	graph, oldToNew, newToOld, reason := validateIdentityFaultMapping(rows, identityFaultIDs(original), identityFaultIDs(faulted), nil)
	if reason != "" {
		return identityFaultGraphFailure(graph, reason), reason
	}
	return validateIdentityFaultGraph(original, faulted, observation, identityFaultArtifact{Rule: rule}, rows, oldToNew, newToOld)
}

func identityFaultGraphFailure(graph IdentityFaultGraphEvidence, reason string) IdentityFaultGraphEvidence {
	graph.SemanticSlotDenominator = identityFaultSemanticSlotDenominator
	graph.Decision, graph.Resolution, graph.Stage, graph.Step, graph.Reason = DecisionFailClosed, ResolutionLower, "identity-fault", "rekey-graph", reason
	return graph
}

func identitySemanticSlotCoverage(records []ClaimIdentityRecord) (int, int) {
	keys := make([]string, 0, len(records))
	for _, record := range records {
		keys = append(keys, identitySemanticInventoryKey(record))
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

func identitySemanticOrdinals(records []ClaimIdentityRecord) map[string]int {
	ordered := append([]ClaimIdentityRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := identitySemanticInventoryKey(ordered[i]), identitySemanticInventoryKey(ordered[j])
		if left != right {
			return left < right
		}
		return ordered[i].StableID < ordered[j].StableID
	})
	result := make(map[string]int, len(ordered))
	for ordinal, record := range ordered {
		result[record.StableID] = ordinal
	}
	return result
}

func identityFaultIDs(records []ClaimIdentityRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.StableID)
	}
	return ids
}

func identityFaultNewIDs(rows []IdentityFaultMappingRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.NewStableID)
	}
	return ids
}

func identityIDsUnique(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	seen := map[string]bool{}
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
	seen := map[string]bool{}
	for _, id := range left {
		seen[id] = true
	}
	for _, id := range right {
		if !seen[id] {
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

func identityFaultRecordsMatch(original, faulted []ClaimIdentityRecord, observation IdentityFaultSourcePair, artifact identityFaultArtifact, oldToNew, newToOld map[string]string) bool {
	byID := map[string]ClaimIdentityRecord{}
	for _, record := range faulted {
		if _, exists := byID[record.StableID]; exists {
			return false
		}
		byID[record.StableID] = record
	}
	for _, before := range original {
		newID, ok := oldToNew[before.StableID]
		if !ok || newID != faultStableID(artifact.Rule, observation, before.StableID) {
			return false
		}
		after, ok := byID[newID]
		if !ok || !identityRecordEqualExceptReferences(before, after) {
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

func identityRecordEqualExceptReferences(left, right ClaimIdentityRecord) bool {
	return left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func identityReferenceCounts(original, faulted []ClaimIdentityRecord, oldToNew, newToOld map[string]string) (int, int, int) {
	byID := map[string]ClaimIdentityRecord{}
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

func normalizeFaultedReferences(records []ClaimIdentityRecord, newToOld map[string]string) []ClaimIdentityRecord {
	result := append([]ClaimIdentityRecord(nil), records...)
	for index := range result {
		if oldReference, ok := newToOld[result[index].PreservationOf]; ok {
			result[index].PreservationOf = oldReference
		}
	}
	return result
}

func graphMappingReverse(rows []IdentityFaultMappingRow) map[string]string {
	result := map[string]string{}
	for _, row := range rows {
		result[row.NewStableID] = row.OldStableID
	}
	return result
}

// identitySemanticGraphDigest is the producer's forward-map/reverse-normalize
// implementation of the common alpha graph encoding. IDs are represented by
// canonical semantic ordinals, so a closed rekey cannot change the digest.
type identitySemanticRow struct {
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

func identitySemanticGraphDigest(records []ClaimIdentityRecord) string {
	ordered := append([]ClaimIdentityRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := identitySemanticInventoryKey(ordered[i]), identitySemanticInventoryKey(ordered[j])
		if left != right {
			return left < right
		}
		return ordered[i].StableID < ordered[j].StableID
	})
	ordinalByID := make(map[string]int, len(ordered))
	for ordinal, record := range ordered {
		ordinalByID[record.StableID] = ordinal
	}
	rows := make([]identitySemanticRow, 0, len(ordered))
	for ordinal, record := range ordered {
		preservationOrdinal := -1
		if record.PreservationOf != "" {
			if found, ok := ordinalByID[record.PreservationOf]; ok {
				preservationOrdinal = found
			} else {
				preservationOrdinal = -2
			}
		}
		rows = append(rows, identitySemanticRow{Ordinal: ordinal, Kind: record.Kind, RelationRole: record.RelationRole, NormalizedProposition: record.NormalizedProposition, PropositionDigest: record.PropositionDigest, TargetAddress: record.TargetAddress, TargetAddressDigest: record.TargetAddressDigest, PreservationOrdinal: preservationOrdinal, EvidenceBeforeSemanticDigest: record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: record.EvidenceAfterSemanticDigest})
	}
	return digestBytes(identityFaultJSON(rows))
}

func identitySemanticInventoryKey(record ClaimIdentityRecord) string {
	return strings.Join([]string{record.Kind, record.RelationRole, record.NormalizedProposition, record.PropositionDigest, record.TargetAddress, record.TargetAddressDigest}, "\x00")
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
