package semanticdeltareceipt

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestProducerIdentityFaultUsesRawFixtureAndRejectsClosedGraphTampering(t *testing.T) {
	receipt := producerIdentityFaultFixtureReceipt(t)
	if !receipt.FaultGraphClosed || !receipt.Graph.Bijection || receipt.Graph.MappingTotal != 7 || receipt.Graph.DanglingReferenceCount != 0 || !receipt.Graph.AlphaEquivalentSemanticGraph || receipt.Graph.RawEvidenceChanged != 7 || receipt.Graph.SemanticSlotDenominator != 7 || receipt.Graph.SemanticSlotUnique != 7 || receipt.Graph.SemanticSlotTotal != 7 || receipt.AlgorithmID != identityFaultAlgorithm || receipt.AlgorithmSourceBytes == 0 || receipt.AlgorithmSourceDigest == "" {
		t.Fatalf("raw fixture identity fault was not exact: %+v", receipt.Graph)
	}

	oldToNew, newToOld := producerIdentityFaultMaps(receipt.Graph.Mapping)
	stale := append([]ClaimIdentityRecord(nil), receipt.FaultedAlternate.Records...)
	for index := range stale {
		if stale[index].PreservationOf != "" {
			stale[index].PreservationOf = "stale-reference"
		}
	}
	graph, reason := validateIdentityFaultGraph(receipt.Alternate.Records, stale, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, receipt.Graph.Mapping, oldToNew, newToOld)
	assertProducerIdentityFaultFailure(t, graph, reason, "IDENTITY_REFERENCE_CLOSURE_BROKEN")

	swapped := append([]IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	swapped[0].NewStableID, swapped[1].NewStableID = swapped[1].NewStableID, swapped[0].NewStableID
	graph, reason = validateIdentityFaultGraph(receipt.Alternate.Records, receipt.FaultedAlternate.Records, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, swapped, oldToNew, newToOld)
	assertProducerIdentityFaultFailure(t, graph, reason, "IDENTITY_FAULT_MAPPING_RULE_MISMATCH")

	duplicate := append([]IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	duplicate[1].OldStableID = duplicate[0].OldStableID
	graph, reason = validateIdentityFaultGraph(receipt.Alternate.Records, receipt.FaultedAlternate.Records, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, duplicate, oldToNew, newToOld)
	assertProducerIdentityFaultFailure(t, graph, reason, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE")
}

func TestProducerRejectsDuplicateSemanticOrdinalSlot(t *testing.T) {
	receipt := producerIdentityFaultFixtureReceipt(t)
	duplicate := append([]ClaimIdentityRecord(nil), receipt.Alternate.Records...)
	duplicate[1].Kind = duplicate[0].Kind
	duplicate[1].RelationRole = duplicate[0].RelationRole
	duplicate[1].NormalizedProposition = duplicate[0].NormalizedProposition
	duplicate[1].PropositionDigest = duplicate[0].PropositionDigest
	duplicate[1].TargetAddress = duplicate[0].TargetAddress
	duplicate[1].TargetAddressDigest = duplicate[0].TargetAddressDigest
	_, graph, reason := rekeyIdentityFault(duplicate, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule})
	if reason != "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS" || graph.Decision != DecisionFailClosed || graph.Resolution != ResolutionLower || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != reason || graph.SemanticSlotUnique != 6 || graph.SemanticSlotTotal != 7 {
		t.Fatalf("duplicate semantic slot reason=%s graph=%+v", reason, graph)
	}
}

func TestProducerIdentityFaultCardinalityCounterexamplesUsePublicEntry(t *testing.T) {
	receipt := producerIdentityFaultFixtureReceipt(t)
	cases := []struct {
		id            string
		unique, total int
		reason        string
	}{
		{id: "six-unique-slots", unique: 6, total: 6, reason: "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"},
		{id: "eight-unique-slots", unique: 8, total: 8, reason: "IDENTITY_SEMANTIC_SLOT_DENOMINATOR_MISMATCH"},
		{id: "seven-slots-one-duplicate", unique: 6, total: 7, reason: "IDENTITY_SEMANTIC_SLOT_AMBIGUOUS"},
	}
	for _, testCase := range cases {
		original, faulted, rows := producerIdentityFaultCardinalityCase(receipt, testCase.id)
		graph, reason := ValidateIdentityFaultGraph(original, faulted, receipt.Alternate.SourcePair, identityFaultRule, rows)
		if reason != testCase.reason || graph.Decision != DecisionFailClosed || graph.Resolution != ResolutionLower || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != testCase.reason || graph.SemanticSlotDenominator != 7 || graph.SemanticSlotUnique != testCase.unique || graph.SemanticSlotTotal != testCase.total {
			t.Fatalf("cardinality case=%s reason=%s graph=%+v want=%s %d/%d", testCase.id, reason, graph, testCase.reason, testCase.unique, testCase.total)
		}
	}
}

func producerIdentityFaultCardinalityCase(receipt IdentityFaultReceipt, caseID string) ([]ClaimIdentityRecord, []ClaimIdentityRecord, []IdentityFaultMappingRow) {
	original := append([]ClaimIdentityRecord(nil), receipt.Alternate.Records...)
	faulted := append([]ClaimIdentityRecord(nil), receipt.FaultedAlternate.Records...)
	rows := append([]IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	switch caseID {
	case "six-unique-slots":
		oldID := original[len(original)-1].StableID
		newID := ""
		for _, row := range rows {
			if row.OldStableID == oldID {
				newID = row.NewStableID
			}
		}
		original = filterProducerIdentityFaultRecords(original, oldID)
		faulted = filterProducerIdentityFaultRecords(faulted, newID)
		rows = filterProducerIdentityFaultRows(rows, oldID)
	case "eight-unique-slots":
		originalExtra := original[0]
		originalExtra.StableID = originalExtra.StableID + "/cardinality-extra"
		originalExtra.Kind = originalExtra.Kind + "-cardinality-extra"
		originalExtra.PreservationOf = ""
		faultedExtra := faulted[0]
		faultedExtra.StableID = faultedExtra.StableID + "/cardinality-extra"
		faultedExtra.Kind = originalExtra.Kind
		faultedExtra.PreservationOf = ""
		original = append(original, originalExtra)
		faulted = append(faulted, faultedExtra)
		rows = append(rows, IdentityFaultMappingRow{OldStableID: originalExtra.StableID, NewStableID: faultedExtra.StableID, Ordinal: 7})
	case "seven-slots-one-duplicate":
		copyProducerIdentityFaultSemanticSlot(&original[1], original[0])
		copyProducerIdentityFaultSemanticSlot(&faulted[1], faulted[0])
	}
	return original, faulted, rows
}

func filterProducerIdentityFaultRecords(records []ClaimIdentityRecord, stableID string) []ClaimIdentityRecord {
	filtered := make([]ClaimIdentityRecord, 0, len(records)-1)
	for _, record := range records {
		if record.StableID != stableID {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func filterProducerIdentityFaultRows(rows []IdentityFaultMappingRow, oldID string) []IdentityFaultMappingRow {
	filtered := make([]IdentityFaultMappingRow, 0, len(rows)-1)
	for _, row := range rows {
		if row.OldStableID != oldID {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func copyProducerIdentityFaultSemanticSlot(target *ClaimIdentityRecord, source ClaimIdentityRecord) {
	target.Kind = source.Kind
	target.RelationRole = source.RelationRole
	target.NormalizedProposition = source.NormalizedProposition
	target.PropositionDigest = source.PropositionDigest
	target.TargetAddress = source.TargetAddress
	target.TargetAddressDigest = source.TargetAddressDigest
}

func producerIdentityFaultFixtureReceipt(t *testing.T) IdentityFaultReceipt {
	t.Helper()
	return IdentityFaultReceiptFromFiles(IdentityFaultInput{
		Baseline:     Input{CaseID: "persistence-probe", SubjectSHA: identityFaultTestSHA, ObservedCheckoutSHA: identityFaultTestSHA, BeforePath: identityFaultFixture("before.gooo"), AfterPath: identityFaultFixture("equivalent-after.gooo")},
		Alternate:    Input{CaseID: "persistence-probe", SubjectSHA: identityFaultTestSHA, ObservedCheckoutSHA: identityFaultTestSHA, BeforePath: identityFaultFixture("persistence-equivalent-before.gooo"), AfterPath: identityFaultFixture("persistence-equivalent-after.gooo")},
		ArtifactPath: identityFaultFixture("claim-identity-fault.json"),
	}).Receipt
}

func producerIdentityFaultMaps(rows []IdentityFaultMappingRow) (map[string]string, map[string]string) {
	oldToNew, newToOld := map[string]string{}, map[string]string{}
	for _, row := range rows {
		oldToNew[row.OldStableID] = row.NewStableID
		newToOld[row.NewStableID] = row.OldStableID
	}
	return oldToNew, newToOld
}

func assertProducerIdentityFaultFailure(t *testing.T, graph IdentityFaultGraphEvidence, reason, want string) {
	t.Helper()
	if reason != want || graph.Decision != DecisionFailClosed || graph.Resolution != ResolutionLower || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != want {
		t.Fatalf("identity fault tamper reason=%s graph=%+v want=%s", reason, graph, want)
	}
}

const identityFaultTestSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func identityFaultFixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	return filepath.Join(root, "examples", "semantic-delta-receipt", name)
}
