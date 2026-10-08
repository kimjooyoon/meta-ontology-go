package bodycodegen

import (
	"context"
	"fmt"
)

// ReviseAssemblySearch applies one source-declared alternative atomically with
// its attempt budget. Bodies, intent, cases and the declared alternatives persist.
func ReviseAssemblySearch(ctx context.Context, filename string, source []byte, activity, alternativeID string, attempts int) ([]byte, error) {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return nil, err
	}
	if !IsSourceIRSearch(spec) || attempts < 1 || attempts > 64 {
		return nil, fmt.Errorf("search revision requires a source search and bounded attempts")
	}
	for _, alternative := range spec.SearchAlternatives {
		if alternative.ID != alternativeID {
			continue
		}
		if attempts > alternative.MaxCandidates ||
			alternative.Grammar == spec.Search.Grammar && alternative.MaxCandidates == spec.Search.MaxCandidates {
			return nil, fmt.Errorf("search revision must change its grammar/bound within the declared limit")
		}
		spec.Search.Grammar, spec.Search.MaxCandidates = alternative.Grammar, alternative.MaxCandidates
		spec.MaxAttempts = attempts
		return rewriteAssemblyContract(filename, source, activity, spec)
	}
	return nil, fmt.Errorf("search alternative %q is absent from the source declaration", alternativeID)
}
