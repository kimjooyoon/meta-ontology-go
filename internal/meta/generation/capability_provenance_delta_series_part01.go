package generation

import (
	"errors"
	"strconv"
	"strings"
)

const CapabilityProvenanceDeltaSeriesSchema = "gooo/capability-provenance-delta-series/v1"

type CapabilityProvenanceDeltaSeries struct {
	Schema           string                      `json:"schema"`
	Deltas           []CapabilityProvenanceDelta `json:"deltas"`
	FinalDelta       CapabilityProvenanceDelta   `json:"final_delta"`
	FirstChainDigest string                      `json:"first_chain_digest"`
	LastChainDigest  string                      `json:"last_chain_digest"`
	Verdict          string                      `json:"verdict"`
	NonAuthorizing   bool                        `json:"non_authorizing"`
	SeriesDigest     string                      `json:"series_digest"`
}

// FoldCapabilityProvenanceDeltas preserves an ordered observation history and
// derives one final non-authorizing witness from its adjacent transitions.
func FoldCapabilityProvenanceDeltas(deltas []CapabilityProvenanceDelta) (CapabilityProvenanceDeltaSeries, error) {
	if len(deltas) == 0 {
		return CapabilityProvenanceDeltaSeries{}, errors.New("capability provenance delta series cannot be empty")
	}
	for index, delta := range deltas {
		if err := validateCapabilityProvenanceDeltaWitness(delta); err != nil {
			return CapabilityProvenanceDeltaSeries{}, errors.New("delta " + strconv.Itoa(index) + ": " + err.Error())
		}
	}

	finalDelta := deltas[0]
	for index := 1; index < len(deltas); index++ {
		composed, err := ComposeCapabilityProvenanceDelta(finalDelta, deltas[index])
		if err != nil {
			return CapabilityProvenanceDeltaSeries{}, errors.New("compose delta " + strconv.Itoa(index) + ": " + err.Error())
		}
		finalDelta = composed
	}

	series := CapabilityProvenanceDeltaSeries{
		Schema:           CapabilityProvenanceDeltaSeriesSchema,
		Deltas:           append([]CapabilityProvenanceDelta(nil), deltas...),
		FinalDelta:       finalDelta,
		FirstChainDigest: deltas[0].PreviousChainDigest,
		LastChainDigest:  finalDelta.CurrentChainDigest,
		Verdict:          finalDelta.Verdict,
		NonAuthorizing:   true,
	}
	series.SeriesDigest = series.StableHash()
	return series, nil
}

func (s CapabilityProvenanceDeltaSeries) Validate() error {
	if s.Schema != CapabilityProvenanceDeltaSeriesSchema || !s.NonAuthorizing {
		return errors.New("capability provenance delta series is not a non-authorizing witness")
	}
	expected, err := FoldCapabilityProvenanceDeltas(s.Deltas)
	if err != nil {
		return err
	}
	if !capabilityProvenanceDeltaSeriesDeltasEqual(s.Deltas, expected.Deltas) ||
		s.FinalDelta.Canonical() != expected.FinalDelta.Canonical() ||
		s.FinalDelta.DeltaDigest != expected.FinalDelta.DeltaDigest ||
		s.FirstChainDigest != expected.FirstChainDigest ||
		s.LastChainDigest != expected.LastChainDigest ||
		s.Verdict != expected.Verdict ||
		s.SeriesDigest != expected.SeriesDigest {
		return errors.New("capability provenance delta series does not match exact observation history")
	}
	return nil
}

func (s CapabilityProvenanceDeltaSeries) Canonical() string {
	deltaDigests := make([]string, 0, len(s.Deltas))
	for _, delta := range s.Deltas {
		deltaDigests = append(deltaDigests, delta.DeltaDigest)
	}
	return strings.Join([]string{
		"capability-provenance-delta-series",
		s.Schema,
		strings.Join(deltaDigests, ""),
		s.FinalDelta.DeltaDigest,
		s.FirstChainDigest,
		s.LastChainDigest,
		s.Verdict,
		boolString(s.NonAuthorizing),
	}, "	")
}

func (s CapabilityProvenanceDeltaSeries) StableHash() string {
	return envelopeDigestString(s.Canonical())
}

func capabilityProvenanceDeltaSeriesDeltasEqual(left, right []CapabilityProvenanceDelta) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Canonical() != right[index].Canonical() ||
			left[index].DeltaDigest != right[index].DeltaDigest {
			return false
		}
	}
	return true
}
