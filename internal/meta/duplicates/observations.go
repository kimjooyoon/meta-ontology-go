package duplicates

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func duplicateObservations(groups map[string][]string) []sourcepolicy.Observation {
	keys := make([]string, 0, len(groups))
	for key, members := range groups {
		if len(members) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	observations := make([]sourcepolicy.Observation, 0, len(keys))
	for _, key := range keys {
		members := groups[key]
		sort.Strings(members)
		observations = append(observations, sourcepolicy.Observation{
			Subject: key, Dimension: sourcepolicy.DimensionRefactorDuplicate, Value: len(members) - 1,
			Detail: "members=" + strings.Join(members, ","), Producer: "duplicates.Analyze",
		})
	}
	return observations
}
