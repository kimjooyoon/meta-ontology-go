package lanefrontier

import (
	"sort"
)

func normalizedInput(input Input) Input {
	input.OwnedPathPrefixes, _ = canonicalPrefixes(input.OwnedPathPrefixes)
	input.ChangedPaths = append([]string(nil), input.ChangedPaths...)
	sort.Strings(input.ChangedPaths)
	return input
}
