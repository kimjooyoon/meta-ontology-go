package shadow

import (
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func productionProofPath(t *testing.T, rootID, obligationID, commandID string, base, head semantic.SnapshotDigests) (plannersci.ProvenancePath, proofsci.Path, semantic.InferencePathV1, []semantic.ID) {
	t.Helper()
	root, obligation, command := productionID(rootID), productionID(obligationID), productionID(commandID)
	receipt := productionID("urn:gooo:shadow/receipt/test")
	rule := semantic.RuleBinding{ID: productionID("urn:gooo:shadow/rule/v1"), Version: "1", Digest: productionDigest("rule")}
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
			controls.PolicyDigest, independent = productionDigest("verification-policy"), true
		}
		evidenceID := productionID("urn:gooo:shadow/evidence/" + label)
		evidenceDigest := productionDigest("evidence/" + label)
		edge := semantic.InferenceEdge{RecordID: productionID("urn:gooo:shadow/record/" + label), SubjectID: subject, ObjectID: object, Rule: rule, Phase: phase, Before: base, After: head, Authority: authority, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}}, Controls: controls, Kind: kind}
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
	proofPath := proofsci.Path{PathID: productionID("urn:gooo:shadow/path/main"), RootID: root, ObligationID: obligation, CommandID: command, ReceiptID: receipt, RecordIDs: []semantic.ID{first.RecordID, second.RecordID, third.RecordID}, ExpectedKinds: []semantic.InferenceKind{first.Kind, second.Kind, third.Kind}}
	evidenceIDs := []semantic.ID{firstEvidence.ID, secondEvidence.ID, thirdEvidence.ID}
	return plannerPath, proofPath, semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}, evidenceIDs
}
