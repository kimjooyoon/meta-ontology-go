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
	Schema                       string                      `json:"schema"`
	ProducerBaseline             persistenceObservation      `json:"producer_baseline"`
	ProducerAlternate            persistenceObservation      `json:"producer_alternate"`
	ConsumerBaseline             persistenceObservation      `json:"consumer_baseline"`
	ConsumerAlternate            persistenceObservation      `json:"consumer_alternate"`
	IdentityFault                *identityFaultEvidence      `json:"identity_fault,omitempty"`
	ProducerIdentityFaultGraph   *identityFaultGraphEvidence `json:"producer_identity_fault_graph,omitempty"`
	ConsumerIdentityFaultGraph   *identityFaultGraphEvidence `json:"consumer_identity_fault_graph,omitempty"`
	ProducerFaultedAlternate     *persistenceObservation     `json:"producer_faulted_alternate,omitempty"`
	ConsumerFaultedAlternate     *persistenceObservation     `json:"consumer_faulted_alternate,omitempty"`
	ProducerRawSemanticPreserved bool                        `json:"producer_raw_semantic_preserved"`
	ConsumerRawSemanticPreserved bool                        `json:"consumer_raw_semantic_preserved"`
	ProducerFaultGraphClosed     bool                        `json:"producer_fault_graph_closed"`
	ConsumerFaultGraphClosed     bool                        `json:"consumer_fault_graph_closed"`
	ProducerPersistence          persistenceMapping          `json:"producer_persistence"`
	ConsumerPersistence          persistenceMapping          `json:"consumer_persistence"`
	ExpectedClaimTotal           int                         `json:"expected_claim_total"`
	ReconstructionExact          bool                        `json:"reconstruction_exact"`
	PersistenceSatisfied         bool                        `json:"persistence_satisfied"`
	Decision                     string                      `json:"decision"`
	Resolution                   string                      `json:"resolution"`
	Stage                        string                      `json:"stage"`
	Step                         string                      `json:"step"`
	Reason                       string                      `json:"reason"`
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
	producerBaselinePair := producerSourcePair(producerBaseline)
	producerAlternatePair := producerSourcePair(producerAlternate)
	consumerBaselineSourcePair := consumerSourcePair(consumerBaselinePair)
	consumerAlternateSourcePair := consumerSourcePair(consumerAlternatePair)
	producerAlternateForComparison := producerAlternate.Records
	consumerAlternateForComparison := consumerAlternate
	var faultArtifact identityFaultArtifact
	var faultEvidence *identityFaultEvidence
	var producerFaulted producerIdentityFaultResult
	var consumerFaulted consumerIdentityFaultResult
	if options.identityFault != "" {
		artifact, evidence, err := readIdentityFaultArtifact(options.identityFault)
		if err != nil {
			return persistenceProbeReport{Schema: "gooo/semantic-delta-claim-identity-persistence-probe/v1", ExpectedClaimTotal: len(producerBaseline.Records), Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "identity-fault", Step: "read-artifact", Reason: "IDENTITY_FAULT_ARTIFACT_UNAVAILABLE"}
		}
		faultArtifact, evidenceCopy := artifact, evidence
		faultEvidence = &evidenceCopy
		producerFaulted = mutateProducerIdentityFault(producerAlternate.Records, producerAlternatePair, faultArtifact)
		consumerFaulted = mutateConsumerIdentityFault(consumerAlternate, consumerAlternateSourcePair, faultArtifact)
		if producerFaulted.Valid {
			producerAlternateForComparison = normalizeProducerFaultedReferences(producerFaulted.Records, producerFaulted.NewToOld)
		} else {
			producerAlternateForComparison = producerFaulted.Records
		}
		if consumerFaulted.Valid {
			consumerAlternateForComparison = normalizeConsumerFaultedReferences(consumerFaulted.Records, consumerFaulted.NewToOld)
		} else {
			consumerAlternateForComparison = consumerFaulted.Records
		}
	}
	producerMappingValue := producerMapping(producer.CompareClaimIdentityRecords(producerBaseline.Records, producerAlternateForComparison))
	consumerMappingValue := consumerMapping(consumer.CompareClaimIdentityRecords(consumerBaseline, consumerAlternateForComparison))
	if options.identityFault != "" {
		if !producerFaulted.Valid {
			producerMappingValue.Decision, producerMappingValue.Resolution, producerMappingValue.Reason = producer.DecisionFailClosed, producer.ResolutionLower, producerFaulted.FailReason
		}
		if !consumerFaulted.Valid {
			consumerMappingValue.Decision, consumerMappingValue.Resolution, consumerMappingValue.Reason = producer.DecisionFailClosed, producer.ResolutionLower, consumerFaulted.FailReason
		}
	}
	expectedTotal := len(producerBaseline.Records)
	producerRawSemanticPreserved := sourcePairSemanticPreserved(producerBaselinePair, producerAlternatePair)
	consumerRawSemanticPreserved := sourcePairSemanticPreserved(consumerBaselineSourcePair, consumerAlternateSourcePair)
	producerFaultGraphClosed := options.identityFault != "" && producerFaulted.Valid
	consumerFaultGraphClosed := options.identityFault != "" && consumerFaulted.Valid
	// Reconstruction agreement is independent of whether the alternate claim
	// set has additions or removals; that distinction belongs to adjudication.
	reconstructionExact := producerBaselineErr == nil && producerAlternateErr == nil && consumerBaselineErr == nil && consumerAlternateErr == nil && producerRecordsEqual(producerBaseline.Records, consumerBaseline) && producerRecordsEqual(producerAlternate.Records, consumerAlternate)
	if options.identityFault != "" {
		reconstructionExact = reconstructionExact && producerFaultGraphClosed && consumerFaultGraphClosed && producerRecordsEqual(producerFaulted.Records, consumerFaulted.Records)
	}
	mappingExact := persistenceMappingsEqual(producerMappingValue, consumerMappingValue)
	persistenceSatisfied := reconstructionExact && mappingExact && persistenceMappingSatisfies(producerMappingValue, expectedTotal) && persistenceMappingSatisfies(consumerMappingValue, expectedTotal)
	report := persistenceProbeReport{Schema: "gooo/semantic-delta-claim-identity-persistence-probe/v1", ProducerBaseline: persistenceObservation{SourcePair: producerBaselinePair, Records: producerRecordSnapshots(producerBaseline.Records)}, ProducerAlternate: persistenceObservation{SourcePair: producerAlternatePair, Records: producerRecordSnapshots(producerAlternate.Records)}, ConsumerBaseline: persistenceObservation{SourcePair: consumerBaselineSourcePair, Records: consumerRecordSnapshots(consumerBaseline)}, ConsumerAlternate: persistenceObservation{SourcePair: consumerAlternateSourcePair, Records: consumerRecordSnapshots(consumerAlternate)}, IdentityFault: faultEvidence, ProducerPersistence: producerMappingValue, ConsumerPersistence: consumerMappingValue, ExpectedClaimTotal: expectedTotal, ReconstructionExact: reconstructionExact, PersistenceSatisfied: persistenceSatisfied, ProducerRawSemanticPreserved: producerRawSemanticPreserved, ConsumerRawSemanticPreserved: consumerRawSemanticPreserved, ProducerFaultGraphClosed: producerFaultGraphClosed, ConsumerFaultGraphClosed: consumerFaultGraphClosed, Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower, Stage: "claim-identity-persistence", Step: "compare-v3-observations", Reason: "INDEPENDENT_RECONSTRUCTION_MISMATCH"}
	if options.identityFault != "" {
		report.ProducerIdentityFaultGraph = &producerFaulted.Graph
		report.ConsumerIdentityFaultGraph = &consumerFaulted.Graph
		report.ProducerFaultedAlternate = &persistenceObservation{SourcePair: producerAlternatePair, Records: producerRecordSnapshots(producerFaulted.Records)}
		report.ConsumerFaultedAlternate = &persistenceObservation{SourcePair: consumerAlternateSourcePair, Records: consumerRecordSnapshots(consumerFaulted.Records)}
	}
	if reconstructionExact && persistenceSatisfied {
		report.Decision, report.Resolution, report.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "V3_CLAIM_IDENTITY_PERSISTED_ACROSS_RAW_INTERVENTION"
	} else if reconstructionExact {
		report.Reason = persistenceFailureReason(producerMappingValue, consumerMappingValue, expectedTotal)
	} else if options.identityFault != "" {
		report.Reason = firstIdentityFaultFailureReason(producerFaulted.FailReason, consumerFaulted.FailReason)
	}
	return report
}

func firstIdentityFaultFailureReason(producerReason, consumerReason string) string {
	if producerReason != "" {
		return producerReason
	}
	if consumerReason != "" {
		return consumerReason
	}
	return "IDENTITY_REFERENCE_CLOSURE_BROKEN"
}

func consumerSourcePair(observation consumer.SourcePairObservation) evolutionSourcePair {
	return evolutionSourcePair{BeforePath: observation.BeforePath, AfterPath: observation.AfterPath, BeforeRawDigest: observation.BeforeRawDigest, AfterRawDigest: observation.AfterRawDigest, BeforeSemanticDigest: observation.BeforeSemanticDigest, AfterSemanticDigest: observation.AfterSemanticDigest}
}

func sourcePairSemanticPreserved(left, right evolutionSourcePair) bool {
	return left.BeforeSemanticDigest != "" && left.AfterSemanticDigest != "" && left.BeforeSemanticDigest == right.BeforeSemanticDigest && left.AfterSemanticDigest == right.AfterSemanticDigest
}
