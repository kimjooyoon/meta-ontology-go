package bidir

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

func validateDeltaConsistency(delta BXDeltaEvidence) error {
	if err := validateDeltaHashes(delta); err != nil {
		return err
	}
	return validateCanonicalDelta(delta)
}

func validateDeltaHashes(delta BXDeltaEvidence) error {
	for name, value := range map[string]string{
		"sequence": delta.SequenceHash, "order": delta.OrderHash,
		"locality closure": delta.LocalityClosureHash, "port order": delta.PortOrderHash,
		"relation order": delta.RelationOrderHash, "evidence": delta.EvidenceHash,
	} {
		if !isSHA256(value) {
			return fmt.Errorf("%s hash is not a SHA-256 digest", name)
		}
	}
	if delta.LocalityClosureHash != localityDigest(delta.Locality) {
		return errors.New("locality closure hash does not match locality")
	}
	if delta.PortOrderHash != sequenceHash(delta.PortSequence) || delta.RelationOrderHash != sequenceHash(delta.RelationSequence) {
		return errors.New("ordered port/relation hash does not match sequence")
	}
	if delta.OrderHash != deltaOrderHash(delta.SequenceHash, delta.PortOrderHash, delta.RelationOrderHash) {
		return errors.New("order hash does not match sequence and ordered collections")
	}
	if delta.EvidenceHash != evidenceSpanSetHash(delta.EvidenceSpans) {
		return errors.New("evidence hash does not match IDs, fact keys, and spans")
	}
	return nil
}

func validateCanonicalDelta(delta BXDeltaEvidence) error {
	var payload canonicalDeltaEvidence
	decoder := json.NewDecoder(bytes.NewReader([]byte(delta.CanonicalJSON)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return fmt.Errorf("canonical delta JSON is invalid: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("canonical delta JSON has trailing data")
	}
	expected := canonicalDeltaEvidence{
		SequenceHash: delta.SequenceHash, OrderHash: delta.OrderHash,
		Added: delta.Added, Removed: delta.Removed, Candidates: delta.Candidates,
		PortSequence: delta.PortSequence, RelationSequence: delta.RelationSequence,
		Touched: delta.Locality.Touched, Affected: delta.Locality.Affected,
		ClosureMembers: delta.ClosureMembers, ClosureHash: delta.LocalityClosureHash,
		EvidenceIDs: delta.EvidenceSpans.IDs, EvidenceFactKeys: delta.EvidenceSpans.FactKeys,
		EvidenceSpans: spanTexts(delta.EvidenceSpans.Spans), EvidenceRecords: canonicalEvidenceRecords(delta.EvidenceSpans.Records), EvidenceIDCount: delta.EvidenceSpans.IDCount, EvidenceSpanCount: delta.EvidenceSpans.SpanCount,
		EvidenceIDAuthority: delta.EvidenceSpans.EvidenceIDAuthority, EvidenceHash: delta.EvidenceHash,
		Partial: delta.PartialObservation,
	}
	if !reflect.DeepEqual(payload, expected) {
		return errors.New("canonical delta JSON does not match evidence fields")
	}
	if got := delta.LocalityCanonicalJSON; got != localityJSON(delta.Locality, delta.LocalityClosureHash) {
		return errors.New("canonical locality JSON does not match locality evidence")
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != hex.EncodedLen(sha256.Size) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
