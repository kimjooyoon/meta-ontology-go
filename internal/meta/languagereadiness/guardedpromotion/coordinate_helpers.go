package guardedpromotion

import "regexp"

const (
	statusSatisfied    = "SATISFIED"
	statusNotSatisfied = "NOT_SATISFIED"
	statusUnresolved   = "UNRESOLVED"
)

var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func resolvedCoordinate(id, choice string, known, satisfied bool) Coordinate {
	coordinate := Coordinate{ID: id, ProofChoice: choice}
	if !known {
		coordinate.Status = statusUnresolved
		coordinate.Reason = "COORDINATE_EVIDENCE_UNKNOWN"
		return coordinate
	}
	if satisfied {
		coordinate.Status = statusSatisfied
		coordinate.Reason = "COORDINATE_EXACTLY_PROVEN"
		return coordinate
	}
	coordinate.Status = statusNotSatisfied
	coordinate.Reason = "COORDINATE_GUARD_NOT_MET"
	return coordinate
}

func validSHA(value string) bool {
	return shaPattern.MatchString(value)
}

func validDigest(value string) bool {
	return digestPattern.MatchString(value)
}

func expectedCIName(event string) (string, bool) {
	switch event {
	case "push":
		return "CI [push full]", true
	case "pull_request":
		return "CI [PR authoritative]", true
	default:
		return "", false
	}
}
