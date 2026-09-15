package main

import (
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func buildAnalyzerShadowSnapshot(t *testing.T, source, id, registryDigest string) analyzersci.Snapshot {
	t.Helper()
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: "cmd/gooo/fixture.gooo", PackagePath: "fixture", Source: []byte(source)}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semantic binding fixture = %#v, err = %v", result, err)
	}
	snapshot, err := analyzersci.Build(analyzersci.SnapshotInput{
		Sources:         []analyzersci.SourceInput{{Path: "cmd/gooo/fixture.gooo", BlobDigest: prefixedShadowDigest(source), Bindings: result.Bindings}},
		SourceMapDigest: prefixedShadowDigest("source-map"), RegistryDigest: registryDigest, RegisteredIDs: []string{id},
	})
	if err != nil {
		t.Fatalf("analyzer snapshot fixture = %v", err)
	}
	return snapshot
}
func shadowResourceEnvelope() resourceenvelope.Envelope {
	samples := []resourceenvelope.Sample{
		{CPUCoreNS: 1, WallNS: 10}, {CPUCoreNS: 10, WallNS: 10}, {CPUCoreNS: 20, WallNS: 10}, {CPUCoreNS: 30, WallNS: 10}, {CPUCoreNS: 40, WallNS: 10}, {CPUCoreNS: 50, WallNS: 10},
	}
	return resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: shadowDigest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: 100, PeakRSSBytes: 1000, ReadBytes: 1000, WriteBytes: 1000}, Samples: samples}
}
func shadowProofPath(t *testing.T, rootID, obligationID, commandID string, base, head semantic.SnapshotDigests) (plannersci.ProvenancePath, proofsci.Path, semantic.InferencePathV1, []semantic.ID) {
	t.Helper()
	root, obligation, command := commandIDToID(rootID), commandIDToID(obligationID), commandIDToID(commandID)
	receipt := commandIDToID("urn:gooo:shadow/receipt/test")
	rule := semantic.RuleBinding{ID: commandIDToID("urn:gooo:shadow/rule/v1"), Version: "1", Digest: shadowDigest("rule")}
	makeEdge := func(label string, kind semantic.InferenceKind, subject, object semantic.ID) (semantic.InferenceEdge, semantic.InferenceEvidence) {
		phase := semantic.PhasePlacement{Ordinal: 1}
		authority := semantic.AuthorityBinding{}
		controls := semantic.InferenceControls{}
		sourceBacked, independent := false, false
		switch kind {
		case semantic.InferenceAuthoritativeDeclaration:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDeclaration, semantic.AuthoritySource, semantic.AuthorityDeclare
			sourceBacked = true
		case semantic.InferenceDeterministicDerivation:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDerivation, semantic.AuthoritySemantic, semantic.AuthorityDerive
		case semantic.InferenceIndependentVerification:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseVerification, semantic.AuthorityVerification, semantic.AuthorityVerify
			controls.PolicyDigest, independent = shadowDigest("verification-policy"), true
		}
		evidenceID := commandIDToID("urn:gooo:shadow/evidence/" + label)
		evidenceDigest := shadowDigest("evidence/" + label)
		edge := semantic.InferenceEdge{RecordID: commandIDToID("urn:gooo:shadow/record/" + label), SubjectID: subject, ObjectID: object, Rule: rule, Phase: phase, Before: base, After: head, Authority: authority, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}}, Controls: controls, Kind: kind}
		if kind == semantic.InferenceAuthoritativeDeclaration {
			edge.SourceRoots = []semantic.ID{root}
		}
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: evidenceDigest, Before: base, After: head, SourceBacked: sourceBacked, Independent: independent, Controls: controls}
		return edge, evidence
	}
	first, firstEvidence := makeEdge("01-declaration", semantic.InferenceAuthoritativeDeclaration, root, obligation)
	second, secondEvidence := makeEdge("02-derivation", semantic.InferenceDeterministicDerivation, obligation, command)
	third, thirdEvidence := makeEdge("03-verification", semantic.InferenceIndependentVerification, command, receipt)
	edges := []semantic.InferenceEdge{first, second, third}
	evidence := []semantic.InferenceEvidence{firstEvidence, secondEvidence, thirdEvidence}
	recordIDs := []string{first.RecordID.String(), second.RecordID.String(), third.RecordID.String()}
	kinds := []string{string(first.Kind), string(second.Kind), string(third.Kind)}
	plannerPath := plannersci.ProvenancePath{CommandID: commandID, Path: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}, Requirement: plannersci.PathRequirement{PathID: "urn:gooo:shadow/path/main", RecordIDs: recordIDs, ExpectedKinds: kinds, StartID: rootID, EndID: receipt.String()}}
	proofPath := proofsci.Path{PathID: commandIDToID("urn:gooo:shadow/path/main"), RootID: root, ObligationID: obligation, CommandID: command, ReceiptID: receipt, RecordIDs: []semantic.ID{first.RecordID, second.RecordID, third.RecordID}, ExpectedKinds: []semantic.InferenceKind{first.Kind, second.Kind, third.Kind}}
	evidenceIDs := []semantic.ID{firstEvidence.ID, secondEvidence.ID, thirdEvidence.ID}
	return plannerPath, proofPath, semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}, evidenceIDs
}
