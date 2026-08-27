package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

const evolutionSchema = "gooo/semantic-delta-claim-expectation-evolution/v2"

type evolutionExpectation struct {
	Schema                      string                    `json:"schema"`
	DenominatorID               string                    `json:"denominator_id"`
	ClaimCountContractVersion   string                    `json:"claim_count_contract_version"`
	DenominatorEvolutionReceipt string                    `json:"denominator_evolution_receipt"`
	FixedClaimTotal             int                       `json:"fixed_claim_total"`
	FixedCaseTotal              int                       `json:"fixed_case_total"`
	Cases                       []evolutionExpectationRow `json:"cases"`
}

type evolutionExpectationRow struct {
	ID                               string                         `json:"id"`
	ExpectedClaimIDs                 []string                       `json:"expected_claim_ids"`
	ExpectedClaims                   []consumer.ClaimIdentityRecord `json:"expected_claims"`
	ExpectedTransitionIdentityDigest string                         `json:"expected_transition_identity_digest"`
	ExpectedClaimTotal               int                            `json:"expected_claim_total"`
	CaseRowDigest                    string                         `json:"case_row_digest"`
}

type legacyExpectation struct {
	Schema                      string                 `json:"schema"`
	DenominatorID               string                 `json:"denominator_id"`
	ClaimCountContractVersion   string                 `json:"claim_count_contract_version"`
	DenominatorEvolutionReceipt string                 `json:"denominator_evolution_receipt"`
	FixedClaimTotal             int                    `json:"fixed_claim_total"`
	FixedCaseTotal              int                    `json:"fixed_case_total"`
	Cases                       []legacyExpectationRow `json:"cases"`
}

type legacyExpectationRow struct {
	ID                               string   `json:"id"`
	ExpectedClaimIDs                 []string `json:"expected_claim_ids"`
	ExpectedTransitionIdentityDigest string   `json:"expected_transition_identity_digest"`
	ExpectedClaimTotal               int      `json:"expected_claim_total"`
	CaseRowDigest                    string   `json:"case_row_digest"`
}

type evolutionCase struct {
	CaseID                           string                    `json:"case_id"`
	OldExpectedIDs                   []string                  `json:"old_expected_ids"`
	OldObservedIDs                   []string                  `json:"old_observed_ids"`
	ProducerOldObservedIDs           []string                  `json:"producer_old_observed_ids"`
	ConsumerOldObservedIDs           []string                  `json:"consumer_old_observed_ids"`
	OldArtifactExact                 bool                      `json:"old_artifact_exact"`
	OldProducerConsumerExact         bool                      `json:"old_producer_consumer_exact"`
	NewExpectedIDs                   []string                  `json:"new_expected_ids"`
	ProducerObservedIDs              []string                  `json:"producer_observed_ids"`
	ConsumerObservedIDs              []string                  `json:"consumer_observed_ids"`
	NewExpectationProducerExact      bool                      `json:"new_expectation_producer_exact"`
	NewExpectationConsumerExact      bool                      `json:"new_expectation_consumer_exact"`
	RemovedIDs                       []string                  `json:"removed_ids"`
	AddedIDs                         []string                  `json:"added_ids"`
	PropositionTargetChanges         []propositionTargetChange `json:"proposition_target_changes"`
	StableIdentityPreserved          int                       `json:"stable_identity_preserved"`
	StableIdentityTotal              int                       `json:"stable_identity_total"`
	EvidenceOnlyChanges              int                       `json:"evidence_only_changes"`
	EvidenceOnlyTotal                int                       `json:"evidence_only_total"`
	ClaimRecreatedDueOnlyToRaw       int                       `json:"claim_recreated_due_only_to_raw_digest"`
	ClaimRecreatedDueOnlyToRawTotal  int                       `json:"claim_recreated_due_only_to_raw_digest_total"`
	ObservedSourcePair               evolutionSourcePair       `json:"observed_source_pair"`
	ProducerConsumerExact            bool                      `json:"producer_consumer_exact"`
	ConsumerPropositionTargetChanges []propositionTargetChange `json:"consumer_proposition_target_changes"`
	Decision                         string                    `json:"decision"`
	Resolution                       string                    `json:"resolution"`
	Stage                            string                    `json:"stage"`
	Step                             string                    `json:"step"`
	Reason                           string                    `json:"reason"`
}

type propositionTargetChange struct {
	OldID                string `json:"old_id"`
	NewID                string `json:"new_id"`
	OldPropositionDigest string `json:"old_proposition_digest,omitempty"`
	NewPropositionDigest string `json:"new_proposition_digest,omitempty"`
	OldTargetAddress     string `json:"old_target_address,omitempty"`
	NewTargetAddress     string `json:"new_target_address,omitempty"`
	Reason               string `json:"reason"`
}

type evolutionSourcePair struct {
	BeforePath           string `json:"before_path"`
	AfterPath            string `json:"after_path"`
	BeforeRawDigest      string `json:"before_raw_digest"`
	AfterRawDigest       string `json:"after_raw_digest"`
	BeforeSemanticDigest string `json:"before_semantic_digest"`
	AfterSemanticDigest  string `json:"after_semantic_digest"`
}

type evolutionReport struct {
	Schema                          string          `json:"schema"`
	Authority                       string          `json:"authority"`
	OldArtifactPath                 string          `json:"old_artifact_path"`
	OldArtifactBytes                int             `json:"old_artifact_bytes"`
	OldArtifactDigest               string          `json:"old_artifact_digest"`
	NewArtifactPath                 string          `json:"new_artifact_path"`
	NewArtifactBytes                int             `json:"new_artifact_bytes"`
	NewArtifactDigest               string          `json:"new_artifact_digest"`
	DenominatorID                   string          `json:"denominator_id"`
	DenominatorUnchanged            bool            `json:"denominator_unchanged"`
	FixedClaimTotalBefore           int             `json:"fixed_claim_total_before"`
	FixedClaimTotalAfter            int             `json:"fixed_claim_total_after"`
	StableIdentityPreserved         int             `json:"stable_identity_preserved"`
	StableIdentityTotal             int             `json:"stable_identity_total"`
	PersistentClaimIdentity         int             `json:"persistent_claim_identity"`
	PersistentClaimIdentityTotal    int             `json:"persistent_claim_identity_total"`
	PropositionChanges              int             `json:"proposition_changes"`
	EvidenceOnlyChanges             int             `json:"evidence_only_changes"`
	EvolutionRowsReconstructed      int             `json:"evolution_rows_independently_reconstructed"`
	EvolutionRowsTotal              int             `json:"evolution_rows_total"`
	EvolutionClaimRowsReconstructed int             `json:"evolution_claim_rows_independently_reconstructed"`
	EvolutionClaimRowsTotal         int             `json:"evolution_claim_rows_total"`
	RawEvidenceChangedNonsemantic   int             `json:"raw_evidence_changed_on_nonsemantic"`
	RawEvidenceNonsemanticTotal     int             `json:"raw_evidence_nonsemantic_total"`
	SemanticTargetPreserved         int             `json:"semantic_target_preserved_on_nonsemantic"`
	SemanticTargetNonsemanticTotal  int             `json:"semantic_target_nonsemantic_total"`
	ClaimRecreatedDueOnlyToRaw      int             `json:"claim_recreated_due_only_to_raw_digest"`
	ClaimRecreatedDueOnlyToRawTotal int             `json:"claim_recreated_due_only_to_raw_digest_total"`
	ChangeKind                      string          `json:"change_kind"`
	Cases                           []evolutionCase `json:"cases"`
	Decision                        string          `json:"decision"`
	Resolution                      string          `json:"resolution"`
	Stage                           string          `json:"stage"`
	Step                            string          `json:"step"`
	Reason                          string          `json:"reason"`
}

func reconstructEvolution(oldPath, newPath string) (evolutionReport, error) {
	oldRaw, err := os.ReadFile(oldPath)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("read old expectation artifact: %w", err)
	}
	newRaw, err := os.ReadFile(newPath)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("read new expectation artifact: %w", err)
	}
	newArtifact, err := decodeEvolutionExpectation(newRaw)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("decode new expectation artifact: %w", err)
	}
	oldArtifact, err := decodeLegacyExpectation(oldRaw)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("decode old expectation artifact: %w", err)
	}
	report := evolutionReport{Schema: evolutionSchema, Authority: "SOURCE_DERIVED_SEMANTIC_CLAIM_CONTRACT", OldArtifactPath: oldPath, OldArtifactBytes: len(oldRaw), OldArtifactDigest: bytesDigest(oldRaw), NewArtifactPath: newPath, NewArtifactBytes: len(newRaw), NewArtifactDigest: bytesDigest(newRaw), DenominatorID: newArtifact.DenominatorID, DenominatorUnchanged: oldArtifact.DenominatorID == newArtifact.DenominatorID && oldArtifact.FixedClaimTotal == newArtifact.FixedClaimTotal, FixedClaimTotalBefore: oldArtifact.FixedClaimTotal, FixedClaimTotalAfter: newArtifact.FixedClaimTotal, EvolutionRowsTotal: len(newArtifact.Cases), Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "claim-identity-evolution", Step: "reconstruct-old-new-artifacts", Reason: "CLAIM_IDENTITY_EVOLUTION_UNKNOWN"}
	oldByID := make(map[string]legacyExpectationRow, len(oldArtifact.Cases))
	for _, row := range oldArtifact.Cases {
		oldByID[row.ID] = row
	}
	newByID := make(map[string]evolutionExpectationRow, len(newArtifact.Cases))
	for _, row := range newArtifact.Cases {
		newByID[row.ID] = row
	}
	for _, definition := range producer.Denominator() {
		oldRow, oldOK := oldByID[definition.ID]
		newRow, newOK := newByID[definition.ID]
		input := producer.Input{CaseID: definition.ID, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath, SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ObservedCheckoutSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
		producerReceipt, producerErr := producer.ProduceFiles(input)
		producerRecords := producer.ClaimIdentityRecords(producerReceipt)
		consumerRecords, sourcePair, consumerErr := consumer.ClaimIdentityRecordsFromFiles(consumer.Input{CaseID: input.CaseID, BeforePath: input.BeforePath, AfterPath: input.AfterPath, SubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA})
		oldObservedIDs, oldRecords := legacyClaimIdentity(producerReceipt, input)
		consumerOldRecords, _, consumerOldErr := consumer.LegacyClaimIdentityRecordsFromFiles(consumer.Input{CaseID: input.CaseID, BeforePath: input.BeforePath, AfterPath: input.AfterPath, SubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA})
		row := evolutionCase{CaseID: definition.ID, Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "claim-identity-evolution", Step: "reconstruct-old-new-artifacts", Reason: "CLAIM_IDENTITY_EVOLUTION_UNKNOWN"}
		if oldOK {
			row.OldExpectedIDs = sortedIDs(oldRow.ExpectedClaimIDs)
		}
		row.OldObservedIDs = oldObservedIDs
		row.ProducerOldObservedIDs = oldObservedIDs
		for _, record := range consumerOldRecords {
			row.ConsumerOldObservedIDs = append(row.ConsumerOldObservedIDs, record.StableID)
		}
		row.ConsumerOldObservedIDs = sortedIDs(row.ConsumerOldObservedIDs)
		row.OldProducerConsumerExact = consumerOldErr == nil && recordIDsEqual(row.ProducerOldObservedIDs, row.ConsumerOldObservedIDs)
		row.OldArtifactExact = oldOK && recordIDsEqual(row.OldExpectedIDs, row.ProducerOldObservedIDs) && recordIDsEqual(row.OldExpectedIDs, row.ConsumerOldObservedIDs) && row.OldProducerConsumerExact
		if newOK {
			row.NewExpectedIDs = sortedIDs(newRow.ExpectedClaimIDs)
		}
		row.ProducerObservedIDs = producerRecordIDs(producerRecords)
		row.ConsumerObservedIDs = consumerRecordIDs(consumerRecords)
		row.RemovedIDs, row.AddedIDs = setDiff(row.OldExpectedIDs, row.NewExpectedIDs)
		row.ProducerConsumerExact = producerRecordsEqual(producerRecords, consumerRecords)
		row.NewExpectationProducerExact = expectedProducerRecordsEqual(newRow.ExpectedClaims, producerRecords)
		row.NewExpectationConsumerExact = expectedConsumerRecordsEqual(newRow.ExpectedClaims, consumerRecords)
		row.ObservedSourcePair = evolutionSourcePair{BeforePath: sourcePair.BeforePath, AfterPath: sourcePair.AfterPath, BeforeRawDigest: sourcePair.BeforeRawDigest, AfterRawDigest: sourcePair.AfterRawDigest, BeforeSemanticDigest: sourcePair.BeforeSemanticDigest, AfterSemanticDigest: sourcePair.AfterSemanticDigest}
		if consumerErr == nil {
			row.ClaimRecreatedDueOnlyToRaw, row.ClaimRecreatedDueOnlyToRawTotal = rawEvidenceIdentityProbe(consumerRecords)
			report.ClaimRecreatedDueOnlyToRaw += row.ClaimRecreatedDueOnlyToRaw
			report.ClaimRecreatedDueOnlyToRawTotal += row.ClaimRecreatedDueOnlyToRawTotal
		}
		if producerErr == nil && consumerErr == nil && consumerOldErr == nil && oldOK && newOK && row.OldArtifactExact && recordIDsEqual(row.NewExpectedIDs, row.ProducerObservedIDs) && recordIDsEqual(row.NewExpectedIDs, row.ConsumerObservedIDs) && row.ProducerConsumerExact && row.NewExpectationProducerExact && row.NewExpectationConsumerExact {
			row.StableIdentityPreserved, row.StableIdentityTotal = stableRecordIntersection(newRow.ExpectedClaims, producerRecords)
			row.EvidenceOnlyChanges, row.EvidenceOnlyTotal = evidenceOnlyChanges(newRow.ExpectedClaims, producerRecords)
			row.PropositionTargetChanges = legacyPropositionTargetChanges(oldRecords, newRow.ExpectedClaims)
			row.ConsumerPropositionTargetChanges = consumerPropositionTargetChanges(consumerOldRecords, consumerRecords)
			if !propositionChangesEqual(row.PropositionTargetChanges, row.ConsumerPropositionTargetChanges) {
				row.Decision, row.Resolution, row.Stage, row.Step, row.Reason = producer.DecisionFailClosed, producer.ResolutionLower, "claim-identity-evolution", "compare-source-derived-records", "CLAIM_IDENTITY_EVOLUTION_PROPOSITION_MAPPING_MISMATCH"
				row.PropositionTargetChanges = nil
				row.ConsumerPropositionTargetChanges = nil
				report.Cases = append(report.Cases, row)
				continue
			}
			row.Decision, row.Resolution, row.Stage, row.Step, row.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "claim-identity-evolution", "compare-source-derived-records", "CLAIM_IDENTITY_EVOLUTION_EXACT"
			report.EvolutionRowsReconstructed++
			report.StableIdentityPreserved += row.StableIdentityPreserved
			report.StableIdentityTotal += row.StableIdentityTotal
			report.PersistentClaimIdentity += row.StableIdentityPreserved
			report.PersistentClaimIdentityTotal += row.StableIdentityTotal
			report.EvolutionClaimRowsReconstructed += row.StableIdentityPreserved
			report.EvolutionClaimRowsTotal += row.StableIdentityTotal
			report.EvidenceOnlyChanges += row.EvidenceOnlyChanges
			report.PropositionChanges += len(row.PropositionTargetChanges)
			if definition.ID == "equivalent" {
				report.RawEvidenceChangedNonsemantic, report.RawEvidenceNonsemanticTotal, report.SemanticTargetPreserved, report.SemanticTargetNonsemanticTotal = nonsemanticEvidenceMetrics(producerRecords)
			}
		}
		report.Cases = append(report.Cases, row)
	}
	if report.EvolutionRowsReconstructed == report.EvolutionRowsTotal && report.EvolutionClaimRowsReconstructed == 31 && report.EvolutionClaimRowsTotal == 31 && report.PersistentClaimIdentity == 31 && report.PersistentClaimIdentityTotal == 31 && report.ClaimRecreatedDueOnlyToRaw == 0 && report.ClaimRecreatedDueOnlyToRawTotal == 31 && report.DenominatorUnchanged && report.NewArtifactDigest == bytesDigest(newRaw) {
		report.Decision, report.Resolution, report.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "CLAIM_IDENTITY_EVOLUTION_EXACT"
		report.ChangeKind = "STABLE_IDENTITY_V3_EVIDENCE_SEPARATED"
	} else {
		report.ChangeKind = "STABLE_IDENTITY_V3_EVOLUTION_FAIL_CLOSED"
	}
	return report, nil
}

func decodeEvolutionExpectation(raw []byte) (evolutionExpectation, error) {
	var value evolutionExpectation
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("trailing expectation data")
	}
	if value.Schema != "gooo/semantic-delta-claim-transition-expectations/v2" || value.DenominatorID != "gooo://semantic-delta-receipt-denominator/v2" || value.ClaimCountContractVersion != "v1" || value.DenominatorEvolutionReceipt != "REQUIRED_FOR_FIXED_CLAIM_COUNT_CHANGE" || value.FixedClaimTotal != 31 || value.FixedCaseTotal != 5 || !fixedEvolutionCaseInventory(value.Cases) {
		return value, fmt.Errorf("new fixed expectation contract mismatch")
	}
	return value, nil
}

func decodeLegacyExpectation(raw []byte) (legacyExpectation, error) {
	var value legacyExpectation
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("trailing old expectation data")
	}
	if value.Schema != "gooo/semantic-delta-claim-transition-expectations/v1" || value.DenominatorID != "gooo://semantic-delta-receipt-denominator/v2" || value.ClaimCountContractVersion != "v1" || value.DenominatorEvolutionReceipt != "REQUIRED_FOR_FIXED_CLAIM_COUNT_CHANGE" || value.FixedClaimTotal != 31 || value.FixedCaseTotal != 5 || !legacyExpectationInventory(value.Cases) {
		return value, fmt.Errorf("old fixed expectation contract mismatch")
	}
	return value, nil
}

func bytesDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fixedEvolutionCaseInventory(cases []evolutionExpectationRow) bool {
	if len(cases) != 5 {
		return false
	}
	want := map[string]bool{"equivalent": true, "semantic-change": true, "value-program-change": true, "indeterminate": true, "ambiguous-match": true}
	seen := map[string]bool{}
	for _, row := range cases {
		observedIDs := make([]string, 0, len(row.ExpectedClaims))
		for _, record := range row.ExpectedClaims {
			observedIDs = append(observedIDs, record.StableID)
		}
		if !want[row.ID] || seen[row.ID] || row.ExpectedClaimTotal <= 0 || len(row.ExpectedClaimIDs) != row.ExpectedClaimTotal || len(row.ExpectedClaims) != row.ExpectedClaimTotal || uniqueStringCount(row.ExpectedClaimIDs) != len(row.ExpectedClaimIDs) || uniqueStringCount(observedIDs) != len(observedIDs) || !recordIDsEqual(row.ExpectedClaimIDs, observedIDs) || !consumer.ValidateClaimIdentityRecords(row.ExpectedClaims) {
			return false
		}
		seen[row.ID] = true
	}
	return len(seen) == len(want)
}

func legacyExpectationInventory(cases []legacyExpectationRow) bool {
	if len(cases) != 5 {
		return false
	}
	want := map[string]bool{"equivalent": true, "semantic-change": true, "value-program-change": true, "indeterminate": true, "ambiguous-match": true}
	seen := map[string]bool{}
	for _, row := range cases {
		if !want[row.ID] || seen[row.ID] || len(row.ExpectedClaimIDs) == 0 || uniqueStringCount(row.ExpectedClaimIDs) != len(row.ExpectedClaimIDs) {
			return false
		}
		seen[row.ID] = true
	}
	return len(seen) == len(want)
}

func uniqueStringCount(values []string) int {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		seen[value] = true
	}
	return len(seen)
}

func producerRecordIDs(records []producer.ClaimIdentityRecord) []string {
	result := make([]string, 0, len(records))
	for _, record := range records {
		result = append(result, record.StableID)
	}
	return sortedIDs(result)
}
func consumerRecordIDs(records []consumer.ClaimIdentityRecord) []string {
	result := make([]string, 0, len(records))
	for _, record := range records {
		result = append(result, record.StableID)
	}
	return sortedIDs(result)
}
func sortedIDs(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
func recordIDsEqual(left, right []string) bool {
	return bytes.Equal(mustJSON(sortedIDs(left)), mustJSON(sortedIDs(right)))
}
func mustJSON(value any) []byte { raw, _ := json.Marshal(value); return raw }

func setDiff(oldIDs, newIDs []string) ([]string, []string) {
	oldSet, newSet := map[string]bool{}, map[string]bool{}
	for _, id := range oldIDs {
		oldSet[id] = true
	}
	for _, id := range newIDs {
		newSet[id] = true
	}
	removed, added := []string{}, []string{}
	for id := range oldSet {
		if !newSet[id] {
			removed = append(removed, id)
		}
	}
	for id := range newSet {
		if !oldSet[id] {
			added = append(added, id)
		}
	}
	return sortedIDs(removed), sortedIDs(added)
}

func producerRecordsEqual(left []producer.ClaimIdentityRecord, right []consumer.ClaimIdentityRecord) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]producer.ClaimIdentityRecord(nil), left...)
	rightCopy := append([]consumer.ClaimIdentityRecord(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].StableID < leftCopy[j].StableID })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].StableID < rightCopy[j].StableID })
	for index := range leftCopy {
		if !producerConsumerRecordEqual(leftCopy[index], rightCopy[index]) {
			return false
		}
	}
	return true
}

func expectedProducerRecordsEqual(expected []consumer.ClaimIdentityRecord, observed []producer.ClaimIdentityRecord) bool {
	if len(expected) != len(observed) {
		return false
	}
	left := append([]consumer.ClaimIdentityRecord(nil), expected...)
	right := append([]producer.ClaimIdentityRecord(nil), observed...)
	sort.Slice(left, func(i, j int) bool { return left[i].StableID < left[j].StableID })
	sort.Slice(right, func(i, j int) bool { return right[i].StableID < right[j].StableID })
	for index := range left {
		if !consumerProducerRecordEqual(left[index], right[index]) {
			return false
		}
	}
	return true
}

func expectedConsumerRecordsEqual(expected, observed []consumer.ClaimIdentityRecord) bool {
	if len(expected) != len(observed) {
		return false
	}
	left := append([]consumer.ClaimIdentityRecord(nil), expected...)
	right := append([]consumer.ClaimIdentityRecord(nil), observed...)
	sort.Slice(left, func(i, j int) bool { return left[i].StableID < left[j].StableID })
	sort.Slice(right, func(i, j int) bool { return right[i].StableID < right[j].StableID })
	for index := range left {
		if !consumerConsumerRecordEqual(left[index], right[index]) {
			return false
		}
	}
	return true
}

func producerConsumerRecordEqual(left producer.ClaimIdentityRecord, right consumer.ClaimIdentityRecord) bool {
	return left.StableID == right.StableID && left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.PreservationOf == right.PreservationOf && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func consumerProducerRecordEqual(left consumer.ClaimIdentityRecord, right producer.ClaimIdentityRecord) bool {
	return left.StableID == right.StableID && left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.PreservationOf == right.PreservationOf && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func consumerConsumerRecordEqual(left, right consumer.ClaimIdentityRecord) bool {
	return left.StableID == right.StableID && left.Kind == right.Kind && left.RelationRole == right.RelationRole && left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest && left.PreservationOf == right.PreservationOf && left.BeforeSourcePath == right.BeforeSourcePath && left.AfterSourcePath == right.AfterSourcePath && left.EvidenceBeforeRawDigest == right.EvidenceBeforeRawDigest && left.EvidenceAfterRawDigest == right.EvidenceAfterRawDigest && left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func stableRecordIntersection(expected []consumer.ClaimIdentityRecord, observed []producer.ClaimIdentityRecord) (int, int) {
	byID := map[string]producer.ClaimIdentityRecord{}
	for _, record := range observed {
		byID[record.StableID] = record
	}
	preserved := 0
	for _, record := range expected {
		if other, ok := byID[record.StableID]; ok && record.PropositionDigest == other.PropositionDigest && record.TargetAddress == other.TargetAddress {
			preserved++
		}
	}
	return preserved, len(expected)
}

func evidenceOnlyChanges(expected []consumer.ClaimIdentityRecord, observed []producer.ClaimIdentityRecord) (int, int) {
	byID := map[string]producer.ClaimIdentityRecord{}
	for _, record := range observed {
		byID[record.StableID] = record
	}
	changed := 0
	for _, record := range expected {
		if other, ok := byID[record.StableID]; ok && record.PropositionDigest == other.PropositionDigest && record.TargetAddress == other.TargetAddress && (record.EvidenceBeforeRawDigest != other.EvidenceBeforeRawDigest || record.EvidenceAfterRawDigest != other.EvidenceAfterRawDigest || record.EvidenceBeforeSemanticDigest != other.EvidenceBeforeSemanticDigest || record.EvidenceAfterSemanticDigest != other.EvidenceAfterSemanticDigest) {
			changed++
		}
	}
	return changed, len(expected)
}

// rawEvidenceIdentityProbe changes only observation/provenance fields. A
// record that then fails the consumer's stable-identity validator would have
// recreated its identity from raw evidence, which v3 forbids.
func rawEvidenceIdentityProbe(records []consumer.ClaimIdentityRecord) (int, int) {
	recreated := 0
	for _, record := range records {
		probe := record
		probe.BeforeSourcePath += "#evidence-only-probe"
		probe.AfterSourcePath += "#evidence-only-probe"
		probe.EvidenceBeforeRawDigest = "sha256:" + strings.Repeat("0", 64)
		probe.EvidenceAfterRawDigest = "sha256:" + strings.Repeat("1", 64)
		probe.EvidenceBeforeSemanticDigest = "sha256:" + strings.Repeat("2", 64)
		probe.EvidenceAfterSemanticDigest = "sha256:" + strings.Repeat("3", 64)
		if !consumer.ValidateClaimIdentityRecords([]consumer.ClaimIdentityRecord{probe}) {
			recreated++
		}
	}
	return recreated, len(records)
}

func propositionTargetChanges(expected []consumer.ClaimIdentityRecord, observed []producer.ClaimIdentityRecord) []propositionTargetChange {
	byID := map[string]producer.ClaimIdentityRecord{}
	for _, record := range observed {
		byID[record.StableID] = record
	}
	result := []propositionTargetChange{}
	for _, record := range expected {
		if other, ok := byID[record.StableID]; ok && (record.PropositionDigest != other.PropositionDigest || record.TargetAddress != other.TargetAddress) {
			result = append(result, propositionTargetChange{OldID: record.StableID, NewID: other.StableID, OldPropositionDigest: record.PropositionDigest, NewPropositionDigest: other.PropositionDigest, OldTargetAddress: record.TargetAddress, NewTargetAddress: other.TargetAddress, Reason: "SOURCE_DERIVED_PROPOSITION_OR_TARGET_CHANGED"})
		}
	}
	return result
}

type legacyIdentity struct {
	ID                string
	Kind              string
	ClaimTypeID       string
	RelationRole      string
	Normalized        string
	PropositionDigest string
	TargetAddress     string
}

func legacyClaimIdentity(receipt producer.Receipt, input producer.Input) ([]string, []legacyIdentity) {
	beforeClaims := receipt.Before.Claims
	afterClaims := receipt.After.Claims
	before := make([]legacyIdentity, 0, len(beforeClaims))
	after := make([]legacyIdentity, 0, len(afterClaims))
	for _, claim := range beforeClaims {
		before = append(before, legacyObjectIdentity(claim, input.BeforePath, receipt.Before.SourceDigest, receipt.Before.SemanticDigest, "before"))
	}
	for _, claim := range afterClaims {
		after = append(after, legacyObjectIdentity(claim, input.AfterPath, receipt.After.SourceDigest, receipt.After.SemanticDigest, "after"))
	}
	result := make([]legacyIdentity, 0, 1+len(before)+len(after))
	boundedTarget := input.BeforePath + "->" + input.AfterPath
	boundedObject := input.BeforePath + "\x00" + input.AfterPath + "\x00" + receipt.Before.SourceDigest + "\x00" + receipt.After.SourceDigest + "\x00" + receipt.Before.SemanticDigest + "\x00" + receipt.After.SemanticDigest
	boundedNormalized := strings.Join([]string{"BOUNDED_SEMANTIC_EQUIVALENCE", "source-pair", "bounded-semantic-equivalence", boundedObject}, "\x00")
	result = append(result, legacyIdentity{ID: "gooo://semantic-delta/claim/bounded-equivalence/" + jsonDigest(strings.Join([]string{boundedNormalized}, "\x00"))[len("sha256:"):], Kind: "BOUNDED_SEMANTIC_EQUIVALENCE", RelationRole: "bounded-equivalence", Normalized: boundedNormalized, PropositionDigest: jsonDigest(boundedNormalized), TargetAddress: boundedTarget})
	for _, claim := range before {
		result = append(result, claim)
	}
	for _, claim := range after {
		result = append(result, claim)
	}
	if receipt.Classification == producer.ClassPreserved || receipt.Classification == producer.ClassChanged {
		for _, oldBefore := range before {
			match := legacyIdentity{}
			for _, candidate := range after {
				if oldBefore.Normalized == candidate.Normalized {
					match = candidate
					break
				}
			}
			object := ""
			afterRaw, afterSemantic := "", ""
			if match.ID != "" {
				object, afterRaw, afterSemantic = match.TargetAddress, receipt.After.SourceDigest, receipt.After.SemanticDigest
			}
			identity := jsonDigest(strings.Join([]string{oldBefore.ID, match.ID, object, oldBefore.Normalized, afterRaw, afterSemantic}, "\x00"))
			normalized := strings.Join([]string{"BEFORE_CLAIM_PRESERVATION", oldBefore.ClaimTypeID, "preserves", oldBefore.Normalized}, "\x00")
			result = append(result, legacyIdentity{ID: "gooo://semantic-delta/claim/preservation/" + identity[len("sha256:"):], Kind: "BEFORE_CLAIM_PRESERVATION", ClaimTypeID: oldBefore.ClaimTypeID, RelationRole: "preserves", Normalized: normalized, PropositionDigest: jsonDigest(normalized), TargetAddress: oldBefore.TargetAddress})
		}
	}
	ids := make([]string, 0, len(result))
	for _, identity := range result {
		ids = append(ids, identity.ID)
	}
	return sortedIDs(ids), result
}

func legacyObjectIdentity(claim producer.Claim, path, rawDigest, semanticDigest, side string) legacyIdentity {
	identity := jsonDigest(strings.Join([]string{claim.NormalizedProposition, path, rawDigest, semanticDigest}, "\x00"))
	return legacyIdentity{ID: "gooo://semantic-delta/claim/object/" + identity[len("sha256:"):], Kind: claim.Kind, ClaimTypeID: claim.ClaimTypeID, RelationRole: claim.Predicate + "|observation|" + side, Normalized: claim.NormalizedProposition, PropositionDigest: claim.PropositionDigest, TargetAddress: path}
}

func legacyPropositionTargetChanges(old []legacyIdentity, newer []consumer.ClaimIdentityRecord) []propositionTargetChange {
	byKey := make(map[string]legacyIdentity, len(old))
	for _, identity := range old {
		key := identity.Kind + "\x00" + identity.Normalized + "\x00" + identity.RelationRole
		byKey[key] = identity
	}
	result := []propositionTargetChange{}
	for _, identity := range newer {
		key := identity.Kind + "\x00" + identity.NormalizedProposition + "\x00" + identity.RelationRole
		oldIdentity, ok := byKey[key]
		if !ok {
			continue
		}
		if oldIdentity.PropositionDigest != identity.PropositionDigest || oldIdentity.TargetAddress != identity.TargetAddress {
			result = append(result, propositionTargetChange{OldID: oldIdentity.ID, NewID: identity.StableID, OldPropositionDigest: oldIdentity.PropositionDigest, NewPropositionDigest: identity.PropositionDigest, OldTargetAddress: oldIdentity.TargetAddress, NewTargetAddress: identity.TargetAddress, Reason: "CLAIM_IDENTITY_V3_MOVED_RAW_EVIDENCE_OUT_OF_IDENTITY"})
		}
	}
	return result
}

func consumerPropositionTargetChanges(old, newer []consumer.ClaimIdentityRecord) []propositionTargetChange {
	byKey := make(map[string]consumer.ClaimIdentityRecord, len(old))
	for _, identity := range old {
		key := identity.Kind + "\x00" + identity.NormalizedProposition + "\x00" + identity.RelationRole
		byKey[key] = identity
	}
	result := []propositionTargetChange{}
	for _, identity := range newer {
		key := identity.Kind + "\x00" + identity.NormalizedProposition + "\x00" + identity.RelationRole
		oldIdentity, ok := byKey[key]
		if !ok {
			continue
		}
		if oldIdentity.PropositionDigest != identity.PropositionDigest || oldIdentity.TargetAddress != identity.TargetAddress {
			result = append(result, propositionTargetChange{OldID: oldIdentity.StableID, NewID: identity.StableID, OldPropositionDigest: oldIdentity.PropositionDigest, NewPropositionDigest: identity.PropositionDigest, OldTargetAddress: oldIdentity.TargetAddress, NewTargetAddress: identity.TargetAddress, Reason: "CLAIM_IDENTITY_V3_MOVED_RAW_EVIDENCE_OUT_OF_IDENTITY"})
		}
	}
	return result
}

func propositionChangesEqual(left, right []propositionTargetChange) bool {
	leftRaw, rightRaw := append([]propositionTargetChange(nil), left...), append([]propositionTargetChange(nil), right...)
	sort.Slice(leftRaw, func(i, j int) bool {
		if leftRaw[i].OldID != leftRaw[j].OldID {
			return leftRaw[i].OldID < leftRaw[j].OldID
		}
		return leftRaw[i].NewID < leftRaw[j].NewID
	})
	sort.Slice(rightRaw, func(i, j int) bool {
		if rightRaw[i].OldID != rightRaw[j].OldID {
			return rightRaw[i].OldID < rightRaw[j].OldID
		}
		return rightRaw[i].NewID < rightRaw[j].NewID
	})
	return bytes.Equal(mustJSON(leftRaw), mustJSON(rightRaw))
}

func jsonDigest(value any) string { return bytesDigest(mustJSON(value)) }

func nonsemanticEvidenceMetrics(records []producer.ClaimIdentityRecord) (changed, total, preserved int, semanticTotal int) {
	type evidencePair struct {
		before, after       producer.ClaimIdentityRecord
		hasBefore, hasAfter bool
	}
	pairs := map[string]*evidencePair{}
	for _, record := range records {
		key := record.Kind + "\x00" + record.NormalizedProposition + "\x00" + record.TargetAddress
		role := record.RelationRole
		side := ""
		if strings.HasSuffix(role, "|before") {
			side, role = "before", strings.TrimSuffix(role, "|before")
		} else if strings.HasSuffix(role, "|after") {
			side, role = "after", strings.TrimSuffix(role, "|after")
		}
		key += "\x00" + role
		pair := pairs[key]
		if pair == nil {
			pair = &evidencePair{}
			pairs[key] = pair
		}
		switch side {
		case "before":
			pair.before, pair.hasBefore = record, true
		case "after":
			pair.after, pair.hasAfter = record, true
		default:
			pair.before, pair.after, pair.hasBefore, pair.hasAfter = record, record, true, true
		}
	}
	for _, pair := range pairs {
		if !pair.hasBefore || !pair.hasAfter {
			continue
		}
		total++
		rawChanged := pair.before.EvidenceBeforeRawDigest != pair.after.EvidenceAfterRawDigest
		semanticSame := pair.before.EvidenceBeforeSemanticDigest == pair.after.EvidenceAfterSemanticDigest
		if rawChanged {
			changed++
		}
		if semanticSame {
			preserved++
		}
	}
	return changed, total, preserved, total
}
