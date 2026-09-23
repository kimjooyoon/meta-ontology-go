package generation

import (
	"errors"
	"strconv"
	"strings"
)

const (
	ProvenanceTrendObservationSchema = "gooo/provenance-trend-observation/v1"
	ProvenanceTrendImproved          = "IMPROVED"
	ProvenanceTrendRegressed         = "REGRESSED"
	ProvenanceTrendStable            = "STABLE"
	ProvenanceTrendUnknown           = "UNKNOWN"
)

type ProvenanceTrendObservation struct {
	Schema            string
	PreviousDigest    string
	CurrentDigest     string
	PreviousValue     int
	CurrentValue      int
	Direction         string
	Reason            string
	NonAuthorizing    bool
	ObservationDigest string
}

func ObserveProvenanceTrend(previousDigest, currentDigest string, previousValue, currentValue int) ProvenanceTrendObservation {
	observation := ProvenanceTrendObservation{
		Schema:         ProvenanceTrendObservationSchema,
		PreviousDigest: previousDigest,
		CurrentDigest:  currentDigest,
		PreviousValue:  previousValue,
		CurrentValue:   currentValue,
		Direction:      ProvenanceTrendUnknown,
		Reason:         "PROVENANCE_TREND_UNKNOWN",
		NonAuthorizing: true,
	}
	if previousDigest == "" || currentDigest == "" {
		return finalizeProvenanceTrend(observation)
	}
	if !validEnvelopeDigest(previousDigest) || !validEnvelopeDigest(currentDigest) {
		observation.Reason = "MALFORMED_PROVENANCE_EVIDENCE_DIGEST"
		return finalizeProvenanceTrend(observation)
	}
	switch {
	case currentValue > previousValue:
		observation.Direction = ProvenanceTrendImproved
		observation.Reason = "PROVENANCE_METRIC_INCREASED"
	case currentValue < previousValue:
		observation.Direction = ProvenanceTrendRegressed
		observation.Reason = "PROVENANCE_METRIC_DECREASED"
	default:
		observation.Direction = ProvenanceTrendStable
		observation.Reason = "PROVENANCE_METRIC_UNCHANGED"
	}
	return finalizeProvenanceTrend(observation)
}

func finalizeProvenanceTrend(observation ProvenanceTrendObservation) ProvenanceTrendObservation {
	observation.ObservationDigest = envelopeDigestString(observation.Canonical())
	return observation
}

func (observation ProvenanceTrendObservation) Canonical() string {
	return strings.Join([]string{
		ProvenanceTrendObservationSchema,
		observation.PreviousDigest,
		observation.CurrentDigest,
		strconv.Itoa(observation.PreviousValue),
		strconv.Itoa(observation.CurrentValue),
		observation.Direction,
		observation.Reason,
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\t")
}

func (observation ProvenanceTrendObservation) Validate() error {
	if observation.Schema != ProvenanceTrendObservationSchema ||
		!observation.NonAuthorizing ||
		observation.Direction == "" ||
		observation.Reason == "" ||
		!validEnvelopeDigest(observation.ObservationDigest) {
		return errors.New("provenance trend observation identity is invalid")
	}
	if observation.PreviousDigest != "" && !validEnvelopeDigest(observation.PreviousDigest) {
		if observation.Reason != "MALFORMED_PROVENANCE_EVIDENCE_DIGEST" {
			return errors.New("provenance trend previous digest is invalid")
		}
	}
	if observation.CurrentDigest != "" && !validEnvelopeDigest(observation.CurrentDigest) {
		if observation.Reason != "MALFORMED_PROVENANCE_EVIDENCE_DIGEST" {
			return errors.New("provenance trend current digest is invalid")
		}
	}
	expected := ObserveProvenanceTrend(
		observation.PreviousDigest,
		observation.CurrentDigest,
		observation.PreviousValue,
		observation.CurrentValue,
	)
	if observation != expected {
		return errors.New("provenance trend observation does not match evidence")
	}
	return nil
}
