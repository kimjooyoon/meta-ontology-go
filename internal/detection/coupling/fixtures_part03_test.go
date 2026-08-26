package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
