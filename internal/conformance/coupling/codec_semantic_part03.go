package coupling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"io"
)

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
