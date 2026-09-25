package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func testID(value string) semantic.ID { return semantic.MustIdentity("selective-ci-test://" + value) }
func testDigest(value string) string  { return semantic.StableHashString("selective-ci-test/" + value) }

type testFixture struct {
	input      Input
	root       semantic.ID
	obligation semantic.ID
	command    semantic.ID
	receipt    semantic.ID
	branch     semantic.InferenceEdge
}

func completeFixture() testFixture {
	base := semantic.SnapshotDigests{Source: testDigest("base-source"), Semantic: testDigest("base-semantic")}
	head := semantic.SnapshotDigests{Source: testDigest("head-source"), Semantic: testDigest("head-semantic")}
	binding := SnapshotBinding{Base: base, Head: head}
	root, obligation := testID("root"), testID("obligation")
	command, receipt := testID("command"), testID("receipt")
	registry, plan := testDigest("registry"), testDigest("plan")
	rule := semantic.RuleBinding{ID: testID("rule"), Version: "1", Digest: testDigest("rule")}
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
			controls.PolicyDigest, independent = testDigest("verification-policy"), true
		default:
			panic("unsupported fixture kind")
		}
		evidenceID, evidenceDigest := testID("evidence/"+label), testDigest("evidence/"+label)
		edge := semantic.InferenceEdge{
			RecordID: testID("record/" + label), SubjectID: subject, ObjectID: object, Rule: rule,
			Phase: phase, Before: base, After: head, Authority: authority,
			Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}}, Controls: controls, Kind: kind}
		if kind == semantic.InferenceAuthoritativeDeclaration {
			edge.SourceRoots = []semantic.ID{root}
		}
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: evidenceDigest, Before: base, After: head, SourceBacked: sourceBacked, Independent: independent, Controls: controls}
		return edge, evidence
	}
	first, firstEvidence := makeEdge("declaration", semantic.InferenceAuthoritativeDeclaration, root, obligation)
	second, secondEvidence := makeEdge("derivation", semantic.InferenceDeterministicDerivation, obligation, command)
	third, thirdEvidence := makeEdge("verification", semantic.InferenceIndependentVerification, command, receipt)
	path := Path{PathID: testID("path/main"), RootID: root, ObligationID: obligation, CommandID: command, ReceiptID: receipt,
		RecordIDs: []semantic.ID{first.RecordID, second.RecordID, third.RecordID}, ExpectedKinds: []semantic.InferenceKind{first.Kind, second.Kind, third.Kind}}
	commandReceipt := CommandReceipt{CommandID: command, ReceiptID: receipt, Status: ReceiptVerified,
		ProviderReceiptDigest: testDigest("provider"), PhaseReceiptDigest: testDigest("phase"), ResourceReceiptDigest: testDigest("resource"), RegistryDigest: registry, PlanDigest: plan}
	commandReceipt.Digest = commandReceipt.ExpectedDigest(binding)
	return testFixture{input: Input{Schema: SchemaVersion, Snapshots: binding, RegistryDigest: registry, PlanDigest: plan,
		ChangedRootIDs: []semantic.ID{root}, SelectedCommandIDs: []semantic.ID{command}, ObligationIDs: []semantic.ID{obligation}, Paths: []Path{path}, CommandReceipts: []CommandReceipt{commandReceipt},
		EvidenceIDs: []semantic.ID{firstEvidence.ID, secondEvidence.ID, thirdEvidence.ID}, InferencePath: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion,
			Edges: []semantic.InferenceEdge{first, second, third}, Evidence: []semantic.InferenceEvidence{firstEvidence, secondEvidence, thirdEvidence}}}, root: root, obligation: obligation, command: command, receipt: receipt}
}
