package coupling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func cloneSemanticIR(input SemanticIR) SemanticIR {
	input.Nodes = append([]SemanticNode(nil), input.Nodes...)
	input.Relations = append([]SemanticRelation(nil), input.Relations...)
	for i := range input.Nodes {
		input.Nodes[i].Aliases = append([]string(nil), input.Nodes[i].Aliases...)
	}
	return input
}

func normalizeSemanticIR(input *SemanticIR) {
	sort.Slice(input.Nodes, func(i, j int) bool { return input.Nodes[i].ID < input.Nodes[j].ID })
	sort.Slice(input.Relations, func(i, j int) bool {
		left := input.Relations[i].Subject + "\x00" + input.Relations[i].Predicate + "\x00" + input.Relations[i].Object
		right := input.Relations[j].Subject + "\x00" + input.Relations[j].Predicate + "\x00" + input.Relations[j].Object
		return left < right
	})
	for i := range input.Nodes {
		sort.Strings(input.Nodes[i].Aliases)
	}
}

func pathToWire(path semantic.InferencePathV1) wirePath {
	out := wirePath{Version: path.Version, Edges: make([]wireEdge, 0, len(path.Edges)), Claims: make([]wireClaim, 0, len(path.Claims)), Evidence: make([]wireEvidence, 0, len(path.Evidence))}
	for _, edge := range path.Edges {
		out.Edges = append(out.Edges, wireEdgeFromSemantic(edge))
	}
	for _, claim := range path.Claims {
		out.Claims = append(out.Claims, wireClaimFromSemantic(claim))
	}
	for _, evidence := range path.Evidence {
		out.Evidence = append(out.Evidence, wireEvidenceFromSemantic(evidence))
	}
	return out
}

func pathFromWire(raw wirePath) (semantic.InferencePathV1, error) {
	out := semantic.InferencePathV1{Version: raw.Version, Edges: make([]semantic.InferenceEdge, 0, len(raw.Edges)), Claims: make([]semantic.SemanticChangeClaim, 0, len(raw.Claims)), Evidence: make([]semantic.InferenceEvidence, 0, len(raw.Evidence))}
	for _, edge := range raw.Edges {
		value, err := semanticEdgeFromWire(edge)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Edges = append(out.Edges, value)
	}
	for _, claim := range raw.Claims {
		value, err := semanticClaimFromWire(claim)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Claims = append(out.Claims, value)
	}
	for _, evidence := range raw.Evidence {
		value, err := semanticEvidenceFromWire(evidence)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Evidence = append(out.Evidence, value)
	}
	return out, nil
}

func wireEdgeFromSemantic(edge semantic.InferenceEdge) wireEdge {
	record := wireRecordFromSemantic(edge.InferenceRecord)
	return wireEdge{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: edge.Kind.String(), SourceRoots: idsToStrings(edge.SourceRoots), AcceptanceReceipt: edge.AcceptanceReceipt.String()}
}

func wireClaimFromSemantic(claim semantic.SemanticChangeClaim) wireClaim {
	record := wireRecordFromSemantic(claim.InferenceRecord)
	return wireClaim{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: claim.Kind.String(), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}
}

func wireRecordFromSemantic(record semantic.InferenceRecord) wireRecord {
	return wireRecord{RecordID: record.RecordID.String(), SubjectID: record.SubjectID.String(), ObjectID: record.ObjectID.String(), Rule: wireRule{ID: record.Rule.ID.String(), Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: record.Phase.Phase.String(), Ordinal: record.Phase.Ordinal, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: wireAuthority{Layer: record.Authority.Layer.String(), Effect: record.Authority.Effect.String()}, Evidence: evidenceRefsToWire(record.Evidence), Controls: wireControlsFromSemantic(record.Controls)}
}

func wireEvidenceFromSemantic(record semantic.InferenceEvidence) wireEvidence {
	return wireEvidence{ID: record.ID.String(), Digest: record.Digest, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, SourceBacked: record.SourceBacked, Independent: record.Independent, Controls: wireControlsFromSemantic(record.Controls)}
}

func wireControlsFromSemantic(value semantic.InferenceControls) wireControls {
	return wireControls{CatalogDigest: value.CatalogDigest, PolicyDigest: value.PolicyDigest, Profile: wireProfile{ID: value.Profile.ID, Version: value.Profile.Version, Digest: value.Profile.Digest}}
}

func evidenceRefsToWire(refs []semantic.EvidenceReference) []wireEvidenceRef {
	out := make([]wireEvidenceRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, wireEvidenceRef{ID: ref.ID.String(), Digest: ref.Digest})
	}
	return out
}

func idsToStrings(ids []semantic.ID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func semanticEdgeFromWire(raw wireEdge) (semantic.InferenceEdge, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.InferenceEdge{}, err
	}
	kind := semantic.InferenceKind(raw.Kind)
	if !kind.Valid() {
		return semantic.InferenceEdge{}, fmt.Errorf("unknown inference kind %q", raw.Kind)
	}
	roots := make([]semantic.ID, 0, len(raw.SourceRoots))
	for _, root := range raw.SourceRoots {
		id, err := semantic.ParseIdentity(root)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
		roots = append(roots, id)
	}
	var receipt semantic.ID
	if raw.AcceptanceReceipt != "" {
		receipt, err = semantic.ParseIdentity(raw.AcceptanceReceipt)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	return semantic.InferenceEdge{InferenceRecord: record, Kind: kind, SourceRoots: roots, AcceptanceReceipt: receipt}, nil
}

func semanticClaimFromWire(raw wireClaim) (semantic.SemanticChangeClaim, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.SemanticChangeClaim{}, err
	}
	return semantic.SemanticChangeClaim{InferenceRecord: record, Kind: semantic.SemanticChangeKind(raw.Kind), CanonicalDelta: raw.CanonicalDelta, DeltaDigest: raw.DeltaDigest}, nil
}

func semanticEvidenceFromWire(raw wireEvidence) (semantic.InferenceEvidence, error) {
	id, err := semantic.ParseIdentity(raw.ID)
	if err != nil {
		return semantic.InferenceEvidence{}, err
	}
	return semantic.InferenceEvidence{ID: id, Digest: raw.Digest, Before: semantic.SnapshotDigests{Source: raw.Before.Source, Semantic: raw.Before.Semantic}, After: semantic.SnapshotDigests{Source: raw.After.Source, Semantic: raw.After.Semantic}, SourceBacked: raw.SourceBacked, Independent: raw.Independent, Controls: semanticControlsFromWire(raw.Controls)}, nil
}

func semanticRecordFromWire(raw wireRecord) (semantic.InferenceRecord, error) {
	recordID, err := semantic.ParseIdentity(raw.RecordID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	subjectID, err := semantic.ParseIdentity(raw.SubjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	objectID, err := semantic.ParseIdentity(raw.ObjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	ruleID, err := semantic.ParseIdentity(raw.Rule.ID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	evidence := make([]semantic.EvidenceReference, 0, len(raw.Evidence))
	for _, ref := range raw.Evidence {
		id, parseErr := semantic.ParseIdentity(ref.ID)
		if parseErr != nil {
			return semantic.InferenceRecord{}, parseErr
		}
		evidence = append(evidence, semantic.EvidenceReference{ID: id, Digest: ref.Digest})
	}
	return semantic.InferenceRecord{RecordID: recordID, SubjectID: subjectID, ObjectID: objectID, Rule: semantic.RuleBinding{ID: ruleID, Version: raw.Rule.Version, Digest: raw.Rule.Digest}, Phase: semantic.PhasePlacement{Phase: semantic.InferencePhase(raw.Phase), Ordinal: raw.Ordinal}, Before: semantic.SnapshotDigests{Source: raw.Before.Source, Semantic: raw.Before.Semantic}, After: semantic.SnapshotDigests{Source: raw.After.Source, Semantic: raw.After.Semantic}, Authority: semantic.AuthorityBinding{Layer: semantic.AuthorityLayer(raw.Authority.Layer), Effect: semantic.AuthorityEffect(raw.Authority.Effect)}, Evidence: evidence, Controls: semanticControlsFromWire(raw.Controls)}, nil
}

func semanticControlsFromWire(raw wireControls) semantic.InferenceControls {
	return semantic.InferenceControls{CatalogDigest: raw.CatalogDigest, PolicyDigest: raw.PolicyDigest, Profile: semantic.ProfileBinding{ID: raw.Profile.ID, Version: raw.Profile.Version, Digest: raw.Profile.Digest}}
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate JSON key %q", key)
		}
		seen[key] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	out := copyValues[:1]
	for _, value := range copyValues[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
