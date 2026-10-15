package bodyexecution

import "reflect"

// Optional assembly metadata refines a site's explanation while the existing
// lexical-write identity remains stable across this additive record format.
func localSiteIdentity(site CompositionLocalSite) string {
	site.ExpressionID = ""
	site.Source.Construction, site.Source.ConstructionReason = nil, ""
	return compositionDigest(site)
}

func equalRetainedLocalSites(current, recorded []CompositionLocalSite) bool {
	if len(current) != len(recorded) {
		return false
	}
	for index, prior := range recorded {
		expected := current[index]
		if prior.Source.Construction == nil && prior.Source.ConstructionReason == "" {
			expected.Source.Construction, expected.Source.ConstructionReason = nil, ""
		}
		if !reflect.DeepEqual(expected, prior) {
			return false
		}
	}
	return true
}
