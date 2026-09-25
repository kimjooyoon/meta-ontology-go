package workfrontier

import (
	"strconv"
	"strings"
)

func sortedUnique(values []string) []string {
	values = sortedCopy(values)
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
func stateKey(state ObligationState) string {
	return state.obligationID() + "\x00" + state.Status
}
func pathKey(path RepairPath) string {
	parts := []string{
		path.stableID(), path.ObligationID,
		strings.Join(path.PrerequisiteObligationIDs, "\x00"),
		strings.Join(path.ReadSet, "\x00"), strings.Join(path.WriteSet, "\x00"),
		strings.Join(path.RequiredPressureIDs, "\x00"),
		strconv.FormatUint(uint64(path.PolicyPriority), 10),
		strconv.FormatUint(path.CPUCoreNSUpperBound, 10),
	}
	return strings.Join(parts, "\x00")
}
