package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func makePathAndReceipt(input Input, registry registryView, beforeDigest, afterDigest, deltaText string, claim ChangeClaim, candidate bool) ([]CouplingReceipt, semantic.InferencePathV1) {
	binding := registry.bySurface["urn:gooo:surface:billing/pay-order"]
	root := semantic.MustIdentity("urn:gooo:source:billing")
	owner := semantic.MustIdentity(binding.SemanticOwnerID)
	term := semantic.MustIdentity("urn:gooo:term:billing/pay-order")
	code := semantic.MustIdentity(binding.CodeSymbolID)
	receiptID := semantic.MustIdentity("urn:gooo:receipt:billing/pay-order")
	controls := semantic.InferenceControls{CatalogDigest: digestText("catalog-v1"), PolicyDigest: digestText("policy-v1"), Profile: semantic.ProfileBinding{ID: input.Config.Profile.ID, Version: input.Config.Profile.Version, Digest: input.Config.Profile.Digest}}
	snapshotBefore, snapshotAfter := semantic.SnapshotDigests{Source: sourceDigest(input.AuthoritySourceBefore), Semantic: beforeDigest}, semantic.SnapshotDigests{Source: sourceDigest(input.AuthoritySourceAfter), Semantic: afterDigest}
	evidence := []semantic.InferenceEvidence{}
	addEvidence := func(id string, independent, sourceBacked bool) semantic.EvidenceReference {
		parsed := semantic.MustIdentity(id)
		digest := digestText("evidence:" + id)
		evidence = append(evidence, semantic.InferenceEvidence{ID: parsed, Digest: digest, Before: snapshotBefore, After: snapshotAfter, SourceBacked: sourceBacked, Independent: independent, Controls: controls})
		return semantic.EvidenceReference{ID: parsed, Digest: digest}
	}
	declEvidence := addEvidence("urn:gooo:evidence:declaration", false, false)
	deriveEvidence := addEvidence("urn:gooo:evidence:derivation", false, false)
	projectionEvidence := addEvidence("urn:gooo:evidence:projection", false, false)
	verificationEvidence := addEvidence("urn:gooo:evidence:verification", true, true)
	record := func(id string, subject, object semantic.ID, rule string, phase semantic.InferencePhase, ordinal uint64, layer semantic.AuthorityLayer, effect semantic.AuthorityEffect, refs []semantic.EvidenceReference) semantic.InferenceRecord {
		return semantic.InferenceRecord{RecordID: semantic.MustIdentity(id), SubjectID: subject, ObjectID: object, Rule: semantic.RuleBinding{ID: semantic.MustIdentity("urn:gooo:rule:" + rule), Version: "v1", Digest: digestText("rule:" + rule)}, Phase: semantic.PhasePlacement{Phase: phase, Ordinal: ordinal}, Before: snapshotBefore, After: snapshotAfter, Authority: semantic.AuthorityBinding{Layer: layer, Effect: effect}, Evidence: refs, Controls: controls}
	}
	edges := []semantic.InferenceEdge{
		{InferenceRecord: record("urn:gooo:path:declaration", root, owner, "declaration", semantic.PhaseDeclaration, 1, semantic.AuthoritySource, semantic.AuthorityDeclare, []semantic.EvidenceReference{declEvidence}), Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{root}},
		{InferenceRecord: record("urn:gooo:path:derivation", owner, term, "derivation", semantic.PhaseDerivation, 2, semantic.AuthoritySemantic, semantic.AuthorityDerive, []semantic.EvidenceReference{deriveEvidence}), Kind: semantic.InferenceDeterministicDerivation},
	}
	if candidate {
		candidateID := semantic.MustIdentity("urn:gooo:candidate:billing/pay-order")
		candidateRecord := record("urn:gooo:path:candidate", term, candidateID, "candidate", semantic.PhaseObservation, 3, semantic.AuthorityCandidate, semantic.AuthorityObserve, []semantic.EvidenceReference{projectionEvidence})
		candidateRecord.Before.Semantic, candidateRecord.After.Semantic = beforeDigest, beforeDigest
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: candidateRecord, Kind: semantic.InferenceObservationCandidate})
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: record("urn:gooo:path:projection", candidateID, code, "projection", semantic.PhaseProjection, 4, semantic.AuthorityDerived, semantic.AuthorityProject, []semantic.EvidenceReference{projectionEvidence}), Kind: semantic.InferenceDerivedProjection})
	} else {
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: record("urn:gooo:path:projection", term, code, "projection", semantic.PhaseProjection, 3, semantic.AuthorityDerived, semantic.AuthorityProject, []semantic.EvidenceReference{projectionEvidence}), Kind: semantic.InferenceDerivedProjection})
	}
	verification := record("urn:gooo:path:verification", code, receiptID, "verification", semantic.PhaseVerification, 5, semantic.AuthorityVerification, semantic.AuthorityVerify, []semantic.EvidenceReference{verificationEvidence})
	edges = append(edges, semantic.InferenceEdge{InferenceRecord: verification, Kind: semantic.InferenceIndependentVerification})
	claimRecord := record("urn:gooo:claim:billing/pay-order", owner, receiptID, "claim", semantic.PhaseVerification, 6, semantic.AuthoritySemantic, semantic.AuthorityNoChange, []semantic.EvidenceReference{verificationEvidence})
	semanticClaimKind := semantic.NoSemanticDelta
	if claim == ClaimDelta {
		claimRecord.Authority.Effect = semantic.AuthorityDelta
		semanticClaimKind = semantic.SemanticDelta
	}
	claimRecord.Before.Semantic, claimRecord.After.Semantic = beforeDigest, afterDigest
	claimRecord.Before.Source, claimRecord.After.Source = snapshotBefore.Source, snapshotAfter.Source
	claims := []semantic.SemanticChangeClaim{{InferenceRecord: claimRecord, Kind: semanticClaimKind}}
	if claim == ClaimDelta {
		claims[0].CanonicalDelta, claims[0].DeltaDigest = deltaText, digestText(deltaText)
	}
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Claims: claims, Evidence: evidence}
	originPathIDs := []string{"urn:gooo:path:verification", "urn:gooo:path:projection", "urn:gooo:path:derivation", "urn:gooo:path:declaration"}
	if candidate {
		originPathIDs = []string{"urn:gooo:path:verification", "urn:gooo:path:projection", "urn:gooo:path:candidate", "urn:gooo:path:derivation", "urn:gooo:path:declaration"}
	}
	receipt := CouplingReceipt{ReceiptID: receiptID.String(), SurfaceID: binding.RegisteredSurfaceID, SemanticOwnerID: binding.SemanticOwnerID, CodeSymbolID: binding.CodeSymbolID, SourceMapBindingDigest: binding.BindingDigest, SnapshotDigest: snapshotDigest(input, beforeDigest, afterDigest, registry.digest), RegistryDigest: registry.digest, ToolchainDigest: input.Config.ToolchainDigest, ProfileDigest: input.Config.Profile.Digest, BeforeIRDigest: beforeDigest, AfterIRDigest: afterDigest, AuthoritySourceBeforeDigest: sourceDigest(input.AuthoritySourceBefore), AuthoritySourceAfterDigest: sourceDigest(input.AuthoritySourceAfter), ChangeClaim: claim, ReceiptKind: ReceiptNoSemanticDelta, OriginPathIDs: originPathIDs, ClaimRecordID: claimRecord.RecordID.String(), EvidenceRefs: []string{declEvidence.ID.String(), deriveEvidence.ID.String(), projectionEvidence.ID.String(), verificationEvidence.ID.String()}, State: "CURRENT"}
	if claim == ClaimDelta {
		receipt.ReceiptKind = ReceiptSemanticDelta
		receipt.SemanticDelta, receipt.SemanticDeltaDigest, receipt.AuthoritativeSourceRef = deltaText, digestText(deltaText), "gooo://billing/source#authorize-payment"
	}
	return []CouplingReceipt{receipt}, path
}
