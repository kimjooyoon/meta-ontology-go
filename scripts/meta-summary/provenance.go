package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func decodeProvenance(data []byte) (provenanceEnvelope, error) {
	var envelope provenanceEnvelope
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return envelope, fmt.Errorf("decode artifact provenance: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return envelope, fmt.Errorf("artifact provenance must contain one JSON value")
	}
	if err := validateProvenance(envelope); err != nil {
		return envelope, err
	}
	return envelope, nil
}

func validateProvenance(envelope provenanceEnvelope) error {
	if envelope.SchemaVersion != "gooo/meta-artifact-provenance/v1" {
		return fmt.Errorf("unsupported artifact provenance schema %q", envelope.SchemaVersion)
	}
	if envelope.Decision != "BOUND" || envelope.Reason != "ARTIFACT_PROVENANCE_BOUND" {
		return fmt.Errorf("artifact provenance is not bound")
	}
	if envelope.PromotionAuthorized || envelope.Summary.Fail != 0 || envelope.Summary.Unknown != 0 {
		return fmt.Errorf("artifact provenance is not non-authorizing and closed")
	}
	if envelope.Summary.Pass < 1 || envelope.Summary.Pass != len(envelope.Indicators) {
		return fmt.Errorf("artifact provenance indicator summary is inconsistent")
	}
	if !hexDigest(envelope.BaseSHA, 40) || !hexDigest(envelope.HeadSHA, 40) {
		return fmt.Errorf("artifact provenance commit binding is invalid")
	}
	for _, indicator := range envelope.Indicators {
		if !indicator.valid() {
			return fmt.Errorf("artifact provenance indicator binding is invalid")
		}
	}
	if envelope.IndicatorDecisionLedgerCount < 1 || !ledgerDigest(envelope.IndicatorDecisionLedgerDigest) {
		return fmt.Errorf("artifact provenance ledger binding is invalid")
	}
	for _, value := range []string{envelope.PlanDigest, envelope.ExecutionManifestDigest,
		envelope.ReceiptReportDigest, envelope.InputDigest, envelope.EnvelopeDigest, envelope.ReplayDigest} {
		if !hexDigest(value, 64) {
			return fmt.Errorf("artifact provenance digest is invalid")
		}
	}
	return nil
}

func ledgerDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && hexDigest(strings.TrimPrefix(value, "sha256:"), 64)
}

func hexDigest(value string, size int) bool {
	if len(value) != size {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
