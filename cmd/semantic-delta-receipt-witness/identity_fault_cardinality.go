package main

import (
	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

const identityFaultCardinalityDenominator = 3

var identityFaultCardinalityCaseIDs = []string{
	"six-unique-slots",
	"eight-unique-slots",
	"seven-slots-one-duplicate",
}

type identityFaultCardinalityReport struct {
	Schema                 string                               `json:"schema"`
	DenominatorID          string                               `json:"denominator_id"`
	ExpectedCaseIDs        []string                             `json:"expected_case_ids"`
	ObservedCaseIDs        []string                             `json:"observed_case_ids"`
	CardinalityNumerator   int                                  `json:"cardinality_counterexample_numerator"`
	CardinalityDenominator int                                  `json:"cardinality_counterexample_denominator"`
	Cases                  []identityFaultCardinalityCaseResult `json:"cases"`
	Decision               string                               `json:"decision"`
	Resolution             string                               `json:"resolution"`
	Stage                  string                               `json:"stage"`
	Step                   string                               `json:"step"`
	Reason                 string                               `json:"reason"`
}

type identityFaultCardinalityCaseResult struct {
	CaseID                string `json:"case_id"`
	ExpectedUnique        int    `json:"expected_unique"`
	ExpectedTotal         int    `json:"expected_total"`
	ExpectedReason        string `json:"expected_reason"`
	ProducerUnique        int    `json:"producer_unique"`
	ProducerTotal         int    `json:"producer_total"`
	ProducerDenominator   int    `json:"producer_denominator"`
	ProducerDecision      string `json:"producer_decision"`
	ProducerResolution    string `json:"producer_resolution"`
	ProducerStage         string `json:"producer_stage"`
	ProducerStep          string `json:"producer_step"`
	ProducerReason        string `json:"producer_reason"`
	ConsumerUnique        int    `json:"consumer_unique"`
	ConsumerTotal         int    `json:"consumer_total"`
	ConsumerDenominator   int    `json:"consumer_denominator"`
	ConsumerDecision      string `json:"consumer_decision"`
	ConsumerResolution    string `json:"consumer_resolution"`
	ConsumerStage         string `json:"consumer_stage"`
	ConsumerStep          string `json:"consumer_step"`
	ConsumerReason        string `json:"consumer_reason"`
	ProducerConsumerExact bool   `json:"producer_consumer_exact"`
	Passed                bool   `json:"passed"`
}

func runIdentityFaultCardinalityProbe(options options) identityFaultCardinalityReport {
	producerReceipt := producer.IdentityFaultReceiptFromFiles(producer.IdentityFaultInput{
		Baseline:     producer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceBefore, AfterPath: options.persistenceAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA},
		Alternate:    producer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceAlternateBefore, AfterPath: options.persistenceAlternateAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA},
		ArtifactPath: options.identityFault,
	}).Receipt
	consumerReceipt := consumer.IdentityFaultReceiptFromFiles(consumer.IdentityFaultInput{
		Baseline:     consumer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceBefore, AfterPath: options.persistenceAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA},
		Alternate:    consumer.Input{CaseID: "persistence-probe", BeforePath: options.persistenceAlternateBefore, AfterPath: options.persistenceAlternateAfter, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA},
		ArtifactPath: options.identityFault,
	}).Receipt

	report := identityFaultCardinalityReport{
		Schema:          "gooo/semantic-delta-claim-identity-fault-cardinality/v1",
		DenominatorID:   "identity-fault-semantic-slot-cardinality/v1",
		ExpectedCaseIDs: append([]string(nil), identityFaultCardinalityCaseIDs...),
		Decision:        producer.DecisionFailClosed, Resolution: producer.ResolutionLower,
		Stage: "identity-fault", Step: "cardinality-counterexamples", Reason: "IDENTITY_SEMANTIC_SLOT_CARDINALITY_UNKNOWN",
	}
	for _, definition := range identityFaultCardinalityDefinitions() {
		producerOriginal, producerFaulted, producerRows := producerIdentityFaultCardinalityRecords(producerReceipt, definition.ID)
		consumerOriginal, consumerFaulted, consumerRows := consumerIdentityFaultCardinalityRecords(consumerReceipt, definition.ID)
		producerGraph, producerReason := producer.ValidateIdentityFaultGraph(producerOriginal, producerFaulted, producerReceipt.Alternate.SourcePair, producerReceipt.Artifact.Rule, producerRows)
		consumerGraph, consumerReason := consumer.ValidateIdentityFaultGraph(consumerOriginal, consumerFaulted, consumerReceipt.Alternate.SourcePair, consumerReceipt.Artifact.Rule, consumerRows)
		passed := producerReason == definition.Reason && consumerReason == definition.Reason && producerGraph.Decision == producer.DecisionFailClosed && consumerGraph.Decision == "FAIL_CLOSED" && producerGraph.Resolution == producer.ResolutionLower && consumerGraph.Resolution == "LOWER_RESOLUTION" && producerGraph.Stage == "identity-fault" && consumerGraph.Stage == "identity-fault" && producerGraph.Step == "rekey-graph" && consumerGraph.Step == "rekey-graph" && producerGraph.Reason == definition.Reason && consumerGraph.Reason == definition.Reason && producerGraph.SemanticSlotUnique == definition.Unique && producerGraph.SemanticSlotTotal == definition.Total && consumerGraph.SemanticSlotUnique == definition.Unique && consumerGraph.SemanticSlotTotal == definition.Total && producerGraph.SemanticSlotDenominator == 7 && consumerGraph.SemanticSlotDenominator == 7
		if passed {
			report.CardinalityNumerator++
		}
		report.Cases = append(report.Cases, identityFaultCardinalityCaseResult{
			CaseID: definition.ID, ExpectedUnique: definition.Unique, ExpectedTotal: definition.Total, ExpectedReason: definition.Reason,
			ProducerUnique: producerGraph.SemanticSlotUnique, ProducerTotal: producerGraph.SemanticSlotTotal, ProducerDenominator: producerGraph.SemanticSlotDenominator, ProducerDecision: producerGraph.Decision, ProducerResolution: producerGraph.Resolution, ProducerStage: producerGraph.Stage, ProducerStep: producerGraph.Step, ProducerReason: producerGraph.Reason,
			ConsumerUnique: consumerGraph.SemanticSlotUnique, ConsumerTotal: consumerGraph.SemanticSlotTotal, ConsumerDenominator: consumerGraph.SemanticSlotDenominator, ConsumerDecision: consumerGraph.Decision, ConsumerResolution: consumerGraph.Resolution, ConsumerStage: consumerGraph.Stage, ConsumerStep: consumerGraph.Step, ConsumerReason: consumerGraph.Reason,
			ProducerConsumerExact: producerGraph.SemanticSlotUnique == consumerGraph.SemanticSlotUnique && producerGraph.SemanticSlotTotal == consumerGraph.SemanticSlotTotal && producerReason == consumerReason,
			Passed:                passed,
		})
	}
	report.ObservedCaseIDs = make([]string, 0, len(report.Cases))
	for _, result := range report.Cases {
		report.ObservedCaseIDs = append(report.ObservedCaseIDs, result.CaseID)
	}
	report.CardinalityDenominator = identityFaultCardinalityDenominator
	if report.CardinalityNumerator == identityFaultCardinalityDenominator && stringSliceEqual(report.ExpectedCaseIDs, report.ObservedCaseIDs) {
		report.Decision, report.Resolution, report.Reason = "PASS", "EXACT", "IDENTITY_SEMANTIC_SLOT_CARDINALITY_COUNTEREXAMPLES_EXACT"
	}
	return report
}

type identityFaultCardinalityDefinition struct {
	ID     string
	Unique int
	Total  int
	Reason string
}

func identityFaultCardinalityDefinitions() []identityFaultCardinalityDefinition {
	return []identityFaultCardinalityDefinition{
		{ID: "six-unique-slots", Unique: 6, Total: 6, Reason: "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"},
		{ID: "eight-unique-slots", Unique: 8, Total: 8, Reason: "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"},
		{ID: "seven-slots-one-duplicate", Unique: 6, Total: 7, Reason: "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS"},
	}
}

func producerIdentityFaultCardinalityRecords(receipt producer.IdentityFaultReceipt, caseID string) ([]producer.ClaimIdentityRecord, []producer.ClaimIdentityRecord, []producer.IdentityFaultMappingRow) {
	original := append([]producer.ClaimIdentityRecord(nil), receipt.Alternate.Records...)
	faulted := append([]producer.ClaimIdentityRecord(nil), receipt.FaultedAlternate.Records...)
	rows := append([]producer.IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	if len(original) < 2 || len(faulted) < 2 {
		return original, faulted, rows
	}
	switch caseID {
	case "six-unique-slots":
		oldID := original[len(original)-1].StableID
		newID := producerIdentityFaultMappedNewID(rows, oldID)
		original = producerFilterRecords(original, oldID)
		faulted = producerFilterRecords(faulted, newID)
		rows = producerFilterRows(rows, oldID)
	case "eight-unique-slots":
		originalExtra := original[0]
		originalExtra.StableID += "/cardinality-extra"
		originalExtra.Kind += "-cardinality-extra"
		originalExtra.PreservationOf = ""
		faultedExtra := faulted[0]
		faultedExtra.StableID += "/cardinality-extra"
		faultedExtra.Kind = originalExtra.Kind
		faultedExtra.PreservationOf = ""
		original = append(original, originalExtra)
		faulted = append(faulted, faultedExtra)
		rows = append(rows, producer.IdentityFaultMappingRow{OldStableID: originalExtra.StableID, NewStableID: faultedExtra.StableID, Ordinal: 7})
	case "seven-slots-one-duplicate":
		producerCopySemanticSlot(&original[1], original[0])
		producerCopySemanticSlot(&faulted[1], faulted[0])
	}
	return original, faulted, rows
}

func consumerIdentityFaultCardinalityRecords(receipt consumer.IdentityFaultReceipt, caseID string) ([]consumer.ClaimIdentityRecord, []consumer.ClaimIdentityRecord, []consumer.IdentityFaultMappingRow) {
	original := append([]consumer.ClaimIdentityRecord(nil), receipt.Alternate.Records...)
	faulted := append([]consumer.ClaimIdentityRecord(nil), receipt.FaultedAlternate.Records...)
	rows := append([]consumer.IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	if len(original) < 2 || len(faulted) < 2 {
		return original, faulted, rows
	}
	switch caseID {
	case "six-unique-slots":
		oldID := original[len(original)-1].StableID
		newID := consumerIdentityFaultMappedNewID(rows, oldID)
		original = consumerFilterRecords(original, oldID)
		faulted = consumerFilterRecords(faulted, newID)
		rows = consumerFilterRows(rows, oldID)
	case "eight-unique-slots":
		originalExtra := original[0]
		originalExtra.StableID += "/cardinality-extra"
		originalExtra.Kind += "-cardinality-extra"
		originalExtra.PreservationOf = ""
		faultedExtra := faulted[0]
		faultedExtra.StableID += "/cardinality-extra"
		faultedExtra.Kind = originalExtra.Kind
		faultedExtra.PreservationOf = ""
		original = append(original, originalExtra)
		faulted = append(faulted, faultedExtra)
		rows = append(rows, consumer.IdentityFaultMappingRow{OldStableID: originalExtra.StableID, NewStableID: faultedExtra.StableID, Ordinal: 7})
	case "seven-slots-one-duplicate":
		consumerCopySemanticSlot(&original[1], original[0])
		consumerCopySemanticSlot(&faulted[1], faulted[0])
	}
	return original, faulted, rows
}

func producerIdentityFaultMappedNewID(rows []producer.IdentityFaultMappingRow, oldID string) string {
	for _, row := range rows {
		if row.OldStableID == oldID {
			return row.NewStableID
		}
	}
	return ""
}

func consumerIdentityFaultMappedNewID(rows []consumer.IdentityFaultMappingRow, oldID string) string {
	for _, row := range rows {
		if row.OldStableID == oldID {
			return row.NewStableID
		}
	}
	return ""
}

func producerFilterRecords(records []producer.ClaimIdentityRecord, stableID string) []producer.ClaimIdentityRecord {
	filtered := make([]producer.ClaimIdentityRecord, 0, len(records)-1)
	for _, record := range records {
		if record.StableID != stableID {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func consumerFilterRecords(records []consumer.ClaimIdentityRecord, stableID string) []consumer.ClaimIdentityRecord {
	filtered := make([]consumer.ClaimIdentityRecord, 0, len(records)-1)
	for _, record := range records {
		if record.StableID != stableID {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func producerFilterRows(rows []producer.IdentityFaultMappingRow, oldID string) []producer.IdentityFaultMappingRow {
	filtered := make([]producer.IdentityFaultMappingRow, 0, len(rows)-1)
	for _, row := range rows {
		if row.OldStableID != oldID {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func consumerFilterRows(rows []consumer.IdentityFaultMappingRow, oldID string) []consumer.IdentityFaultMappingRow {
	filtered := make([]consumer.IdentityFaultMappingRow, 0, len(rows)-1)
	for _, row := range rows {
		if row.OldStableID != oldID {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func producerCopySemanticSlot(target *producer.ClaimIdentityRecord, source producer.ClaimIdentityRecord) {
	target.Kind, target.RelationRole = source.Kind, source.RelationRole
	target.NormalizedProposition, target.PropositionDigest = source.NormalizedProposition, source.PropositionDigest
	target.TargetAddress, target.TargetAddressDigest = source.TargetAddress, source.TargetAddressDigest
}

func consumerCopySemanticSlot(target *consumer.ClaimIdentityRecord, source consumer.ClaimIdentityRecord) {
	target.Kind, target.RelationRole = source.Kind, source.RelationRole
	target.NormalizedProposition, target.PropositionDigest = source.NormalizedProposition, source.PropositionDigest
	target.TargetAddress, target.TargetAddressDigest = source.TargetAddress, source.TargetAddressDigest
}

func stringSliceEqual(left, right []string) bool {
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
