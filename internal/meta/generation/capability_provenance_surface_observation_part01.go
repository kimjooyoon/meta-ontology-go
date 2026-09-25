package generation

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const CapabilityProvenanceSurfaceObservationSchema = "gooo/capability-provenance-surface-observation/v1"

const (
	CapabilityProvenanceSurfaceClosed  = "CLOSED"
	CapabilityProvenanceSurfaceUnknown = "UNKNOWN"
	CapabilityProvenanceSurfaceRefuted = "REFUTED"
)

type CapabilityProvenanceSurfaceObservationInput struct {
	SourceDigest             string
	SemanticDigest           string
	GeneratedDigest          string
	ReverseObservationDigest string
	DeltaSeries              CapabilityProvenanceDeltaSeries
}

type CapabilityProvenanceSurfaceMetrics struct {
	DeltaCount    int    `json:"delta_count"`
	ChangeCount   int    `json:"change_count"`
	ClosedDeltas  int    `json:"closed_deltas"`
	UnknownDeltas int    `json:"unknown_deltas"`
	RefutedDeltas int    `json:"refuted_deltas"`
	FinalVerdict  string `json:"final_verdict"`
}

type CapabilityProvenanceSurfaceObservation struct {
	Schema                   string                             `json:"schema"`
	Decision                 string                             `json:"decision"`
	Reason                   string                             `json:"reason"`
	SourceDigest             string                             `json:"source_digest,omitempty"`
	SemanticDigest           string                             `json:"semantic_digest,omitempty"`
	GeneratedDigest          string                             `json:"generated_digest,omitempty"`
	ReverseObservationDigest string                             `json:"reverse_observation_digest,omitempty"`
	DeltaSeriesDigest        string                             `json:"delta_series_digest,omitempty"`
	Metrics                  CapabilityProvenanceSurfaceMetrics `json:"metrics"`
	NonAuthorizing           bool                               `json:"non_authorizing"`
	ObservationDigest        string                             `json:"observation_digest"`
}

// ObserveCapabilityProvenanceSurface binds the source, lowered IR, generated
// artifact, reverse observation, and ordered delta history. It reports only
// evidence binding; it never claims semantic improvement or grants authority.
func ObserveCapabilityProvenanceSurface(input CapabilityProvenanceSurfaceObservationInput) CapabilityProvenanceSurfaceObservation {
	observation := CapabilityProvenanceSurfaceObservation{
		Schema: CapabilityProvenanceSurfaceObservationSchema, Decision: CapabilityProvenanceSurfaceUnknown,
		Reason: "PROVENANCE_SURFACE_BINDING_UNKNOWN", NonAuthorizing: true,
		SourceDigest: input.SourceDigest, SemanticDigest: input.SemanticDigest,
		GeneratedDigest: input.GeneratedDigest, ReverseObservationDigest: input.ReverseObservationDigest,
		DeltaSeriesDigest: input.DeltaSeries.SeriesDigest,
	}
	ordered := []struct {
		name  string
		value string
	}{
		{name: "source_digest", value: input.SourceDigest},
		{name: "semantic_digest", value: input.SemanticDigest},
		{name: "generated_digest", value: input.GeneratedDigest},
		{name: "reverse_observation_digest", value: input.ReverseObservationDigest},
	}
	for _, binding := range ordered {
		if binding.value == "" {
			observation.Reason = "MISSING_" + strings.ToUpper(binding.name)
			return finalizeCapabilityProvenanceSurfaceObservation(observation)
		}
		if !cache.Digest(binding.value).Known() {
			observation.Decision = CapabilityProvenanceSurfaceRefuted
			observation.Reason = "MALFORMED_" + strings.ToUpper(binding.name)
			return finalizeCapabilityProvenanceSurfaceObservation(observation)
		}
	}
	if input.DeltaSeries.Schema == "" || len(input.DeltaSeries.Deltas) == 0 {
		observation.Reason = "MISSING_DELTA_SERIES"
		return finalizeCapabilityProvenanceSurfaceObservation(observation)
	}
	if err := input.DeltaSeries.Validate(); err != nil {
		observation.Decision = CapabilityProvenanceSurfaceRefuted
		observation.Reason = "INVALID_DELTA_SERIES"
		return finalizeCapabilityProvenanceSurfaceObservation(observation)
	}
	observation.Metrics = capabilityProvenanceSurfaceMetrics(input.DeltaSeries)
	observation.Decision = CapabilityProvenanceSurfaceClosed
	observation.Reason = "PROVENANCE_SURFACE_BOUND"
	return finalizeCapabilityProvenanceSurfaceObservation(observation)
}

func capabilityProvenanceSurfaceMetrics(series CapabilityProvenanceDeltaSeries) CapabilityProvenanceSurfaceMetrics {
	metrics := CapabilityProvenanceSurfaceMetrics{DeltaCount: len(series.Deltas), FinalVerdict: series.Verdict}
	for _, delta := range series.Deltas {
		metrics.ChangeCount += len(delta.Changes)
		switch delta.Verdict {
		case CapabilityProvenanceSurfaceClosed:
			metrics.ClosedDeltas++
		case CapabilityProvenanceSurfaceRefuted, "REJECTED":
			metrics.RefutedDeltas++
		default:
			metrics.UnknownDeltas++
		}
	}
	return metrics
}

func finalizeCapabilityProvenanceSurfaceObservation(observation CapabilityProvenanceSurfaceObservation) CapabilityProvenanceSurfaceObservation {
	observation.ObservationDigest = cache.HashBytes([]byte(observation.Canonical())).String()
	return observation
}

func (observation CapabilityProvenanceSurfaceObservation) Canonical() string {
	return strings.Join([]string{
		CapabilityProvenanceSurfaceObservationSchema, observation.Decision, observation.Reason,
		observation.SourceDigest, observation.SemanticDigest, observation.GeneratedDigest,
		observation.ReverseObservationDigest, observation.DeltaSeriesDigest,
		strconv.Itoa(observation.Metrics.DeltaCount), strconv.Itoa(observation.Metrics.ChangeCount),
		strconv.Itoa(observation.Metrics.ClosedDeltas), strconv.Itoa(observation.Metrics.UnknownDeltas),
		strconv.Itoa(observation.Metrics.RefutedDeltas), observation.Metrics.FinalVerdict,
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\x1f")
}

func (observation CapabilityProvenanceSurfaceObservation) Validate() error {
	if observation.Schema != CapabilityProvenanceSurfaceObservationSchema || !observation.NonAuthorizing || observation.Decision == "" || observation.Reason == "" {
		return errors.New("capability provenance surface observation identity is invalid")
	}
	for _, value := range []string{observation.SourceDigest, observation.SemanticDigest, observation.GeneratedDigest, observation.ReverseObservationDigest, observation.DeltaSeriesDigest, observation.ObservationDigest} {
		if value != "" && !cache.Digest(value).Known() {
			return errors.New("capability provenance surface observation contains an invalid digest")
		}
	}
	if observation.ObservationDigest != cache.HashBytes([]byte(observation.Canonical())).String() {
		return errors.New("capability provenance surface observation digest mismatch")
	}
	if observation.Decision == CapabilityProvenanceSurfaceClosed &&
		(observation.SourceDigest == "" || observation.SemanticDigest == "" || observation.GeneratedDigest == "" || observation.ReverseObservationDigest == "" || observation.DeltaSeriesDigest == "") {
		return errors.New("closed capability provenance surface observation is incomplete")
	}
	return nil
}
