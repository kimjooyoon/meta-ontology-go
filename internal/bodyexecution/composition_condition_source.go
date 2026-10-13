package bodyexecution

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func comparableConditionSites(runtime CompositionRuntime) []CompositionConditionSite {
	if runtime.ConditionSites == nil {
		return nil
	}
	sites := append([]CompositionConditionSite(nil), runtime.ConditionSites...)
	for i := range sites {
		// Normalize the explicit absence marker for receipts made before origin
		// retention. Mapped and other origin claims remain comparison inputs.
		absent := unmappedConditionSource(runtime.SelectedSourceSHA256, "NO_RETAINED_ORIGIN")
		if reflect.DeepEqual(sites[i].Source, absent) {
			sites[i].Source = nil
		}
	}
	return sites
}

type conditionOriginKey struct {
	RootID, ActivityID string
	Ordinal            int
}

func compositionSourceConditions(prior Composition) (map[conditionOriginKey]*bodycodegen.BodyConditionSource, error) {
	var origins []bodycodegen.BodyConditionOrigin
	var keys []conditionOriginKey
	for _, step := range prior.Steps {
		for _, origin := range step.Generation.Report.BodyConditions {
			origins = append(origins, origin)
			keys = append(keys, conditionOriginKey{step.Generation.Report.ActivityID, origin.ActivityID, origin.NativeOrdinal})
		}
	}
	result := make(map[conditionOriginKey]*bodycodegen.BodyConditionSource, len(origins))
	if len(origins) == 0 {
		return result, nil
	}
	bound, err := bodycodegen.BindBodyConditionSources([]byte(prior.GoooSource), origins)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		result[key] = bound[i]
		for _, step := range prior.Steps {
			if fill := step.Generation.Report.BodyFill; fill != nil && fill.SourceOrigin != nil {
				if link := fill.SourceOrigin.ConditionOrigin(origins[i]); link != nil {
					bound[i].Construction = link
					break
				}
			}
		}
		for _, preparation := range prior.Preparations {
			if fill := preparation.Generation.Report.BodyFill; fill != nil && fill.SourceOrigin != nil {
				if link := fill.SourceOrigin.ConditionOrigin(origins[i]); link != nil {
					bound[i].Construction = link
					break
				}
			}
		}
	}
	return result, nil
}

func unmappedConditionSource(sourceSHA, reason string) *bodycodegen.BodyConditionSource {
	return bodycodegen.UnmappedConditionSource(sourceSHA, reason)
}
