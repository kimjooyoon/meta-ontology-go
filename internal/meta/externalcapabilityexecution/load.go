package externalcapabilityexecution

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadObservation(path string) Observation {
	raw, err := os.ReadFile(path)
	if err != nil {
		return unavailable(Observation{Schema: ObservationSchema}, err)
	}
	var observation Observation
	if err := json.Unmarshal(raw, &observation); err != nil {
		return unavailable(Observation{Schema: ObservationSchema}, err)
	}
	sealed := sealObservation(observation)
	if sealed.ObservationDigest != observation.ObservationDigest {
		return unavailable(observation, fmt.Errorf("observation digest mismatch"))
	}
	return observation
}
