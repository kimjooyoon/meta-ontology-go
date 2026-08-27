package denominatorevolution

import (
	"strings"
)

var requiredEntities = []string{
	"FixedDenominator", "DenominatorVersion", "ChangeReason", "PredecessorEvidence", "MigrationReceipt", "ClaimTransition", "IndependentDecision",
}

var requiredActivities = []string{
	"DeclareFixedDenominator", "ProposeDenominatorChange", "BindPredecessorDigest", "RecordChangeReasons", "IssueMigrationReceipt", "TransitionClaim", "IndependentlyDecide",
}

func projectSource(raw []byte) SourceProjection {
	entities, activities := []string{}, []string{}
	valid := true
	for _, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "package", "namespace":
			if len(fields) != 2 {
				valid = false
			}
		case "entity":
			if len(fields) < 2 {
				valid = false
			} else {
				entities = append(entities, fields[1])
			}
		case "activity":
			if len(fields) < 2 {
				valid = false
			} else {
				activities = append(activities, strings.Split(fields[1], "(")[0])
			}
		default:
			valid = false
		}
	}
	return SourceProjection{EntityCount: len(entities), ActivityCount: len(activities), RequiredEntities: missing(requiredEntities, entities), RequiredActivities: missing(requiredActivities, activities), Exact: valid && len(entities) == len(requiredEntities) && len(activities) == len(requiredActivities) && len(missing(requiredEntities, entities)) == 0 && len(missing(requiredActivities, activities)) == 0}
}

func missing(required, actual []string) []string {
	result := []string{}
	for _, want := range required {
		found := false
		for _, got := range actual {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			result = append(result, want)
		}
	}
	return result
}
