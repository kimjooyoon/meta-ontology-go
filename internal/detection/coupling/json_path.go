package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

type pathWire struct {
	Version  string         `json:"version"`
	Edges    []edgeWire     `json:"edges"`
	Claims   []claimWire    `json:"claims"`
	Evidence []evidenceWire `json:"evidence"`
}

type snapshotWire struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}

type profileWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type controlsWire struct {
	CatalogDigest string      `json:"catalog_digest"`
	PolicyDigest  string      `json:"policy_digest"`
	Profile       profileWire `json:"profile"`
}

type ruleWire struct {
	ID      semantic.ID `json:"id"`
	Version string      `json:"version"`
	Digest  string      `json:"digest"`
}

type phaseWire struct {
	Phase   semantic.InferencePhase `json:"phase"`
	Ordinal uint64                  `json:"ordinal"`
}

type authorityWire struct {
	Layer  semantic.AuthorityLayer  `json:"layer"`
	Effect semantic.AuthorityEffect `json:"effect"`
}

type evidenceRefWire struct {
	ID     semantic.ID `json:"id"`
	Digest string      `json:"digest"`
}

type recordWire struct {
	RecordID  semantic.ID       `json:"record_id"`
	SubjectID semantic.ID       `json:"subject_id"`
	ObjectID  semantic.ID       `json:"object_id"`
	Rule      ruleWire          `json:"rule"`
	Phase     phaseWire         `json:"phase"`
	Before    snapshotWire      `json:"before"`
	After     snapshotWire      `json:"after"`
	Authority authorityWire     `json:"authority"`
	Evidence  []evidenceRefWire `json:"evidence"`
	Controls  controlsWire      `json:"controls"`
}

type edgeWire struct {
	Record            recordWire             `json:"record"`
	Kind              semantic.InferenceKind `json:"kind"`
	SourceRoots       []semantic.ID          `json:"source_roots"`
	AcceptanceReceipt semantic.ID            `json:"acceptance_receipt"`
}

type claimWire struct {
	Record         recordWire                  `json:"record"`
	Kind           semantic.SemanticChangeKind `json:"kind"`
	CanonicalDelta string                      `json:"canonical_delta"`
	DeltaDigest    string                      `json:"delta_digest"`
}

type evidenceWire struct {
	ID           semantic.ID  `json:"id"`
	Digest       string       `json:"digest"`
	Before       snapshotWire `json:"before"`
	After        snapshotWire `json:"after"`
	SourceBacked bool         `json:"source_backed"`
	Independent  bool         `json:"independent"`
	Controls     controlsWire `json:"controls"`
}

func pathToWire(path semantic.InferencePathV1) pathWire {
	out := pathWire{Version: path.Version}
	for _, edge := range path.Edges {
		out.Edges = append(out.Edges, edgeWire{Record: recordToWire(edge.InferenceRecord), Kind: edge.Kind, SourceRoots: edge.SourceRoots, AcceptanceReceipt: edge.AcceptanceReceipt})
	}
	for _, claim := range path.Claims {
		out.Claims = append(out.Claims, claimWire{Record: recordToWire(claim.InferenceRecord), Kind: claim.Kind, CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest})
	}
	for _, evidence := range path.Evidence {
		out.Evidence = append(out.Evidence, evidenceWire{ID: evidence.ID, Digest: evidence.Digest, Before: snapshotToWire(evidence.Before), After: snapshotToWire(evidence.After), SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsToWire(evidence.Controls)})
	}
	return out
}

func pathFromWire(path pathWire) semantic.InferencePathV1 {
	out := semantic.InferencePathV1{Version: path.Version}
	for _, edge := range path.Edges {
		out.Edges = append(out.Edges, semantic.InferenceEdge{InferenceRecord: recordFromWire(edge.Record), Kind: edge.Kind, SourceRoots: edge.SourceRoots, AcceptanceReceipt: edge.AcceptanceReceipt})
	}
	for _, claim := range path.Claims {
		out.Claims = append(out.Claims, semantic.SemanticChangeClaim{InferenceRecord: recordFromWire(claim.Record), Kind: claim.Kind, CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest})
	}
	for _, evidence := range path.Evidence {
		out.Evidence = append(out.Evidence, semantic.InferenceEvidence{ID: evidence.ID, Digest: evidence.Digest, Before: snapshotFromWire(evidence.Before), After: snapshotFromWire(evidence.After), SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsFromWire(evidence.Controls)})
	}
	return out
}

func recordToWire(record semantic.InferenceRecord) recordWire {
	refs := make([]evidenceRefWire, 0, len(record.Evidence))
	for _, ref := range record.Evidence {
		refs = append(refs, evidenceRefWire{ID: ref.ID, Digest: ref.Digest})
	}
	return recordWire{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: ruleWire{ID: record.Rule.ID, Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: phaseWire{Phase: record.Phase.Phase, Ordinal: record.Phase.Ordinal}, Before: snapshotToWire(record.Before), After: snapshotToWire(record.After), Authority: authorityWire{Layer: record.Authority.Layer, Effect: record.Authority.Effect}, Evidence: refs, Controls: controlsToWire(record.Controls)}
}

func recordFromWire(record recordWire) semantic.InferenceRecord {
	refs := make([]semantic.EvidenceReference, 0, len(record.Evidence))
	for _, ref := range record.Evidence {
		refs = append(refs, semantic.EvidenceReference{ID: ref.ID, Digest: ref.Digest})
	}
	return semantic.InferenceRecord{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: semantic.RuleBinding{ID: record.Rule.ID, Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: semantic.PhasePlacement{Phase: record.Phase.Phase, Ordinal: record.Phase.Ordinal}, Before: snapshotFromWire(record.Before), After: snapshotFromWire(record.After), Authority: semantic.AuthorityBinding{Layer: record.Authority.Layer, Effect: record.Authority.Effect}, Evidence: refs, Controls: controlsFromWire(record.Controls)}
}

func snapshotToWire(snapshot semantic.SnapshotDigests) snapshotWire {
	return snapshotWire{Source: snapshot.Source, Semantic: snapshot.Semantic}
}

func snapshotFromWire(snapshot snapshotWire) semantic.SnapshotDigests {
	return semantic.SnapshotDigests{Source: snapshot.Source, Semantic: snapshot.Semantic}
}

func controlsToWire(controls semantic.InferenceControls) controlsWire {
	return controlsWire{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: profileWire{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}

func controlsFromWire(controls controlsWire) semantic.InferenceControls {
	return semantic.InferenceControls{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: semantic.ProfileBinding{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}
