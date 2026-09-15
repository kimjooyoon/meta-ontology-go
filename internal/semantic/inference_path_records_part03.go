package semantic

import (
	"errors"
	"fmt"
)

func (r InferenceRecord) normalized() (InferenceRecord, error) {
	var out InferenceRecord
	var err error
	if out.RecordID, err = ParseIdentity(r.RecordID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("record ID: %w", err)
	}
	if out.SubjectID, err = ParseIdentity(r.SubjectID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("subject ID: %w", err)
	}
	if out.ObjectID, err = ParseIdentity(r.ObjectID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("object ID: %w", err)
	}
	if out.Rule, err = r.Rule.normalized(); err != nil {
		return InferenceRecord{}, err
	}
	out.Phase = r.Phase
	if !out.Phase.Phase.Valid() || out.Phase.Ordinal == 0 {
		return InferenceRecord{}, errors.New("phase and non-zero phase ordinal are required")
	}
	if out.Before, err = r.Before.normalized(); err != nil {
		return InferenceRecord{}, fmt.Errorf("before: %w", err)
	}
	if out.After, err = r.After.normalized(); err != nil {
		return InferenceRecord{}, fmt.Errorf("after: %w", err)
	}
	out.Authority = r.Authority
	if !out.Authority.Layer.Valid() || !out.Authority.Effect.Valid() {
		return InferenceRecord{}, errors.New("unknown authority layer or effect")
	}
	if out.Evidence, err = normalizeEvidenceReferences(r.Evidence); err != nil {
		return InferenceRecord{}, err
	}
	if out.Controls, err = r.Controls.normalized(); err != nil {
		return InferenceRecord{}, err
	}
	return out, nil
}
func (e InferenceEvidence) normalized() (InferenceEvidence, error) {
	id, err := ParseIdentity(e.ID.String())
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence record ID: %w", err)
	}
	digest, err := normalizeDigest(e.Digest)
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence record digest: %w", err)
	}
	before, err := e.Before.normalized()
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence before: %w", err)
	}
	after, err := e.After.normalized()
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence after: %w", err)
	}
	controls, err := e.Controls.normalized()
	if err != nil {
		return InferenceEvidence{}, err
	}
	return InferenceEvidence{
		ID: id, Digest: digest, Before: before, After: after,
		SourceBacked: e.SourceBacked, Independent: e.Independent, Controls: controls,
	}, nil
}
func (e InferenceEvidence) Normalized() (InferenceEvidence, error) { return e.normalized() }
func (e InferenceEvidence) Validate() error                        { _, err := e.normalized(); return err }
