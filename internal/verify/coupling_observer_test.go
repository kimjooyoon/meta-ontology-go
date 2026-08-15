package verify

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func couplingDigest(value string) string {
	return semantic.StableHashString("coupling-fixture/" + value)
}
func couplingID(value string) semantic.ID { return semantic.MustIdentity("gooo://coupling/" + value) }
func couplingPathID(value string) semantic.ID {
	if strings.Contains(value, "://") {
		return semantic.MustIdentity(value)
	}
	return couplingID(value)
}

func couplingSurface() CouplingSurface {
	return CouplingSurface{
		SurfaceID: "gooo://coupling/surface/pay-order", CodeSymbolID: "gooo://coupling/code/pay-order",
		SemanticOwnerID: "gooo://coupling/semantic/pay-order", ScopeID: "gooo://coupling/scope/billing",
		CodePathPatterns: []string{"internal/billing/pay.go"}, SourceMapID: "gooo://coupling/map/pay-order",
		SourceMapBindingDigest: couplingDigest("source-map"), SemanticSourceIDs: []string{"gooo://billing/activity/pay-order"},
		SemanticSourcePaths: []string{"examples/billing/main.gooo"}, ProfileID: "gooo.profile.billing.v1", ProfileVersion: "1",
		ProfileDigest: couplingDigest("profile"), ToolchainDigest: couplingDigest("toolchain"),
		RuleDigests: []string{couplingDigest("rule")}, Applicability: CouplingApplicable,
	}
}

func couplingFixture(t *testing.T, claim string) CouplingInput {
	t.Helper()
	surface := couplingSurface()
	registry := CouplingRegistry{Schema: CouplingRegistrySchemaVersion, Version: "1", Surfaces: []CouplingSurface{surface}}
	before := couplingDigest("ir-before")
	after := before
	if claim == "DELTA" {
		after = couplingDigest("ir-after")
	}
	envelope := CouplingEnvelope{
		Schema: CouplingEnvelopeSchemaVersion, ContractDigest: CouplingContractDigest, Repository: "owner/repo",
		Event: "pull_request", Ref: "refs/pull/17/merge", EventRef: "refs/pull/17/merge", CheckoutRef: strings.Repeat("2", 40),
		BaseRef: "dev", BaseSHA: strings.Repeat("1", 40), HeadSHA: strings.Repeat("2", 40), WorkflowSHA: strings.Repeat("3", 40),
		PRNumber: 17, RunID: 101, RunAttempt: 1, CatalogDigest: couplingDigest("catalog"), PolicyDigest: couplingDigest("policy"),
		RegistryDigest: registry.Digest(), ProfileDigest: surface.ProfileDigest, ToolchainDigest: surface.ToolchainDigest,
		SchemaDigest: semantic.StableHashString(CouplingEvidenceSchemaVersion), SnapshotDigest: couplingDigest("snapshot"),
		SemanticSourceHead: CouplingSemanticSourceHead, SemanticPathSchema: semantic.InferencePathSchemaVersion,
	}
	path := couplingPath(claim, before, after, envelope.PolicyDigest, surface.SemanticOwnerID, surface.SemanticSourceIDs[0], surface.RuleDigests[0])
	receipt := CouplingReceipt{
		ReceiptID: "gooo://coupling/receipt/pay-order", SurfaceID: surface.SurfaceID, SemanticOwnerID: surface.SemanticOwnerID,
		CodeSymbolID: surface.CodeSymbolID, EnvelopeDigest: envelope.TupleDigest(), SnapshotDigest: envelope.SnapshotDigest,
		RegistryDigest: envelope.RegistryDigest, ProfileDigest: surface.ProfileDigest, ToolchainDigest: surface.ToolchainDigest,
		RuleDigest: surface.RuleDigests[0], SourceMapBindingDigest: surface.SourceMapBindingDigest, BeforeIRDigest: before,
		AfterIRDigest: after, ChangeClaim: claim, ReceiptKind: expectedReceiptKind(claim), PathDigest: path.StableHash(),
		OriginPathIDs: edgeIDs(path.Edges), EvidenceRefs: evidenceIDs(path.Evidence), State: CouplingCurrent, Path: path,
	}
	if claim == "DELTA" {
		receipt.AuthoritySourceBeforeDigest = couplingDigest("source-before")
		receipt.AuthoritySourceAfterDigest = couplingDigest("source-after")
		receipt.CanonicalDelta = "relation\tpay-order\tchanged"
		receipt.DeltaDigest = semantic.StableHashString(receipt.CanonicalDelta)
	}
	refreshCouplingReceipt(t, &receipt)
	return CouplingInput{Schema: CouplingEvidenceSchemaVersion, Envelope: envelope, Registry: registry, ChangedSites: []ChangedCodeSite{{Path: "internal/billing/pay.go", CodeSymbolID: surface.CodeSymbolID, SourceMapBindingDigest: surface.SourceMapBindingDigest}}, Receipts: []CouplingReceipt{receipt}}
}

func couplingPath(claim string, before, after, policy, owner, source, ruleDigest string) semantic.InferencePathV1 {
	rule := semantic.RuleBinding{ID: couplingID("rule/pay-order"), Version: "1", Digest: ruleDigest}
	controls := semantic.InferenceControls{}
	beforeSnapshot := semantic.SnapshotDigests{Source: couplingDigest("source-before"), Semantic: before}
	afterSnapshot := semantic.SnapshotDigests{Source: couplingDigest("source-after"), Semantic: after}
	if claim == "NO_DELTA" {
		afterSnapshot.Source = beforeSnapshot.Source
	}
	makeRecord := func(id, subject, object string, phase semantic.InferencePhase, layer semantic.AuthorityLayer, effect semantic.AuthorityEffect, evidence semantic.ID, extra semantic.InferenceControls) semantic.InferenceRecord {
		return semantic.InferenceRecord{RecordID: couplingID(id), SubjectID: couplingPathID(subject), ObjectID: couplingPathID(object), Rule: rule, Phase: semantic.PhasePlacement{Phase: phase, Ordinal: 1}, Before: beforeSnapshot, After: afterSnapshot, Authority: semantic.AuthorityBinding{Layer: layer, Effect: effect}, Evidence: []semantic.EvidenceReference{{ID: evidence, Digest: couplingDigest(id + "/evidence")}}, Controls: extra}
	}
	evidence := func(id string, record semantic.InferenceRecord, sourceBacked, independent bool) semantic.InferenceEvidence {
		return semantic.InferenceEvidence{ID: record.Evidence[0].ID, Digest: record.Evidence[0].Digest, Before: record.Before, After: record.After, SourceBacked: sourceBacked, Independent: independent, Controls: record.Controls}
	}
	declarationEvidence := couplingID("evidence/declaration")
	derivationEvidence := couplingID("evidence/derivation")
	verificationEvidence := couplingID("evidence/verification")
	claimEvidence := couplingID("evidence/claim")
	declaration := makeRecord("edge/declaration", source, "gooo://coupling/path/derived", semantic.PhaseDeclaration, semantic.AuthoritySource, semantic.AuthorityDeclare, declarationEvidence, controls)
	derivation := makeRecord("edge/derivation", "gooo://coupling/path/derived", "gooo://coupling/path/verified", semantic.PhaseDerivation, semantic.AuthoritySemantic, semantic.AuthorityDerive, derivationEvidence, controls)
	verificationControls := semantic.InferenceControls{PolicyDigest: policy}
	verification := makeRecord("edge/verification", "gooo://coupling/path/verified", owner, semantic.PhaseVerification, semantic.AuthorityVerification, semantic.AuthorityVerify, verificationEvidence, verificationControls)
	claimControls := controls
	claimAuthority := semantic.AuthorityBinding{Layer: semantic.AuthoritySemantic, Effect: semantic.AuthorityNoChange}
	kind := semantic.NoSemanticDelta
	delta := ""
	deltaDigest := ""
	if claim == "DELTA" {
		kind, claimAuthority.Effect = semantic.SemanticDelta, semantic.AuthorityDelta
		delta = "relation\tpay-order\tchanged"
		deltaDigest = semantic.StableHashString(delta)
	}
	claimRecord := semantic.InferenceRecord{RecordID: couplingID("claim/pay-order"), SubjectID: couplingPathID(source), ObjectID: couplingPathID(owner), Rule: rule, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDerivation, Ordinal: 1}, Before: beforeSnapshot, After: afterSnapshot, Authority: claimAuthority, Evidence: []semantic.EvidenceReference{{ID: claimEvidence, Digest: couplingDigest("evidence/claim/evidence")}}, Controls: claimControls}
	declaration.Evidence[0].Digest = couplingDigest("edge/declaration/evidence")
	derivation.Evidence[0].Digest = couplingDigest("edge/derivation/evidence")
	verification.Evidence[0].Digest = couplingDigest("edge/verification/evidence")
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: []semantic.InferenceEdge{{InferenceRecord: declaration, Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{couplingPathID(source)}}, {InferenceRecord: derivation, Kind: semantic.InferenceDeterministicDerivation}, {InferenceRecord: verification, Kind: semantic.InferenceIndependentVerification}}, Claims: []semantic.SemanticChangeClaim{{InferenceRecord: claimRecord, Kind: kind, CanonicalDelta: delta, DeltaDigest: deltaDigest}}, Evidence: []semantic.InferenceEvidence{evidence("declaration", declaration, true, false), evidence("derivation", derivation, false, false), evidence("verification", verification, false, true), {ID: claimEvidence, Digest: couplingDigest("evidence/claim/evidence"), Before: beforeSnapshot, After: afterSnapshot, SourceBacked: claim == "DELTA", Controls: claimControls}}}
	return path
}

func refreshCouplingReceipt(t *testing.T, receipt *CouplingReceipt) {
	t.Helper()
	receipt.PathDigest = receipt.Path.StableHash()
	payload, err := receipt.ExpectedCanonicalPayload()
	if err != nil {
		t.Fatal(err)
	}
	receipt.CanonicalPayload = payload
}

func hasCouplingFailure(evidence CouplingEvidence, reason string) bool {
	want := CouplingFailureCodePrefix + reason
	for _, failure := range evidence.Failures {
		if failure.Code == want {
			return true
		}
	}
	return false
}

type couplingFixtureCase struct {
	name, claim, wantDecision, wantReason string
	mutate                                func(*CouplingInput)
}

func couplingAdversarialFixtures(t *testing.T) []couplingFixtureCase {
	return []couplingFixtureCase{
		{"code changed no receipt", "NO_DELTA", CouplingDecisionFailClosed, "missing-receipt", func(in *CouplingInput) { in.Receipts = nil }},
		{"meaningless forced edit", "NO_DELTA", CouplingDecisionFailClosed, "no-delta-without-equality", func(in *CouplingInput) {
			in.Receipts[0].CanonicalDelta = "forced"
			in.Receipts[0].DeltaDigest = couplingDigest("forced")
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"equal digest DELTA", "NO_DELTA", CouplingDecisionFailClosed, "delta-without-change", func(in *CouplingInput) {
			in.Receipts[0].ChangeClaim = "DELTA"
			in.Receipts[0].ReceiptKind = expectedReceiptKind("DELTA")
			in.Receipts[0].CanonicalDelta = "forced"
			in.Receipts[0].DeltaDigest = semantic.StableHashString("forced")
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"unequal NO_DELTA", "DELTA", CouplingDecisionFailClosed, "no-delta-without-equality", func(in *CouplingInput) {
			in.Receipts[0].ChangeClaim = "NO_DELTA"
			in.Receipts[0].ReceiptKind = expectedReceiptKind("NO_DELTA")
			in.Receipts[0].CanonicalDelta = ""
			in.Receipts[0].DeltaDigest = ""
			in.Receipts[0].AuthoritySourceBeforeDigest = ""
			in.Receipts[0].AuthoritySourceAfterDigest = ""
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"stale registry", "NO_DELTA", CouplingDecisionUnknown, "registry-mismatch", func(in *CouplingInput) { in.Envelope.RegistryDigest = couplingDigest("stale-registry") }},
		{"wrong owner", "NO_DELTA", CouplingDecisionFailClosed, "surface-owner-mismatch", func(in *CouplingInput) {
			in.Receipts[0].SemanticOwnerID = "gooo://coupling/semantic/wrong"
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"alias collision", "NO_DELTA", CouplingDecisionUnknown, "ambiguous-origin", func(in *CouplingInput) {
			second := couplingSurface()
			second.SurfaceID = "gooo://coupling/surface/alias"
			second.CodeSymbolID = "gooo://coupling/code/alias"
			second.SemanticOwnerID = "gooo://coupling/semantic/alias"
			in.Registry.Surfaces = append(in.Registry.Surfaces, second)
			in.ChangedSites[0].CodeSymbolID = ""
			in.Envelope.RegistryDigest = in.Registry.Digest()
		}},
		{"observation promotion", "DELTA", CouplingDecisionFailClosed, "observation-promotion", func(in *CouplingInput) {
			edge := &in.Receipts[0].Path.Edges[0]
			edge.Kind = semantic.InferenceObservationCandidate
			edge.Phase = semantic.PhasePlacement{Phase: semantic.PhaseObservation, Ordinal: 1}
			edge.Authority = semantic.AuthorityBinding{Layer: semantic.AuthorityCandidate, Effect: semantic.AuthorityObserve}
			edge.Controls = semantic.InferenceControls{CatalogDigest: in.Envelope.CatalogDigest}
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"missing rule", "NO_DELTA", CouplingDecisionFailClosed, "rule-mismatch", func(in *CouplingInput) {
			in.Receipts[0].RuleDigest = couplingDigest("missing-rule")
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"missing evidence", "NO_DELTA", CouplingDecisionFailClosed, "path-incomplete", func(in *CouplingInput) {
			in.Receipts[0].Path.Evidence = in.Receipts[0].Path.Evidence[:1]
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"duplicate receipt", "NO_DELTA", CouplingDecisionFailClosed, "duplicate-receipt", func(in *CouplingInput) { in.Receipts = append(in.Receipts, in.Receipts[0]) }},
		{"orphan receipt", "NO_DELTA", CouplingDecisionFailClosed, "orphan-receipt", func(in *CouplingInput) {
			orphan := in.Receipts[0]
			orphan.SurfaceID = "gooo://coupling/surface/orphan"
			in.Receipts = append(in.Receipts, orphan)
		}},
		{"unrelated green run", "NO_DELTA", CouplingDecisionFailClosed, "wrong-tuple", func(in *CouplingInput) {
			in.Receipts[0].EnvelopeDigest = couplingDigest("unrelated-run")
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"wrong tuple", "NO_DELTA", CouplingDecisionFailClosed, "wrong-tuple", func(in *CouplingInput) {
			in.Receipts[0].EnvelopeDigest = couplingDigest("wrong-head")
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
		{"noncanonical serialization", "NO_DELTA", CouplingDecisionFailClosed, "noncanonical-evidence", func(in *CouplingInput) { in.Receipts[0].CanonicalPayload = "{\"surface_id\":\"noncanonical\"}" }},
	}
}

func couplingPositiveFixtures() []couplingFixtureCase {
	return []couplingFixtureCase{
		{"valid semantic delta", "DELTA", CouplingDecisionPass, "", func(*CouplingInput) {}},
		{"valid implementation-only no delta", "NO_DELTA", CouplingDecisionPass, "", func(*CouplingInput) {}},
	}
}

func runCouplingFixtureCase(t *testing.T, test couplingFixtureCase) {
	t.Helper()
	input := couplingFixture(t, test.claim)
	test.mutate(&input)
	got := VerifyCoupling(input)
	if got.RawDecision != test.wantDecision {
		t.Fatalf("decision = %s, want %s; failures=%+v", got.RawDecision, test.wantDecision, got.Failures)
	}
	if test.wantReason != "" && !hasCouplingFailure(got, test.wantReason) {
		t.Fatalf("missing failure reason %s: %+v", test.wantReason, got.Failures)
	}
	if err := got.ValidateCanonical(); err != nil {
		t.Fatal(err)
	}
}

func TestCouplingObserverFixtureMatrix(t *testing.T) {
	tests := couplingAdversarialFixtures(t)
	tests = append(tests, couplingPositiveFixtures()...)
	if len(tests) != 17 {
		t.Fatalf("fixture denominator = %d, want 17", len(tests))
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCouplingFixtureCase(t, test)
		})
	}
}

func TestCouplingObserverInsertionReplayAndNoWrite(t *testing.T) {
	input := couplingFixture(t, "DELTA")
	before := input
	first := VerifyCoupling(input)
	second := VerifyCoupling(input)
	left, _ := first.CanonicalJSON()
	right, _ := second.CanonicalJSON()
	if !strings.Contains(string(left), `"receipt_id"`) || string(left) != string(right) || !reflect.DeepEqual(input, before) {
		t.Fatal("clean replay was not deterministic or mutated input")
	}
	cleanRootOne := VerifyCoupling(couplingFixture(t, "DELTA"))
	cleanRootTwo := VerifyCoupling(couplingFixture(t, "DELTA"))
	rootOneJSON, _ := cleanRootOne.CanonicalJSON()
	rootTwoJSON, _ := cleanRootTwo.CanonicalJSON()
	if string(rootOneJSON) != string(rootTwoJSON) {
		t.Fatal("two clean roots produced different canonical evidence")
	}
	reordered := input
	reordered.Receipts = append([]CouplingReceipt(nil), input.Receipts...)
	reordered.Receipts[0].Path.Edges = append([]semantic.InferenceEdge(nil), input.Receipts[0].Path.Edges...)
	reordered.Receipts[0].Path.Evidence = append([]semantic.InferenceEvidence(nil), input.Receipts[0].Path.Evidence...)
	for i, j := 0, len(reordered.Receipts[0].Path.Edges)-1; i < j; i, j = i+1, j-1 {
		reordered.Receipts[0].Path.Edges[i], reordered.Receipts[0].Path.Edges[j] = reordered.Receipts[0].Path.Edges[j], reordered.Receipts[0].Path.Edges[i]
	}
	for i, j := 0, len(reordered.Receipts[0].Path.Evidence)-1; i < j; i, j = i+1, j-1 {
		reordered.Receipts[0].Path.Evidence[i], reordered.Receipts[0].Path.Evidence[j] = reordered.Receipts[0].Path.Evidence[j], reordered.Receipts[0].Path.Evidence[i]
	}
	refreshCouplingReceipt(t, &reordered.Receipts[0])
	third := VerifyCoupling(reordered)
	thirdJSON, _ := third.CanonicalJSON()
	if string(left) != string(thirdJSON) {
		t.Fatal("insertion-order replay changed canonical evidence")
	}
	lines, err := first.CanonicalJSONL()
	if err != nil || len(lines) == 0 || lines[len(lines)-1] != '\n' {
		t.Fatalf("JSONL evidence = %q, err=%v", lines, err)
	}
}

func TestCouplingRegistryDigestIgnoresInsertionOrder(t *testing.T) {
	first := couplingSurface()
	second := first
	second.SurfaceID = "gooo://coupling/surface/second"
	second.CodeSymbolID = "gooo://coupling/code/second"
	second.SemanticOwnerID = "gooo://coupling/semantic/second"
	second.CodePathPatterns = []string{"internal/billing/second.go"}
	left := CouplingRegistry{Schema: CouplingRegistrySchemaVersion, Version: "1", Surfaces: []CouplingSurface{first, second}}
	right := CouplingRegistry{Schema: CouplingRegistrySchemaVersion, Version: "1", Surfaces: []CouplingSurface{second, first}}
	if left.Digest() != right.Digest() {
		t.Fatal("registry digest depends on insertion order")
	}
}
