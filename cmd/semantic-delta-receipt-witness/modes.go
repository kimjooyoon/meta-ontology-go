package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

const semanticClaimDeltaManifestSchema = "gooo/semantic-delta-claim-delta-fixtures/v1"

type semanticClaimDeltaManifest struct {
	Schema string                          `json:"schema"`
	Cases  []semanticClaimDeltaManifestRow `json:"cases"`
}

type semanticClaimDeltaManifestRow struct {
	ID                 string   `json:"id"`
	BeforePath         string   `json:"before_path"`
	AfterPath          string   `json:"after_path"`
	ExpectedAddedIDs   []string `json:"expected_added_ids"`
	ExpectedRemovedIDs []string `json:"expected_removed_ids"`
}

type semanticClaimDeltaManifestReport struct {
	Schema            string                             `json:"schema"`
	ManifestPath      string                             `json:"manifest_path"`
	ManifestBytes     int                                `json:"manifest_bytes"`
	ManifestDigest    string                             `json:"manifest_digest"`
	RowsReconstructed int                                `json:"rows_reconstructed"`
	RowsTotal         int                                `json:"rows_total"`
	RowsExact         int                                `json:"rows_exact"`
	Decision          string                             `json:"decision"`
	Resolution        string                             `json:"resolution"`
	Stage             string                             `json:"stage"`
	Step              string                             `json:"step"`
	Reason            string                             `json:"reason"`
	Cases             []semanticClaimDeltaManifestResult `json:"cases"`
}

type semanticClaimDeltaManifestResult struct {
	CaseID                string   `json:"case_id"`
	BeforePath            string   `json:"before_path"`
	AfterPath             string   `json:"after_path"`
	ExpectedAddedIDs      []string `json:"expected_added_ids"`
	ExpectedRemovedIDs    []string `json:"expected_removed_ids"`
	ProducerAddedIDs      []string `json:"producer_added_ids"`
	ProducerRemovedIDs    []string `json:"producer_removed_ids"`
	ConsumerAddedIDs      []string `json:"consumer_added_ids"`
	ConsumerRemovedIDs    []string `json:"consumer_removed_ids"`
	ProducerConsumerExact bool     `json:"producer_consumer_exact"`
	Decision              string   `json:"decision"`
	Resolution            string   `json:"resolution"`
	Stage                 string   `json:"stage"`
	Step                  string   `json:"step"`
	Reason                string   `json:"reason"`
}

type persistenceProbeReport struct {
	Schema               string                 `json:"schema"`
	ProducerBaseline     persistenceObservation `json:"producer_baseline"`
	ProducerAlternate    persistenceObservation `json:"producer_alternate"`
	ConsumerBaseline     persistenceObservation `json:"consumer_baseline"`
	ConsumerAlternate    persistenceObservation `json:"consumer_alternate"`
	ProducerPersistence  persistenceMapping     `json:"producer_persistence"`
	ConsumerPersistence  persistenceMapping     `json:"consumer_persistence"`
	ExpectedClaimTotal   int                    `json:"expected_claim_total"`
	ReconstructionExact  bool                   `json:"reconstruction_exact"`
	PersistenceSatisfied bool                   `json:"persistence_satisfied"`
	Decision             string                 `json:"decision"`
	Resolution           string                 `json:"resolution"`
	Stage                string                 `json:"stage"`
	Step                 string                 `json:"step"`
	Reason               string                 `json:"reason"`
}

func decodeSemanticClaimDeltaManifest(raw []byte) (semanticClaimDeltaManifest, error) {
	var value semanticClaimDeltaManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("trailing semantic claim delta manifest data")
	}
	if value.Schema != semanticClaimDeltaManifestSchema || len(value.Cases) != 2 {
		return value, fmt.Errorf("semantic claim delta manifest contract mismatch")
	}
	seen := map[string]bool{}
	for _, row := range value.Cases {
		if row.ID == "" || seen[row.ID] || row.BeforePath == "" || row.AfterPath == "" || !uniqueIDs(row.ExpectedAddedIDs) || !uniqueIDs(row.ExpectedRemovedIDs) {
			return value, fmt.Errorf("semantic claim delta manifest inventory mismatch")
		}
		seen[row.ID] = true
	}
	if !seen["claim-removal"] || !seen["claim-addition"] {
		return value, fmt.Errorf("semantic claim delta manifest case IDs mismatch")
	}
	return value, nil
}

func reconstructSemanticClaimDeltaManifest(path string, subjectSHA, observedCheckoutSHA string) (semanticClaimDeltaManifestReport, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return semanticClaimDeltaManifestReport{}, fmt.Errorf("read semantic claim delta manifest: %w", err)
	}
	manifest, err := decodeSemanticClaimDeltaManifest(raw)
	if err != nil {
		return semanticClaimDeltaManifestReport{}, fmt.Errorf("decode semantic claim delta manifest: %w", err)
	}
	report := semanticClaimDeltaManifestReport{
		Schema: semanticClaimDeltaManifestSchema, ManifestPath: path, ManifestBytes: len(raw), ManifestDigest: bytesDigest(raw), RowsTotal: len(manifest.Cases),
		Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "semantic-claim-delta-manifest", Step: "reconstruct-raw-source-pairs", Reason: "SEMANTIC_CLAIM_DELTA_MANIFEST_UNKNOWN",
	}
	for _, fixed := range manifest.Cases {
		input := producer.Input{CaseID: fixed.ID, BeforePath: fixed.BeforePath, AfterPath: fixed.AfterPath, SubjectSHA: subjectSHA, ObservedCheckoutSHA: observedCheckoutSHA}
		producerReceipt, producerErr := producer.ProduceFiles(input)
		consumerReceipt, consumerErr := consumer.ReconstructReceiptFromFiles(consumer.Input{CaseID: fixed.ID, BeforePath: fixed.BeforePath, AfterPath: fixed.AfterPath, SubjectSHA: subjectSHA, ObservedCheckoutSHA: observedCheckoutSHA})
		row := semanticClaimDeltaManifestResult{CaseID: fixed.ID, BeforePath: fixed.BeforePath, AfterPath: fixed.AfterPath, ExpectedAddedIDs: sortedIDs(fixed.ExpectedAddedIDs), ExpectedRemovedIDs: sortedIDs(fixed.ExpectedRemovedIDs), Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "semantic-claim-delta-manifest", Step: "reconstruct-raw-source-pairs", Reason: "SEMANTIC_CLAIM_DELTA_MANIFEST_UNKNOWN"}
		if producerErr != nil || consumerErr != nil {
			row.Reason = "SOURCE_PAIR_UNAVAILABLE"
			report.Cases = append(report.Cases, row)
			continue
		}
		row.ProducerAddedIDs, row.ProducerRemovedIDs = producerClaimDeltaIDs(producerReceipt.SemanticClaimDelta)
		row.ConsumerAddedIDs, row.ConsumerRemovedIDs = consumerClaimDeltaIDs(consumerReceipt.SemanticClaimDelta)
		row.ProducerConsumerExact = recordIDsEqual(row.ProducerAddedIDs, row.ConsumerAddedIDs) && recordIDsEqual(row.ProducerRemovedIDs, row.ConsumerRemovedIDs)
		row.ProducerConsumerExact = row.ProducerConsumerExact && recordIDsEqual(row.ExpectedAddedIDs, row.ProducerAddedIDs) && recordIDsEqual(row.ExpectedRemovedIDs, row.ProducerRemovedIDs)
		if row.ProducerConsumerExact {
			row.Decision, row.Resolution, row.Stage, row.Step, row.Reason = "PASS", producer.ResolutionExact, "semantic-claim-delta-manifest", "compare-independent-raw-source-projections", "SEMANTIC_CLAIM_DELTA_MANIFEST_EXACT"
			report.RowsExact++
		} else {
			row.Reason = "SEMANTIC_CLAIM_DELTA_MANIFEST_MISMATCH"
		}
		report.RowsReconstructed++
		report.Cases = append(report.Cases, row)
	}
	if report.RowsExact == report.RowsTotal {
		report.Decision, report.Resolution, report.Stage, report.Step, report.Reason = "PASS", producer.ResolutionExact, "semantic-claim-delta-manifest", "compare-independent-raw-source-projections", "SEMANTIC_CLAIM_DELTA_MANIFEST_EXACT"
	}
	return report, nil
}

func producerClaimDeltaIDs(delta producer.ClaimDelta) ([]string, []string) {
	added := make([]string, 0, len(delta.Added))
	for _, claim := range delta.Added {
		added = append(added, claim.ID)
	}
	removed := make([]string, 0, len(delta.Removed))
	for _, claim := range delta.Removed {
		removed = append(removed, claim.ID)
	}
	return sortedIDs(added), sortedIDs(removed)
}

func consumerClaimDeltaIDs(delta consumer.ClaimDelta) ([]string, []string) {
	added := make([]string, 0, len(delta.Added))
	for _, claim := range delta.Added {
		added = append(added, claim.ID)
	}
	removed := make([]string, 0, len(delta.Removed))
	for _, claim := range delta.Removed {
		removed = append(removed, claim.ID)
	}
	return sortedIDs(added), sortedIDs(removed)
}

func uniqueIDs(ids []string) bool {
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

func runPersistenceProbe(options options) persistenceProbeReport {
	producerInput := producer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceBefore, AfterPath: options.persistenceAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA}
	producerAlternateInput := producer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceAlternateBefore, AfterPath: options.persistenceAlternateAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA}
	producerBaseline, producerBaselineErr := producer.ClaimIdentityObservationFromFiles(producerInput)
	producerAlternate, producerAlternateErr := producer.ClaimIdentityObservationFromFiles(producerAlternateInput)
	consumerBaseline, consumerBaselinePair, consumerBaselineErr := consumer.ClaimIdentityRecordsFromFiles(consumer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceBefore, AfterPath: options.persistenceAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA})
	consumerAlternate, consumerAlternatePair, consumerAlternateErr := consumer.ClaimIdentityRecordsFromFiles(consumer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceAlternateBefore, AfterPath: options.persistenceAlternateAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA})
	producerMappingValue := producerMapping(producer.CompareClaimIdentityRecords(producerBaseline.Records, producerAlternate.Records))
	consumerMappingValue := consumerMapping(consumer.CompareClaimIdentityRecords(consumerBaseline, consumerAlternate))
	expectedTotal := len(producerBaseline.Records)
	reconstructionExact := producerBaselineErr == nil && producerAlternateErr == nil && consumerBaselineErr == nil && consumerAlternateErr == nil && producerRecordsEqual(producerBaseline.Records, consumerBaseline) && producerRecordsEqual(producerAlternate.Records, consumerAlternate) && len(producerAlternate.Records) == expectedTotal
	mappingExact := persistenceMappingsEqual(producerMappingValue, consumerMappingValue)
	persistenceSatisfied := reconstructionExact && mappingExact && persistenceMappingSatisfies(producerMappingValue, expectedTotal) && persistenceMappingSatisfies(consumerMappingValue, expectedTotal)
	report := persistenceProbeReport{Schema: "gooo/semantic-delta-claim-identity-persistence-probe/v1", ProducerBaseline: persistenceObservation{SourcePair: producerSourcePair(producerBaseline), Records: producerRecordSnapshots(producerBaseline.Records)}, ProducerAlternate: persistenceObservation{SourcePair: producerSourcePair(producerAlternate), Records: producerRecordSnapshots(producerAlternate.Records)}, ConsumerBaseline: persistenceObservation{SourcePair: evolutionSourcePair{BeforePath: consumerBaselinePair.BeforePath, AfterPath: consumerBaselinePair.AfterPath, BeforeRawDigest: consumerBaselinePair.BeforeRawDigest, AfterRawDigest: consumerBaselinePair.AfterRawDigest, BeforeSemanticDigest: consumerBaselinePair.BeforeSemanticDigest, AfterSemanticDigest: consumerBaselinePair.AfterSemanticDigest}, Records: consumerRecordSnapshots(consumerBaseline)}, ConsumerAlternate: persistenceObservation{SourcePair: evolutionSourcePair{BeforePath: consumerAlternatePair.BeforePath, AfterPath: consumerAlternatePair.AfterPath, BeforeRawDigest: consumerAlternatePair.BeforeRawDigest, AfterRawDigest: consumerAlternatePair.AfterRawDigest, BeforeSemanticDigest: consumerAlternatePair.BeforeSemanticDigest, AfterSemanticDigest: consumerAlternatePair.AfterSemanticDigest}, Records: consumerRecordSnapshots(consumerAlternate)}, ProducerPersistence: producerMappingValue, ConsumerPersistence: consumerMappingValue, ExpectedClaimTotal: expectedTotal, ReconstructionExact: reconstructionExact, PersistenceSatisfied: persistenceSatisfied, Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "claim-identity-persistence", Step: "compare-v3-observations", Reason: "INDEPENDENT_RECONSTRUCTION_MISMATCH"}
	if reconstructionExact && persistenceSatisfied {
		report.Decision, report.Resolution, report.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "V3_CLAIM_IDENTITY_PERSISTED_ACROSS_RAW_INTERVENTION"
	} else if reconstructionExact {
		report.Reason = persistenceFailureReason(producerMappingValue, consumerMappingValue, expectedTotal)
	}
	return report
}
