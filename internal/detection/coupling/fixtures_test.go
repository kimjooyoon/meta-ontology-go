package coupling

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type couplingFixture struct {
	input            Input
	authorityContext AuthorityContext
	owner            semantic.ID
	code             semantic.ID
	surface          semantic.ID
	authority        semantic.ID
	projection       semantic.ID
	verification     semantic.ID
}

type fixtureContext struct {
	owner, code, surface          semantic.ID
	registry                      Registry
	config                        Config
	manifest                      ChangeManifest
	beforeBlob, afterBlob         string
	beforeSemantic, afterSemantic string
	beforeSource, afterSource     string
}

type fixtureProof struct {
	path                                semantic.InferencePathV1
	receipt                             CouplingReceipt
	authority, projection, verification semantic.ID
}

func fixtureID(name string) semantic.ID { return semantic.MustIdentity("gooo://coupling/" + name) }

func fixtureDigest(name string) string { return semantic.StableHashString("coupling-fixture/" + name) }

func newFixture(t *testing.T, claim ChangeClaim) couplingFixture {
	t.Helper()
	context := fixtureContextFor(t, claim)
	proof := fixtureProofFor(context, claim)
	external := externalReceiptFor(context.config)
	authorityRegistry := context.registry
	authorityRegistry.Surfaces = append([]Surface(nil), context.registry.Surfaces...)
	authority := AuthorityContext{
		Schema: AuthorityContextSchemaV1, Registry: authorityRegistry,
		ToolchainDigest: context.config.ToolchainDigest, ProfileDigest: context.config.ProfileDigest,
		SnapshotDigest:         context.config.SnapshotDigest,
		ExpectedProviderDigest: context.config.ExpectedProviderDigest, ExpectedObserverDigest: context.config.ExpectedObserverDigest,
		Baseline: context.config.Baseline, ExternalReceiptRequired: true,
	}
	return couplingFixture{
		input: Input{Schema: InputSchemaV1, Config: context.config, Registry: context.registry, Manifest: context.manifest,
			Receipts: []CouplingReceipt{proof.receipt}, InferencePath: proof.path, ExternalReceipt: external, WorkspaceRoot: "/workspace"},
		authorityContext: authority,
		owner:            context.owner, code: context.code, surface: context.surface,
		authority: proof.authority, projection: proof.projection, verification: proof.verification,
	}
}

func fixtureContextFor(t *testing.T, claim ChangeClaim) fixtureContext {
	t.Helper()
	owner, code, surface := fixtureID("owner"), fixtureID("code-symbol"), fixtureID("surface")
	sourceMap := fixtureID("source-map")
	registrySurface := Surface{
		SurfaceID: surface, CodeSymbolID: code, SemanticOwnerID: owner,
		Binding:           SourceMapBinding{SourceMapID: sourceMap, PackageLabel: "billing", FileLabel: "old.go", SourceSpan: "1:1-1:2"},
		PresentationLabel: "PayOrder",
	}
	registrySurface.Binding.BindingDigest = bindingDigest(registrySurface)
	registry := Registry{Schema: RegistrySchemaV1, Surfaces: []Surface{registrySurface}}
	registry.Digest = stableDigest(registryCanonical(registry))
	baseline := BaselineConfig{Schema: BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = stableDigest(baselineCanonical(baseline))
	config := Config{
		Schema: ConfigSchemaV1, RegistryDigest: registry.Digest, ToolchainDigest: fixtureDigest("toolchain"),
		ProfileDigest: fixtureDigest("profile"), SnapshotDigest: fixtureDigest("snapshot-after"),
		ExpectedProviderDigest: fixtureDigest("provider"), ExpectedObserverDigest: fixtureDigest("observer"),
		Baseline: baseline, ExternalReceiptRequired: true,
	}
	beforeSemantic, afterSemantic := fixtureDigest("semantic-before"), fixtureDigest("semantic-after")
	beforeSource, afterSource := fixtureDigest("source-before"), fixtureDigest("source-after")
	if claim == ChangeClaimNoDelta {
		afterSemantic, afterSource = beforeSemantic, beforeSource
	}
	beforeBlob, afterBlob := fixtureDigest("blob-before"), fixtureDigest("blob-after")
	entry := ManifestEntry{
		SurfaceID: surface, CodeSymbolID: code, SemanticOwnerID: owner,
		BeforeBindingDigest: registrySurface.Binding.BindingDigest, AfterBindingDigest: registrySurface.Binding.BindingDigest,
		BeforeBlobDigest: beforeBlob, AfterBlobDigest: afterBlob,
		BeforeSourcePath: "/workspace/old.go", AfterSourcePath: "/relocated/new.go",
	}
	manifest := ChangeManifest{
		Schema: ManifestSchemaV1, Complete: true, ZeroChange: false, RegistryDigest: registry.Digest,
		ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeSnapshotDigest: fixtureDigest("snapshot-before"), AfterSnapshotDigest: config.SnapshotDigest,
		Entries: []ManifestEntry{entry},
	}
	manifest.Digest = stableDigest(manifestCanonical(manifest))
	return fixtureContext{owner: owner, code: code, surface: surface, registry: registry, config: config, manifest: manifest,
		beforeBlob: beforeBlob, afterBlob: afterBlob, beforeSemantic: beforeSemantic, afterSemantic: afterSemantic,
		beforeSource: beforeSource, afterSource: afterSource}
}

func fixtureProofFor(context fixtureContext, claim ChangeClaim) fixtureProof {
	owner, code, surface := context.owner, context.code, context.surface
	config := context.config
	authority, projection, verification := fixtureID("authority-edge"), fixtureID("projection-edge"), fixtureID("verification-edge")
	sourceID := fixtureID("authoritative-source")
	authEvidence, projectionEvidence, verificationEvidence := fixtureID("auth-evidence"), fixtureID("projection-evidence"), fixtureID("verification-evidence")
	path := fixturePath(owner, code, surface, authority, projection, verification, sourceID, authEvidence, projectionEvidence, verificationEvidence, context.beforeSemantic, context.afterSemantic, context.beforeSource, context.afterSource, config)
	claimRecord := semantic.InferenceRecord{
		RecordID: fixtureID("claim"), SubjectID: owner, ObjectID: surface,
		Rule:      semantic.RuleBinding{ID: fixtureID("rule"), Version: "v1", Digest: fixtureDigest("rule")},
		Phase:     semantic.PhasePlacement{Phase: semantic.PhaseVerification, Ordinal: 4},
		Before:    semantic.SnapshotDigests{Source: context.beforeSource, Semantic: context.beforeSemantic},
		After:     semantic.SnapshotDigests{Source: context.afterSource, Semantic: context.afterSemantic},
		Authority: semantic.AuthorityBinding{Layer: semantic.AuthoritySemantic, Effect: semantic.AuthorityDelta},
		Evidence:  []semantic.EvidenceReference{{ID: verificationEvidence, Digest: fixtureDigest("evidence-verification-evidence")}},
		Controls:  semantic.InferenceControls{PolicyDigest: config.ProfileDigest},
	}
	if claim == ChangeClaimNoDelta {
		claimRecord.Authority.Effect = semantic.AuthorityNoChange
	}
	canonicalDelta := "owner=" + owner.String() + " relation=used"
	semanticClaim := semantic.SemanticChangeClaim{InferenceRecord: claimRecord, Kind: semantic.SemanticDelta, CanonicalDelta: canonicalDelta, DeltaDigest: stableDigest(canonicalDelta)}
	if claim == ChangeClaimNoDelta {
		semanticClaim = semantic.SemanticChangeClaim{InferenceRecord: claimRecord, Kind: semantic.NoSemanticDelta}
	}
	path.Claims = []semantic.SemanticChangeClaim{semanticClaim}
	receipt := CouplingReceipt{
		Schema: ReceiptSchemaV1, ReceiptID: fixtureID("receipt"), SurfaceID: surface, SemanticOwnerID: owner, CodeSymbolID: code,
		SourceMapBindingDigest: context.registry.Surfaces[0].Binding.BindingDigest, SnapshotDigest: config.SnapshotDigest,
		RegistryDigest: context.registry.Digest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeBlobDigest: context.beforeBlob, AfterBlobDigest: context.afterBlob, BeforeAuthoritySourceDigest: context.beforeSource, AfterAuthoritySourceDigest: context.afterSource,
		BeforeCanonicalSemanticDigest: context.beforeSemantic, AfterCanonicalSemanticDigest: context.afterSemantic,
		ChangeClaim: claim, ReceiptKind: receiptKind(claim), OriginPathIDs: []semantic.ID{verification, projection, authority},
		InferenceClaimID: semanticClaim.RecordID, EvidenceRefs: []semantic.EvidenceReference{
			{ID: authEvidence, Digest: fixtureDigest("evidence-auth-evidence")},
			{ID: projectionEvidence, Digest: fixtureDigest("evidence-projection-evidence")},
			{ID: verificationEvidence, Digest: fixtureDigest("evidence-verification-evidence")},
		}, State: ReceiptStateCurrent,
	}
	if claim == ChangeClaimDelta {
		receipt.CanonicalDelta, receipt.DeltaDigest = canonicalDelta, stableDigest(canonicalDelta)
		receipt.AuthoritativeSource = &AuthoritySource{SourceID: sourceID, Path: "billing/authority.gooo", Span: "10:1-10:12"}
	}
	return fixtureProof{path: path, receipt: receipt, authority: authority, projection: projection, verification: verification}
}

func externalReceiptFor(config Config) *ExternalResourceReceipt {
	cpu, memory, work := uint64(7), uint64(11), uint64(13)
	external := &ExternalResourceReceipt{Schema: ResourceSchemaV1, SnapshotDigest: config.SnapshotDigest, ProviderDigest: fixtureDigest("provider"), ObserverDigest: fixtureDigest("observer"), CPUWorkUnits: &cpu, PeakMemoryBytes: &memory, DeterministicWorkUnits: &work}
	external.Digest = stableDigest(externalCanonical(*external))
	return external
}

func receiptKind(claim ChangeClaim) semantic.SemanticChangeKind {
	if claim == ChangeClaimDelta {
		return semantic.SemanticDelta
	}
	return semantic.NoSemanticDelta
}

func fixturePath(owner, code, surface, authority, projection, verification, sourceID, authEvidence, projectionEvidence, verificationEvidence semantic.ID, beforeSemantic, afterSemantic, beforeSource, afterSource string, config Config) semantic.InferencePathV1 {
	rule := semantic.RuleBinding{ID: fixtureID("rule"), Version: "v1", Digest: fixtureDigest("rule")}
	base := func(recordID, subject, object semantic.ID, phase semantic.InferencePhase, ordinal uint64, kind semantic.InferenceKind, evidence semantic.ID, controls semantic.InferenceControls) semantic.InferenceEdge {
		return semantic.InferenceEdge{InferenceRecord: semantic.InferenceRecord{RecordID: recordID, SubjectID: subject, ObjectID: object, Rule: rule, Phase: semantic.PhasePlacement{Phase: phase, Ordinal: ordinal}, Before: semantic.SnapshotDigests{Source: beforeSource, Semantic: beforeSemantic}, After: semantic.SnapshotDigests{Source: afterSource, Semantic: afterSemantic}, Authority: semantic.AuthorityBinding{Layer: layerFor(kind), Effect: effectFor(kind)}, Evidence: []semantic.EvidenceReference{{ID: evidence, Digest: fixtureDigest("evidence-" + evidence.String()[len("gooo://coupling/"):])}}, Controls: controls}, Kind: kind, SourceRoots: []semantic.ID{sourceID}}
	}
	authControls := semantic.InferenceControls{}
	projectionControls := semantic.InferenceControls{Profile: semantic.ProfileBinding{ID: "coupling-profile", Version: "v1", Digest: config.ProfileDigest}}
	verificationControls := semantic.InferenceControls{PolicyDigest: config.ProfileDigest}
	edges := []semantic.InferenceEdge{
		base(authority, owner, code, semantic.PhaseDeclaration, 1, semantic.InferenceAuthoritativeDeclaration, authEvidence, authControls),
		base(projection, code, surface, semantic.PhaseProjection, 2, semantic.InferenceDerivedProjection, projectionEvidence, projectionControls),
		base(verification, surface, verificationEvidence, semantic.PhaseVerification, 3, semantic.InferenceIndependentVerification, verificationEvidence, verificationControls),
	}
	edges[0].SourceRoots = []semantic.ID{sourceID}
	if beforeSemantic == afterSemantic {
		edges[0].After = edges[0].Before
	}
	if beforeSource == afterSource {
		edges[0].After.Source = edges[0].Before.Source
	}
	evidence := []semantic.InferenceEvidence{
		{ID: authEvidence, Digest: fixtureDigest("evidence-auth-evidence"), Before: edges[0].Before, After: edges[0].After, Controls: authControls, SourceBacked: true},
		{ID: projectionEvidence, Digest: fixtureDigest("evidence-projection-evidence"), Before: edges[1].Before, After: edges[1].After, Controls: projectionControls},
		{ID: verificationEvidence, Digest: fixtureDigest("evidence-verification-evidence"), Before: edges[2].Before, After: edges[2].After, Controls: verificationControls, Independent: true},
	}
	return semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}
}

func layerFor(kind semantic.InferenceKind) semantic.AuthorityLayer {
	if kind == semantic.InferenceAuthoritativeDeclaration {
		return semantic.AuthoritySource
	}
	if kind == semantic.InferenceDerivedProjection {
		return semantic.AuthorityDerived
	}
	return semantic.AuthorityVerification
}

func effectFor(kind semantic.InferenceKind) semantic.AuthorityEffect {
	if kind == semantic.InferenceAuthoritativeDeclaration {
		return semantic.AuthorityDeclare
	}
	if kind == semantic.InferenceDerivedProjection {
		return semantic.AuthorityProject
	}
	return semantic.AuthorityVerify
}
