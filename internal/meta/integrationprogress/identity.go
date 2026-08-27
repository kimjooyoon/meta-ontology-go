package integrationprogress

import "time"

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func parseTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}

func observationHeaderValid(value Observation) bool {
	_, timeOK := parseTime(value.ObservedAt)
	return value.Schema == ObservationSchema && value.Repository == Repository &&
		value.CohortID == CohortID && validSHA(value.ObserverHeadSHA) && timeOK
}
