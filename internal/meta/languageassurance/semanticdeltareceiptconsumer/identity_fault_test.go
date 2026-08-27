package semanticdeltareceiptconsumer

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestConsumerIdentityFaultUsesRawFixtureAndRejectsClosedGraphTampering(t *testing.T) {
	receipt := consumerIdentityFaultFixtureReceipt(t)
	if !receipt.FaultGraphClosed || !receipt.Graph.Bijection || receipt.Graph.MappingTotal != 7 || receipt.Graph.DanglingReferenceCount != 0 || !receipt.Graph.AlphaEquivalentSemanticGraph || receipt.Graph.RawEvidenceChanged != 7 {
		t.Fatalf("raw fixture identity fault was not exact: %+v", receipt.Graph)
	}

	oldToNew, newToOld := consumerIdentityFaultMaps(receipt.Graph.Mapping)
	stale := append([]ClaimIdentityRecord(nil), receipt.FaultedAlternate.Records...)
	for index := range stale {
		if stale[index].PreservationOf != "" {
			stale[index].PreservationOf = "stale-reference"
		}
	}
	graph, reason := validateIdentityFaultGraph(receipt.Alternate.Records, stale, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, receipt.Graph.Mapping, oldToNew, newToOld)
	assertConsumerIdentityFaultFailure(t, graph, reason, "IDENTITY_REFERENCE_CLOSURE_BROKEN")

	swapped := append([]IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	swapped[0].NewStableID, swapped[1].NewStableID = swapped[1].NewStableID, swapped[0].NewStableID
	graph, reason = validateIdentityFaultGraph(receipt.Alternate.Records, receipt.FaultedAlternate.Records, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, swapped, oldToNew, newToOld)
	assertConsumerIdentityFaultFailure(t, graph, reason, "IDENTITY_FAULT_MAPPING_RULE_MISMATCH")

	duplicate := append([]IdentityFaultMappingRow(nil), receipt.Graph.Mapping...)
	duplicate[1].OldStableID = duplicate[0].OldStableID
	graph, reason = validateIdentityFaultGraph(receipt.Alternate.Records, receipt.FaultedAlternate.Records, receipt.Alternate.SourcePair, identityFaultArtifact{Rule: identityFaultRule}, duplicate, oldToNew, newToOld)
	assertConsumerIdentityFaultFailure(t, graph, reason, "IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE")
}

func consumerIdentityFaultFixtureReceipt(t *testing.T) IdentityFaultReceipt {
	t.Helper()
	return IdentityFaultReceiptFromFiles(IdentityFaultInput{
		Baseline:     Input{CaseID: "persistence-probe", SubjectSHA: identityFaultTestSHA, ObservedCheckoutSHA: identityFaultTestSHA, BeforePath: identityFaultFixture("before.gooo"), AfterPath: identityFaultFixture("equivalent-after.gooo")},
		Alternate:    Input{CaseID: "persistence-probe", SubjectSHA: identityFaultTestSHA, ObservedCheckoutSHA: identityFaultTestSHA, BeforePath: identityFaultFixture("persistence-equivalent-before.gooo"), AfterPath: identityFaultFixture("persistence-equivalent-after.gooo")},
		ArtifactPath: identityFaultFixture("claim-identity-fault.json"),
	})
}

func consumerIdentityFaultMaps(rows []IdentityFaultMappingRow) (map[string]string, map[string]string) {
	oldToNew, newToOld := map[string]string{}, map[string]string{}
	for _, row := range rows {
		oldToNew[row.OldStableID] = row.NewStableID
		newToOld[row.NewStableID] = row.OldStableID
	}
	return oldToNew, newToOld
}

func assertConsumerIdentityFaultFailure(t *testing.T, graph IdentityFaultGraphEvidence, reason, want string) {
	t.Helper()
	if reason != want || graph.Decision != decisionFailClosed || graph.Resolution != resolutionLower || graph.Stage != "identity-fault" || graph.Step != "rekey-graph" || graph.Reason != want {
		t.Fatalf("identity fault tamper reason=%s graph=%+v want=%s", reason, graph, want)
	}
}

const identityFaultTestSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func identityFaultFixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	return filepath.Join(root, "examples", "semantic-delta-receipt", name)
}
