package duplicates

import (
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func duplicateObservations(groups map[string][]member) []sourcepolicy.Observation {
	keys := make([]string, 0, len(groups))
	for key, members := range groups {
		if len(members) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	observations := make([]sourcepolicy.Observation, 0, len(keys))
	for _, key := range keys {
		members := make([]string, 0, len(groups[key]))
		paths := make([]string, 0, len(groups[key]))
		for _, member := range groups[key] {
			members = append(members, member.subject)
			paths = append(paths, member.path)
		}
		sort.Strings(members)
		sort.Strings(paths)
		observations = append(observations, sourcepolicy.Observation{
			Subject: key, Dimension: sourcepolicy.DimensionRefactorDuplicate, Value: len(members) - 1,
			Detail: "members=" + strings.Join(members, ","), Producer: "duplicates.Analyze",
			SourcePaths: slices.Compact(paths),
		})
	}
	return observations
}
