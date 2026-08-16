package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type pathWire struct {
	Version  string         `json:"version"`
	Edges    []edgeWire     `json:"edges"`
	Claims   []claimWire    `json:"claims"`
	Evidence []evidenceWire `json:"evidence"`
}

type edgeWire struct {
	Record            recordWire `json:"record"`
	Kind              string     `json:"kind"`
	SourceRoots       []string   `json:"source_roots"`
	AcceptanceReceipt string     `json:"acceptance_receipt"`
}

type claimWire struct {
	Record         recordWire `json:"record"`
	Kind           string     `json:"kind"`
	CanonicalDelta string     `json:"canonical_delta"`
	DeltaDigest    string     `json:"delta_digest"`
}

type recordWire struct {
	RecordID  string            `json:"record_id"`
	SubjectID string            `json:"subject_id"`
	ObjectID  string            `json:"object_id"`
	Rule      ruleWire          `json:"rule"`
	Phase     phaseWire         `json:"phase"`
	Before    snapshotWire      `json:"before"`
	After     snapshotWire      `json:"after"`
	Authority authorityWire     `json:"authority"`
	Evidence  []evidenceRefWire `json:"evidence"`
	Controls  controlsWire      `json:"controls"`
}

type ruleWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type phaseWire struct {
	Phase   string `json:"phase"`
	Ordinal uint64 `json:"ordinal"`
}

type snapshotWire struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}

type authorityWire struct {
	Layer  string `json:"layer"`
	Effect string `json:"effect"`
}

type evidenceRefWire struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

type controlsWire struct {
	CatalogDigest string      `json:"catalog_digest"`
	PolicyDigest  string      `json:"policy_digest"`
	Profile       profileWire `json:"profile"`
}

type profileWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type evidenceWire struct {
	ID           string       `json:"id"`
	Digest       string       `json:"digest"`
	Before       snapshotWire `json:"before"`
	After        snapshotWire `json:"after"`
	SourceBacked bool         `json:"source_backed"`
	Independent  bool         `json:"independent"`
	Controls     controlsWire `json:"controls"`
}

func (path ProvenancePath) MarshalJSON() ([]byte, error) {
	encoded, err := pathWireFromSemantic(path.Path)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		CommandID   string          `json:"command_id"`
		Path        pathWire        `json:"path"`
		Requirement PathRequirement `json:"requirement"`
	}{path.CommandID, encoded, path.Requirement})
}

func (path *ProvenancePath) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	var raw struct {
		CommandID   string          `json:"command_id"`
		Path        pathWire        `json:"path"`
		Requirement PathRequirement `json:"requirement"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("provenance path has trailing data")
	}
	decoded, err := pathWireToSemantic(raw.Path)
	if err != nil {
		return err
	}
	*path = ProvenancePath{CommandID: raw.CommandID, Path: decoded, Requirement: raw.Requirement}
	return nil
}

func pathWireFromSemantic(path semantic.InferencePathV1) (pathWire, error) {
	result := pathWire{Version: path.Version}
	for _, edge := range path.Edges {
		converted, err := edgeWireFromSemantic(edge)
		if err != nil {
			return pathWire{}, err
		}
		result.Edges = append(result.Edges, converted)
	}
	for _, claim := range path.Claims {
		converted, err := claimWireFromSemantic(claim)
		if err != nil {
			return pathWire{}, err
		}
		result.Claims = append(result.Claims, converted)
	}
	for _, evidence := range path.Evidence {
		converted, err := evidenceWireFromSemantic(evidence)
		if err != nil {
			return pathWire{}, err
		}
		result.Evidence = append(result.Evidence, converted)
	}
	return result, nil
}

func pathWireToSemantic(path pathWire) (semantic.InferencePathV1, error) {
	result := semantic.InferencePathV1{Version: path.Version}
	for _, edge := range path.Edges {
		converted, err := edgeWireToSemantic(edge)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Edges = append(result.Edges, converted)
	}
	for _, claim := range path.Claims {
		converted, err := claimWireToSemantic(claim)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Claims = append(result.Claims, converted)
	}
	for _, evidence := range path.Evidence {
		converted, err := evidenceWireToSemantic(evidence)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		result.Evidence = append(result.Evidence, converted)
	}
	return result, nil
}

func edgeWireFromSemantic(edge semantic.InferenceEdge) (edgeWire, error) {
	record, err := recordWireFromSemantic(edge.InferenceRecord)
	if err != nil {
		return edgeWire{}, err
	}
	roots := make([]string, len(edge.SourceRoots))
	for i, root := range edge.SourceRoots {
		roots[i] = root.String()
	}
	return edgeWire{Record: record, Kind: string(edge.Kind), SourceRoots: roots, AcceptanceReceipt: edge.AcceptanceReceipt.String()}, nil
}

func edgeWireToSemantic(edge edgeWire) (semantic.InferenceEdge, error) {
	record, err := recordWireToSemantic(edge.Record)
	if err != nil {
		return semantic.InferenceEdge{}, err
	}
	roots := make([]semantic.ID, len(edge.SourceRoots))
	for i, root := range edge.SourceRoots {
		roots[i], err = semantic.ParseIdentity(root)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	var receipt semantic.ID
	if edge.AcceptanceReceipt != "" {
		receipt, err = semantic.ParseIdentity(edge.AcceptanceReceipt)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	return semantic.InferenceEdge{InferenceRecord: record, Kind: semantic.InferenceKind(edge.Kind), SourceRoots: roots, AcceptanceReceipt: receipt}, nil
}

func claimWireFromSemantic(claim semantic.SemanticChangeClaim) (claimWire, error) {
	record, err := recordWireFromSemantic(claim.InferenceRecord)
	if err != nil {
		return claimWire{}, err
	}
	return claimWire{Record: record, Kind: string(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}, nil
}

func claimWireToSemantic(claim claimWire) (semantic.SemanticChangeClaim, error) {
	record, err := recordWireToSemantic(claim.Record)
	if err != nil {
		return semantic.SemanticChangeClaim{}, err
	}
	return semantic.SemanticChangeClaim{InferenceRecord: record, Kind: semantic.SemanticChangeKind(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}, nil
}

func recordWireFromSemantic(record semantic.InferenceRecord) (recordWire, error) {
	evidence := make([]evidenceRefWire, len(record.Evidence))
	for i, ref := range record.Evidence {
		evidence[i] = evidenceRefWire{ID: ref.ID.String(), Digest: ref.Digest}
	}
	return recordWire{RecordID: record.RecordID.String(), SubjectID: record.SubjectID.String(), ObjectID: record.ObjectID.String(), Rule: ruleWire{ID: record.Rule.ID.String(), Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: phaseWire{Phase: string(record.Phase.Phase), Ordinal: record.Phase.Ordinal}, Before: snapshotWire{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: snapshotWire{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: authorityWire{Layer: string(record.Authority.Layer), Effect: string(record.Authority.Effect)}, Evidence: evidence, Controls: controlsWireFromSemantic(record.Controls)}, nil
}

func recordWireToSemantic(record recordWire) (semantic.InferenceRecord, error) {
	recordID, err := semantic.ParseIdentity(record.RecordID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	subjectID, err := semantic.ParseIdentity(record.SubjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	objectID, err := semantic.ParseIdentity(record.ObjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	ruleID, err := semantic.ParseIdentity(record.Rule.ID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	evidence := make([]semantic.EvidenceReference, len(record.Evidence))
	for i, ref := range record.Evidence {
		id, parseErr := semantic.ParseIdentity(ref.ID)
		if parseErr != nil {
			return semantic.InferenceRecord{}, parseErr
		}
		evidence[i] = semantic.EvidenceReference{ID: id, Digest: ref.Digest}
	}
	return semantic.InferenceRecord{RecordID: recordID, SubjectID: subjectID, ObjectID: objectID, Rule: semantic.RuleBinding{ID: ruleID, Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: semantic.PhasePlacement{Phase: semantic.InferencePhase(record.Phase.Phase), Ordinal: record.Phase.Ordinal}, Before: semantic.SnapshotDigests{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: semantic.SnapshotDigests{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: semantic.AuthorityBinding{Layer: semantic.AuthorityLayer(record.Authority.Layer), Effect: semantic.AuthorityEffect(record.Authority.Effect)}, Evidence: evidence, Controls: controlsWireToSemantic(record.Controls)}, nil
}

func evidenceWireFromSemantic(evidence semantic.InferenceEvidence) (evidenceWire, error) {
	id, err := semantic.ParseIdentity(evidence.ID.String())
	if err != nil {
		return evidenceWire{}, err
	}
	return evidenceWire{ID: id.String(), Digest: evidence.Digest, Before: snapshotWire{Source: evidence.Before.Source, Semantic: evidence.Before.Semantic}, After: snapshotWire{Source: evidence.After.Source, Semantic: evidence.After.Semantic}, SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsWireFromSemantic(evidence.Controls)}, nil
}

func evidenceWireToSemantic(evidence evidenceWire) (semantic.InferenceEvidence, error) {
	id, err := semantic.ParseIdentity(evidence.ID)
	if err != nil {
		return semantic.InferenceEvidence{}, err
	}
	return semantic.InferenceEvidence{ID: id, Digest: evidence.Digest, Before: semantic.SnapshotDigests{Source: evidence.Before.Source, Semantic: evidence.Before.Semantic}, After: semantic.SnapshotDigests{Source: evidence.After.Source, Semantic: evidence.After.Semantic}, SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsWireToSemantic(evidence.Controls)}, nil
}

func controlsWireFromSemantic(controls semantic.InferenceControls) controlsWire {
	return controlsWire{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: profileWire{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}

func controlsWireToSemantic(controls controlsWire) semantic.InferenceControls {
	return semantic.InferenceControls{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: semantic.ProfileBinding{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}
