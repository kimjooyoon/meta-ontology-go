package workfrontier

func validClaims(claims []oracleClaim, registered map[string]bool) bool {
	if len(claims) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, claim := range claims {
		if claim.Path == "" || !registered[claim.Path] || (claim.Mode != "R" && claim.Mode != "W") || seen[claim.Path] {
			return false
		}
		seen[claim.Path] = true
	}
	return true
}
func conflictsWithAny(candidate oraclePressure, selected []oraclePressure) bool {
	for _, prior := range selected {
		if pressuresConflict(candidate, prior) {
			return true
		}
	}
	return false
}
func pressuresConflict(left, right oraclePressure) bool {
	for _, leftClaim := range left.Claims {
		for _, rightClaim := range right.Claims {
			if leftClaim.Path == rightClaim.Path && (leftClaim.Mode == "W" || rightClaim.Mode == "W") {
				return true
			}
		}
	}
	return false
}
func maximumCompatibleSize(pressures []oraclePressure, capacity int) int {
	maximum := 0
	for mask := 1; mask < 1<<len(pressures); mask++ {
		var chosen []oraclePressure
		cpu, valid := 0, true
		for i, pressure := range pressures {
			if mask&(1<<i) == 0 {
				continue
			}
			if cpu+pressure.CPU > capacity || conflictsWithAny(pressure, chosen) {
				valid = false
				break
			}
			cpu += pressure.CPU
			chosen = append(chosen, pressure)
		}
		if valid && len(chosen) > maximum {
			maximum = len(chosen)
		}
	}
	return maximum
}
