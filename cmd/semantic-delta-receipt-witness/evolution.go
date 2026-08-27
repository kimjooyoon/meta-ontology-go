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
	BaselineObservation              persistenceObservation    `json:"baseline_observation"`
	AlternateObservation             persistenceObservation    `json:"alternate_observation"`
	ProducerBaselineObservation      persistenceObservation    `json:"producer_baseline_observation"`
	ProducerAlternateObservation     persistenceObservation    `json:"producer_alternate_observation"`
	ConsumerBaselineObservation      persistenceObservation    `json:"consumer_baseline_observation"`
	ConsumerAlternateObservation     persistenceObservation    `json:"consumer_alternate_observation"`
	ProducerPersistence              persistenceMapping        `json:"producer_persistence"`
	ConsumerPersistence              persistenceMapping        `json:"consumer_persistence"`
	ExpectedPersistenceClaimTotal    int                       `json:"expected_persistence_claim_total"`
	ReconstructionExact              bool                      `json:"reconstruction_exact"`
	PersistenceSatisfied             bool                      `json:"persistence_satisfied"`
	PersistenceExact                 bool                      `json:"persistence_exact"`
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

type persistenceManifest struct {
	Schema          string                    `json:"schema"`
	DenominatorID   string                    `json:"denominator_id"`
	FixedCaseTotal  int                       `json:"fixed_case_total"`
	FixedClaimTotal int                       `json:"fixed_claim_total"`
	Cases           []persistenceManifestCase `json:"cases"`
}

type persistenceManifestCase struct {
	ID                  string `json:"id"`
	ExpectedClaimTotal  int    `json:"expected_claim_total"`
	BaselineBeforePath  string `json:"baseline_before_path"`
	BaselineAfterPath   string `json:"baseline_after_path"`
	AlternateBeforePath string `json:"alternate_before_path"`
	AlternateAfterPath  string `json:"alternate_after_path"`
}

type claimIdentityRecordSnapshot struct {
	StableID                     string `json:"stable_id"`
	Kind                         string `json:"kind"`
	RelationRole                 string `json:"relation_role"`
	NormalizedProposition        string `json:"normalized_proposition"`
	PropositionDigest            string `json:"proposition_digest"`
	TargetAddress                string `json:"target_address"`
	TargetAddressDigest          string `json:"target_address_digest"`
	PreservationOf               string `json:"preservation_of,omitempty"`
	BeforeSourcePath             string `json:"before_source_path,omitempty"`
	AfterSourcePath              string `json:"after_source_path,omitempty"`
	EvidenceBeforeRawDigest      string `json:"evidence_before_raw_digest,omitempty"`
	EvidenceAfterRawDigest       string `json:"evidence_after_raw_digest,omitempty"`
	EvidenceBeforeSemanticDigest string `json:"evidence_before_semantic_digest,omitempty"`
	EvidenceAfterSemanticDigest  string `json:"evidence_after_semantic_digest,omitempty"`
}

type persistenceObservation struct {
	SourcePair evolutionSourcePair           `json:"source_pair"`
	Records    []claimIdentityRecordSnapshot `json:"records"`
}

type persistenceMapping struct {
	BaselineIDs                     []string `json:"baseline_ids"`
	AlternateIDs                    []string `json:"alternate_ids"`
	RemovedIDs                      []string `json:"removed_ids,omitempty"`
	AddedIDs                        []string `json:"added_ids,omitempty"`
	StableIdentityPreserved         int      `json:"stable_identity_preserved"`
	StableIdentityTotal             int      `json:"stable_identity_total"`
	EvidenceOnlyChanges             int      `json:"evidence_only_changes"`
	EvidenceOnlyTotal               int      `json:"evidence_only_total"`
	RawEvidenceChanged              int      `json:"raw_evidence_changed"`
	RawEvidenceTotal                int      `json:"raw_evidence_total"`
	SemanticEvidencePreserved       int      `json:"semantic_evidence_preserved"`
	SemanticEvidenceTotal           int      `json:"semantic_evidence_total"`
	SemanticTargetPreserved         int      `json:"semantic_target_preserved"`
	SemanticTargetTotal             int      `json:"semantic_target_total"`
	ClaimRecreatedDueOnlyToRaw      int      `json:"claim_recreated_due_only_to_raw_digest"`
	ClaimRecreatedDueOnlyToRawTotal int      `json:"claim_recreated_due_only_to_raw_digest_total"`
	Decision                        string   `json:"decision"`
	Resolution                      string   `json:"resolution"`
	Stage                           string   `json:"stage"`
	Step                            string   `json:"step"`
	Reason                          string   `json:"reason"`
}

type evolutionReport struct {
	Schema                            string          `json:"schema"`
	Authority                         string          `json:"authority"`
	OldArtifactPath                   string          `json:"old_artifact_path"`
	OldArtifactBytes                  int             `json:"old_artifact_bytes"`
	OldArtifactDigest                 string          `json:"old_artifact_digest"`
	NewArtifactPath                   string          `json:"new_artifact_path"`
	NewArtifactBytes                  int             `json:"new_artifact_bytes"`
	NewArtifactDigest                 string          `json:"new_artifact_digest"`
	PersistenceManifestPath           string          `json:"persistence_manifest_path"`
	PersistenceManifestBytes          int             `json:"persistence_manifest_bytes"`
	PersistenceManifestDigest         string          `json:"persistence_manifest_digest"`
	DenominatorID                     string          `json:"denominator_id"`
	DenominatorUnchanged              bool            `json:"denominator_unchanged"`
	FixedClaimTotalBefore             int             `json:"fixed_claim_total_before"`
	FixedClaimTotalAfter              int             `json:"fixed_claim_total_after"`
	StableIdentityPreserved           int             `json:"stable_identity_preserved"`
	StableIdentityTotal               int             `json:"stable_identity_total"`
	PersistentClaimIdentity           int             `json:"persistent_claim_identity"`
	PersistentClaimIdentityTotal      int             `json:"persistent_claim_identity_total"`
	PropositionChanges                int             `json:"proposition_changes"`
	EvidenceOnlyChanges               int             `json:"evidence_only_changes"`
	EvidenceOnlyTotal                 int             `json:"evidence_only_total"`
	EvolutionRowsReconstructed        int             `json:"evolution_rows_independently_reconstructed"`
	EvolutionRowsTotal                int             `json:"evolution_rows_total"`
	EvolutionClaimRowsReconstructed   int             `json:"evolution_claim_rows_independently_reconstructed"`
	EvolutionClaimRowsTotal           int             `json:"evolution_claim_rows_total"`
	RawEvidenceChangedNonsemantic     int             `json:"raw_evidence_changed_on_nonsemantic"`
	RawEvidenceNonsemanticTotal       int             `json:"raw_evidence_nonsemantic_total"`
	SemanticTargetPreserved           int             `json:"semantic_target_preserved_on_nonsemantic"`
	SemanticTargetNonsemanticTotal    int             `json:"semantic_target_nonsemantic_total"`
	SemanticEvidencePreserved         int             `json:"semantic_evidence_preserved_on_nonsemantic"`
	SemanticEvidenceNonsemanticTotal  int             `json:"semantic_evidence_nonsemantic_total"`
	ClaimRecreatedDueOnlyToRaw        int             `json:"claim_recreated_due_only_to_raw_digest"`
	ClaimRecreatedDueOnlyToRawTotal   int             `json:"claim_recreated_due_only_to_raw_digest_total"`
	ExpectationConformanceRows        int             `json:"expectation_conformance_rows"`
	ExpectationConformanceRowsTotal   int             `json:"expectation_conformance_rows_total"`
	ExpectationConformanceClaims      int             `json:"expectation_conformance_claim_rows"`
	ExpectationConformanceClaimsTotal int             `json:"expectation_conformance_claim_rows_total"`
	V3ObservationPairsReconstructed   int             `json:"v3_observation_pairs_reconstructed"`
	V3ObservationPairsTotal           int             `json:"v3_observation_pairs_total"`
	V3ProducerConsumerExact           int             `json:"v3_producer_consumer_exact"`
	V3ProducerConsumerExactTotal      int             `json:"v3_producer_consumer_exact_total"`
	V3PersistenceSatisfied            int             `json:"v3_persistence_satisfied"`
	V3PersistenceSatisfiedTotal       int             `json:"v3_persistence_satisfied_total"`
	MigrationRemoved                  int             `json:"historical_migration_removed"`
	MigrationAdded                    int             `json:"historical_migration_added"`
	MigrationMappingRows              int             `json:"historical_migration_mapping_rows"`
	MigrationDecision                 string          `json:"historical_migration_decision"`
	MigrationResolution               string          `json:"historical_migration_resolution"`
	PersistenceDecision               string          `json:"persistence_decision"`
	PersistenceResolution             string          `json:"persistence_resolution"`
	ChangeKind                        string          `json:"change_kind"`
	Cases                             []evolutionCase `json:"cases"`
	Decision                          string          `json:"decision"`
	Resolution                        string          `json:"resolution"`
	Stage                             string          `json:"stage"`
	Step                              string          `json:"step"`
	Reason                            string          `json:"reason"`
}

func reconstructEvolution(oldPath, newPath, persistencePath string) (evolutionReport, error) {
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
	persistenceRaw, err := os.ReadFile(persistencePath)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("read persistence manifest: %w", err)
	}
	persistence, err := decodePersistenceManifest(persistenceRaw)
	if err != nil {
		return evolutionReport{}, fmt.Errorf("decode persistence manifest: %w", err)
	}
	meta, err := producer.ReadMetaContract()
	if err != nil {
		return evolutionReport{}, fmt.Errorf("read source-derived meta contract: %w", err)
	}
	if !persistenceRecipesMatch(meta.PersistenceRecipes, persistence.Cases) {
		return evolutionReport{}, fmt.Errorf("meta persistence recipes do not match fixed manifest")
	}
	report := evolutionReport{
		Schema: evolutionSchema, Authority: "SOURCE_DERIVED_SEMANTIC_CLAIM_CONTRACT",
		OldArtifactPath: oldPath, OldArtifactBytes: len(oldRaw), OldArtifactDigest: bytesDigest(oldRaw),
		NewArtifactPath: newPath, NewArtifactBytes: len(newRaw), NewArtifactDigest: bytesDigest(newRaw),
		PersistenceManifestPath: persistencePath, PersistenceManifestBytes: len(persistenceRaw), PersistenceManifestDigest: bytesDigest(persistenceRaw),
		DenominatorID:         newArtifact.DenominatorID,
		DenominatorUnchanged:  oldArtifact.DenominatorID == newArtifact.DenominatorID && oldArtifact.FixedClaimTotal == newArtifact.FixedClaimTotal,
		FixedClaimTotalBefore: oldArtifact.FixedClaimTotal, FixedClaimTotalAfter: newArtifact.FixedClaimTotal,
		EvolutionRowsTotal: len(newArtifact.Cases), ExpectationConformanceRowsTotal: 5, ExpectationConformanceClaimsTotal: 31,
		V3ObservationPairsTotal: 5, V3ProducerConsumerExactTotal: 5, V3PersistenceSatisfiedTotal: 5,
		PersistenceDecision: producer.DecisionFailClosed, PersistenceResolution: producer.ResolutionLower,
		Decision: producer.DecisionFailClosed, Resolution: producer.ResolutionLower,
		Stage: "claim-identity-evolution", Step: "reconstruct-source-observations", Reason: "CLAIM_IDENTITY_EVOLUTION_UNKNOWN",
	}
	oldByID := make(map[string]legacyExpectationRow, len(oldArtifact.Cases))
	for _, row := range oldArtifact.Cases {
		oldByID[row.ID] = row
	}
	newByID := make(map[string]evolutionExpectationRow, len(newArtifact.Cases))
	for _, row := range newArtifact.Cases {
		newByID[row.ID] = row
	}
	manifestByID := make(map[string]persistenceManifestCase, len(persistence.Cases))
	for _, row := range persistence.Cases {
		manifestByID[row.ID] = row
	}
	for _, definition := range producer.Denominator() {
		oldRow, oldOK := oldByID[definition.ID]
		newRow, newOK := newByID[definition.ID]
		input := producer.Input{CaseID: definition.ID, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath, SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ObservedCheckoutSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
		producerReceipt, producerErr := producer.ProduceFiles(input)
		producerObservation, producerObservationErr := producer.ClaimIdentityObservationFromFiles(input)
		producerRecords := producerObservation.Records
		consumerInput := consumer.Input{CaseID: input.CaseID, BeforePath: input.BeforePath, AfterPath: input.AfterPath, SubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA}
		consumerRecords, sourcePair, consumerErr := consumer.ClaimIdentityRecordsFromFiles(consumerInput)
		oldObservedIDs, oldRecords := legacyClaimIdentity(producerReceipt, input)
		consumerOldRecords, _, consumerOldErr := consumer.LegacyClaimIdentityRecordsFromFiles(consumerInput)
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
		row.ProducerConsumerExact = producerObservationErr == nil && producerErr == nil && consumerErr == nil && producerRecordsEqual(producerRecords, consumerRecords)
		row.NewExpectationProducerExact = expectedProducerRecordsEqual(newRow.ExpectedClaims, producerRecords)
		row.NewExpectationConsumerExact = expectedConsumerRecordsEqual(newRow.ExpectedClaims, consumerRecords)
		row.ObservedSourcePair = evolutionSourcePair{BeforePath: sourcePair.BeforePath, AfterPath: sourcePair.AfterPath, BeforeRawDigest: sourcePair.BeforeRawDigest, AfterRawDigest: sourcePair.AfterRawDigest, BeforeSemanticDigest: sourcePair.BeforeSemanticDigest, AfterSemanticDigest: sourcePair.AfterSemanticDigest}
		row.RemovedIDs, row.AddedIDs = setDiff(row.OldExpectedIDs, row.NewExpectedIDs)
		row.PropositionTargetChanges = legacyPropositionTargetChanges(oldRecords, newRow.ExpectedClaims)
		row.ConsumerPropositionTargetChanges = consumerPropositionTargetChanges(consumerOldRecords, consumerRecords)
		if oldOK && newOK && row.OldArtifactExact && row.NewExpectationProducerExact && row.NewExpectationConsumerExact && row.ProducerConsumerExact && propositionChangesEqual(row.PropositionTargetChanges, row.ConsumerPropositionTargetChanges) {
			report.EvolutionRowsReconstructed++
			report.ExpectationConformanceRows++
			report.ExpectationConformanceClaims += newRow.ExpectedClaimTotal
			report.MigrationRemoved += len(row.RemovedIDs)
			report.MigrationAdded += len(row.AddedIDs)
			report.MigrationMappingRows += len(row.PropositionTargetChanges)
			report.PropositionChanges += len(row.PropositionTargetChanges)
		}
		if alternate, ok := manifestByID[definition.ID]; ok {
			row.ExpectedPersistenceClaimTotal = alternate.ExpectedClaimTotal
			alternateInput := producer.Input{CaseID: definition.ID, BeforePath: alternate.AlternateBeforePath, AfterPath: alternate.AlternateAfterPath, SubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA}
			producerAlternate, producerAlternateErr := producer.ClaimIdentityObservationFromFiles(alternateInput)
			consumerAlternate, consumerAlternatePair, consumerAlternateErr := consumer.ClaimIdentityRecordsFromFiles(consumer.Input{CaseID: definition.ID, BeforePath: alternate.AlternateBeforePath, AfterPath: alternate.AlternateAfterPath, SubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA})
			producerBaselinePair := producerSourcePair(producerObservation)
			producerAlternatePair := producerSourcePair(producerAlternate)
			consumerBaselinePair := evolutionSourcePair{BeforePath: sourcePair.BeforePath, AfterPath: sourcePair.AfterPath, BeforeRawDigest: sourcePair.BeforeRawDigest, AfterRawDigest: sourcePair.AfterRawDigest, BeforeSemanticDigest: sourcePair.BeforeSemanticDigest, AfterSemanticDigest: sourcePair.AfterSemanticDigest}
			consumerAlternateSourcePair := evolutionSourcePair{BeforePath: consumerAlternatePair.BeforePath, AfterPath: consumerAlternatePair.AfterPath, BeforeRawDigest: consumerAlternatePair.BeforeRawDigest, AfterRawDigest: consumerAlternatePair.AfterRawDigest, BeforeSemanticDigest: consumerAlternatePair.BeforeSemanticDigest, AfterSemanticDigest: consumerAlternatePair.AfterSemanticDigest}
			row.BaselineObservation = persistenceObservation{SourcePair: producerBaselinePair, Records: producerRecordSnapshots(producerRecords)}
			row.AlternateObservation = persistenceObservation{SourcePair: producerAlternatePair, Records: producerRecordSnapshots(producerAlternate.Records)}
			row.ProducerBaselineObservation = row.BaselineObservation
			row.ProducerAlternateObservation = row.AlternateObservation
			row.ConsumerBaselineObservation = persistenceObservation{SourcePair: consumerBaselinePair, Records: consumerRecordSnapshots(consumerRecords)}
			row.ConsumerAlternateObservation = persistenceObservation{SourcePair: consumerAlternateSourcePair, Records: consumerRecordSnapshots(consumerAlternate)}
			producerPersistence := producer.CompareClaimIdentityRecords(producerRecords, producerAlternate.Records)
			consumerPersistence := consumer.CompareClaimIdentityRecords(consumerRecords, consumerAlternate)
			row.ProducerPersistence = producerMapping(producerPersistence)
			row.ConsumerPersistence = consumerMapping(consumerPersistence)
			row.ReconstructionExact = producerObservationErr == nil && producerAlternateErr == nil && consumerErr == nil && consumerAlternateErr == nil && producerRecordsEqual(producerRecords, consumerRecords) && producerRecordsEqual(producerAlternate.Records, consumerAlternate) && alternate.ExpectedClaimTotal == len(producerRecords) && alternate.ExpectedClaimTotal == len(producerAlternate.Records)
			mappingExact := persistenceMappingsEqual(row.ProducerPersistence, row.ConsumerPersistence)
			row.PersistenceSatisfied = row.ReconstructionExact && mappingExact && persistenceMappingSatisfies(row.ProducerPersistence, alternate.ExpectedClaimTotal) && persistenceMappingSatisfies(row.ConsumerPersistence, alternate.ExpectedClaimTotal)
			row.PersistenceExact = row.PersistenceSatisfied
			if row.ReconstructionExact {
				report.V3ObservationPairsReconstructed++
			}
			if row.ReconstructionExact && mappingExact {
				report.V3ProducerConsumerExact++
			}
			if row.PersistenceSatisfied {
				report.V3PersistenceSatisfied++
			}
			row.StableIdentityPreserved, row.StableIdentityTotal = producerPersistence.StableIdentityPreserved, producerPersistence.StableIdentityTotal
			row.EvidenceOnlyChanges, row.EvidenceOnlyTotal = producerPersistence.EvidenceOnlyChanges, producerPersistence.EvidenceOnlyTotal
			row.ClaimRecreatedDueOnlyToRaw, row.ClaimRecreatedDueOnlyToRawTotal = producerPersistence.ClaimRecreatedDueOnlyToRaw, producerPersistence.ClaimRecreatedDueOnlyToRawTotal
			report.StableIdentityPreserved += row.StableIdentityPreserved
			report.StableIdentityTotal += row.StableIdentityTotal
			report.PersistentClaimIdentity += row.StableIdentityPreserved
			report.PersistentClaimIdentityTotal += row.StableIdentityTotal
			report.EvidenceOnlyChanges += row.EvidenceOnlyChanges
			report.EvidenceOnlyTotal += row.EvidenceOnlyTotal
			report.ClaimRecreatedDueOnlyToRaw += row.ClaimRecreatedDueOnlyToRaw
			report.ClaimRecreatedDueOnlyToRawTotal += row.ClaimRecreatedDueOnlyToRawTotal
			report.RawEvidenceChangedNonsemantic += producerPersistence.RawEvidenceChanged
			report.RawEvidenceNonsemanticTotal += producerPersistence.RawEvidenceTotal
			report.SemanticEvidencePreserved += producerPersistence.SemanticEvidencePreserved
			report.SemanticEvidenceNonsemanticTotal += producerPersistence.SemanticEvidenceTotal
			report.SemanticTargetPreserved += producerPersistence.SemanticTargetPreserved
			report.SemanticTargetNonsemanticTotal += producerPersistence.SemanticTargetTotal
		}
		if row.PersistenceSatisfied && row.OldArtifactExact && row.NewExpectationProducerExact && row.NewExpectationConsumerExact && row.ProducerConsumerExact {
			row.Decision, row.Resolution, row.Stage, row.Step, row.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "claim-identity-evolution", "compare-source-derived-observations", "CLAIM_IDENTITY_EVOLUTION_EXACT"
		} else if !row.ReconstructionExact {
			row.Stage, row.Step, row.Reason = "claim-identity-persistence", "reconstruct-v3-observations", "INDEPENDENT_RECONSTRUCTION_MISMATCH"
		} else if !row.PersistenceSatisfied {
			row.Stage, row.Step, row.Reason = "claim-identity-persistence", "compare-v3-observations", persistenceFailureReason(row.ProducerPersistence, row.ConsumerPersistence, persistenceExpectedTotal(row))
		}
		report.Cases = append(report.Cases, row)
	}
	report.EvolutionClaimRowsReconstructed = report.ExpectationConformanceClaims
	report.EvolutionClaimRowsTotal = report.ExpectationConformanceClaimsTotal
	if report.EvolutionRowsReconstructed == report.EvolutionRowsTotal {
		report.MigrationDecision, report.MigrationResolution = producer.DecisionFixedPoint, producer.ResolutionExact
	}
	if report.V3PersistenceSatisfied == report.V3PersistenceSatisfiedTotal {
		report.PersistenceDecision, report.PersistenceResolution = producer.DecisionFixedPoint, producer.ResolutionExact
	}
	if report.V3PersistenceSatisfied != report.V3PersistenceSatisfiedTotal {
		for _, row := range report.Cases {
			if !row.PersistenceSatisfied {
				report.Stage, report.Step, report.Reason = row.Stage, row.Step, row.Reason
				break
			}
		}
	}
	if report.EvolutionRowsReconstructed == 5 && report.ExpectationConformanceRows == 5 && report.ExpectationConformanceClaims == 31 && report.V3ObservationPairsReconstructed == 5 && report.V3ProducerConsumerExact == 5 && report.V3PersistenceSatisfied == 5 && report.PersistentClaimIdentity == 31 && report.PersistentClaimIdentityTotal == 31 && report.EvidenceOnlyChanges == 31 && report.RawEvidenceChangedNonsemantic == 31 && report.RawEvidenceNonsemanticTotal == 31 && report.SemanticEvidencePreserved == 31 && report.SemanticEvidenceNonsemanticTotal == 31 && report.SemanticTargetPreserved == 31 && report.SemanticTargetNonsemanticTotal == 31 && report.ClaimRecreatedDueOnlyToRaw == 0 && report.ClaimRecreatedDueOnlyToRawTotal == 31 && report.DenominatorUnchanged && report.NewArtifactDigest == bytesDigest(newRaw) {
		report.Decision, report.Resolution, report.Reason = producer.DecisionFixedPoint, producer.ResolutionExact, "CLAIM_IDENTITY_V3_PERSISTENCE_EXACT"
		report.ChangeKind = "HISTORICAL_SCHEMA_MIGRATION"
	} else {
		report.ChangeKind = "HISTORICAL_SCHEMA_MIGRATION_WITH_PERSISTENCE_FAIL_CLOSED"
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

func decodePersistenceManifest(raw []byte) (persistenceManifest, error) {
	var value persistenceManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("trailing persistence manifest data")
	}
	if value.Schema != "gooo/semantic-delta-claim-identity-persistence/v1" || value.DenominatorID != producer.DenominatorID || value.FixedCaseTotal != 5 || value.FixedClaimTotal != 31 || !fixedPersistenceInventory(value.Cases) {
		return value, fmt.Errorf("persistence manifest contract mismatch")
	}
	return value, nil
}

func fixedPersistenceInventory(cases []persistenceManifestCase) bool {
	if len(cases) != 5 {
		return false
	}
	want := map[string]bool{"equivalent": true, "semantic-change": true, "value-program-change": true, "indeterminate": true, "ambiguous-match": true}
	seen := map[string]bool{}
	for _, definition := range producer.Denominator() {
		var row persistenceManifestCase
		for _, candidate := range cases {
			if candidate.ID == definition.ID {
				if seen[candidate.ID] {
					return false
				}
				row, seen[candidate.ID] = candidate, true
				break
			}
		}
		if !want[row.ID] || row.ExpectedClaimTotal <= 0 || row.ExpectedClaimTotal != expectedClaimTotalForCase(row.ID) || row.BaselineBeforePath != definition.BeforePath || row.BaselineAfterPath != definition.AfterPath || row.AlternateBeforePath == "" || row.AlternateAfterPath == "" {
			return false
		}
	}
	return len(seen) == len(want)
}

func persistenceRecipesMatch(recipes []producer.PersistenceRecipe, manifest []persistenceManifestCase) bool {
	if len(recipes) != len(manifest) {
		return false
	}
	byID := make(map[string]persistenceManifestCase, len(manifest))
	for _, row := range manifest {
		byID[row.ID] = row
	}
	for _, recipe := range recipes {
		row, ok := byID[recipe.ID]
		if !ok || recipe.BaselineBeforePath != row.BaselineBeforePath || recipe.BaselineAfterPath != row.BaselineAfterPath || recipe.AlternateBeforePath != row.AlternateBeforePath || recipe.AlternateAfterPath != row.AlternateAfterPath {
			return false
		}
	}
	return true
}

func expectedClaimTotalForCase(id string) int {
	switch id {
	case "indeterminate":
		return 3
	default:
		return 7
	}
}

func producerRecordSnapshots(records []producer.ClaimIdentityRecord) []claimIdentityRecordSnapshot {
	result := make([]claimIdentityRecordSnapshot, 0, len(records))
	for _, record := range records {
		result = append(result, claimIdentityRecordSnapshot{StableID: record.StableID, Kind: record.Kind, RelationRole: record.RelationRole, NormalizedProposition: record.NormalizedProposition, PropositionDigest: record.PropositionDigest, TargetAddress: record.TargetAddress, TargetAddressDigest: record.TargetAddressDigest, PreservationOf: record.PreservationOf, BeforeSourcePath: record.BeforeSourcePath, AfterSourcePath: record.AfterSourcePath, EvidenceBeforeRawDigest: record.EvidenceBeforeRawDigest, EvidenceAfterRawDigest: record.EvidenceAfterRawDigest, EvidenceBeforeSemanticDigest: record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: record.EvidenceAfterSemanticDigest})
	}
	return result
}

func producerSourcePair(observation producer.ClaimIdentitySourceObservation) evolutionSourcePair {
	return evolutionSourcePair{BeforePath: observation.BeforePath, AfterPath: observation.AfterPath, BeforeRawDigest: observation.BeforeRawDigest, AfterRawDigest: observation.AfterRawDigest, BeforeSemanticDigest: observation.BeforeSemanticDigest, AfterSemanticDigest: observation.AfterSemanticDigest}
}

func consumerRecordSnapshots(records []consumer.ClaimIdentityRecord) []claimIdentityRecordSnapshot {
	result := make([]claimIdentityRecordSnapshot, 0, len(records))
	for _, record := range records {
		result = append(result, claimIdentityRecordSnapshot{StableID: record.StableID, Kind: record.Kind, RelationRole: record.RelationRole, NormalizedProposition: record.NormalizedProposition, PropositionDigest: record.PropositionDigest, TargetAddress: record.TargetAddress, TargetAddressDigest: record.TargetAddressDigest, PreservationOf: record.PreservationOf, BeforeSourcePath: record.BeforeSourcePath, AfterSourcePath: record.AfterSourcePath, EvidenceBeforeRawDigest: record.EvidenceBeforeRawDigest, EvidenceAfterRawDigest: record.EvidenceAfterRawDigest, EvidenceBeforeSemanticDigest: record.EvidenceBeforeSemanticDigest, EvidenceAfterSemanticDigest: record.EvidenceAfterSemanticDigest})
	}
	return result
}

func producerMapping(mapping producer.ClaimIdentityPairComparison) persistenceMapping {
	return persistenceMapping{BaselineIDs: mapping.BaselineIDs, AlternateIDs: mapping.AlternateIDs, RemovedIDs: mapping.RemovedIDs, AddedIDs: mapping.AddedIDs, StableIdentityPreserved: mapping.StableIdentityPreserved, StableIdentityTotal: mapping.StableIdentityTotal, EvidenceOnlyChanges: mapping.EvidenceOnlyChanges, EvidenceOnlyTotal: mapping.EvidenceOnlyTotal, RawEvidenceChanged: mapping.RawEvidenceChanged, RawEvidenceTotal: mapping.RawEvidenceTotal, SemanticEvidencePreserved: mapping.SemanticEvidencePreserved, SemanticEvidenceTotal: mapping.SemanticEvidenceTotal, SemanticTargetPreserved: mapping.SemanticTargetPreserved, SemanticTargetTotal: mapping.SemanticTargetTotal, ClaimRecreatedDueOnlyToRaw: mapping.ClaimRecreatedDueOnlyToRaw, ClaimRecreatedDueOnlyToRawTotal: mapping.ClaimRecreatedDueOnlyToRawTotal, Decision: mapping.Decision, Resolution: mapping.Resolution, Stage: mapping.Stage, Step: mapping.Step, Reason: mapping.Reason}
}

func consumerMapping(mapping consumer.ClaimIdentityPairComparison) persistenceMapping {
	return persistenceMapping{BaselineIDs: mapping.BaselineIDs, AlternateIDs: mapping.AlternateIDs, RemovedIDs: mapping.RemovedIDs, AddedIDs: mapping.AddedIDs, StableIdentityPreserved: mapping.StableIdentityPreserved, StableIdentityTotal: mapping.StableIdentityTotal, EvidenceOnlyChanges: mapping.EvidenceOnlyChanges, EvidenceOnlyTotal: mapping.EvidenceOnlyTotal, RawEvidenceChanged: mapping.RawEvidenceChanged, RawEvidenceTotal: mapping.RawEvidenceTotal, SemanticEvidencePreserved: mapping.SemanticEvidencePreserved, SemanticEvidenceTotal: mapping.SemanticEvidenceTotal, SemanticTargetPreserved: mapping.SemanticTargetPreserved, SemanticTargetTotal: mapping.SemanticTargetTotal, ClaimRecreatedDueOnlyToRaw: mapping.ClaimRecreatedDueOnlyToRaw, ClaimRecreatedDueOnlyToRawTotal: mapping.ClaimRecreatedDueOnlyToRawTotal, Decision: mapping.Decision, Resolution: mapping.Resolution, Stage: mapping.Stage, Step: mapping.Step, Reason: mapping.Reason}
}

func persistenceExpectedTotal(row evolutionCase) int {
	if row.ExpectedPersistenceClaimTotal > 0 {
		return row.ExpectedPersistenceClaimTotal
	}
	return 0
}

func persistenceMappingSatisfies(mapping persistenceMapping, expectedTotal int) bool {
	return expectedTotal > 0 && uniqueNonemptyIDs(mapping.BaselineIDs) && uniqueNonemptyIDs(mapping.AlternateIDs) && len(mapping.BaselineIDs) == expectedTotal && len(mapping.AlternateIDs) == expectedTotal && len(mapping.RemovedIDs) == 0 && len(mapping.AddedIDs) == 0 && mapping.StableIdentityPreserved == expectedTotal && mapping.StableIdentityTotal == expectedTotal && mapping.RawEvidenceChanged == expectedTotal && mapping.RawEvidenceTotal == expectedTotal && mapping.SemanticEvidencePreserved == expectedTotal && mapping.SemanticEvidenceTotal == expectedTotal && mapping.SemanticTargetPreserved == expectedTotal && mapping.SemanticTargetTotal == expectedTotal && mapping.EvidenceOnlyChanges == expectedTotal && mapping.EvidenceOnlyTotal == expectedTotal && mapping.ClaimRecreatedDueOnlyToRaw == 0 && mapping.ClaimRecreatedDueOnlyToRawTotal == expectedTotal && mapping.Decision == producer.DecisionFixedPoint && mapping.Resolution == producer.ResolutionExact
}

func uniqueNonemptyIDs(ids []string) bool {
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

func persistenceFailureReason(producerMapping, consumerMapping persistenceMapping, expectedTotal int) string {
	for _, mapping := range []persistenceMapping{producerMapping, consumerMapping} {
		if !uniqueNonemptyIDs(mapping.BaselineIDs) || !uniqueNonemptyIDs(mapping.AlternateIDs) {
			return "DUPLICATE_STABLE_CLAIM_ID"
		}
		rawOnlySet := mapping.ClaimRecreatedDueOnlyToRaw > 0 && len(mapping.RemovedIDs) == mapping.ClaimRecreatedDueOnlyToRaw && len(mapping.AddedIDs) == mapping.ClaimRecreatedDueOnlyToRaw
		if rawOnlySet {
			return "CLAIM_RECREATED_DUE_ONLY_TO_RAW_DIGEST"
		}
		if len(mapping.RemovedIDs) != 0 || len(mapping.AddedIDs) != 0 || mapping.StableIdentityPreserved != expectedTotal || mapping.StableIdentityTotal != expectedTotal || len(mapping.BaselineIDs) != expectedTotal || len(mapping.AlternateIDs) != expectedTotal {
			return "CLAIM_SET_CHANGED"
		}
		if mapping.SemanticTargetPreserved != expectedTotal || mapping.SemanticTargetTotal != expectedTotal {
			return "SEMANTIC_TARGET_CHANGED"
		}
		if mapping.RawEvidenceChanged != expectedTotal || mapping.RawEvidenceTotal != expectedTotal {
			return "RAW_EVIDENCE_UNCHANGED"
		}
		if mapping.SemanticEvidencePreserved != expectedTotal || mapping.SemanticEvidenceTotal != expectedTotal {
			return "SEMANTIC_EVIDENCE_CHANGED"
		}
		if mapping.EvidenceOnlyChanges != expectedTotal || mapping.EvidenceOnlyTotal != expectedTotal {
			return "EVIDENCE_ONLY_INTERVENTION_UNPROVEN"
		}
		if mapping.ClaimRecreatedDueOnlyToRaw != 0 || mapping.ClaimRecreatedDueOnlyToRawTotal != expectedTotal {
			return "CLAIM_RECREATED_DUE_ONLY_TO_RAW_DIGEST"
		}
		if mapping.Decision != producer.DecisionFixedPoint || mapping.Resolution != producer.ResolutionExact {
			if mapping.Reason != "" && mapping.Reason != "V3_CLAIM_IDENTITY_PERSISTED_ACROSS_RAW_INTERVENTION" {
				return mapping.Reason
			}
			return "PERSISTENCE_DECISION_NOT_FIXED_POINT"
		}
	}
	if !persistenceMappingsEqual(producerMapping, consumerMapping) {
		return "INDEPENDENT_PERSISTENCE_MAPPING_MISMATCH"
	}
	return "CLAIM_IDENTITY_PERSISTENCE_UNKNOWN"
}

func persistenceMappingsEqual(left, right persistenceMapping) bool {
	return bytes.Equal(mustJSON(left), mustJSON(right))
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
