package adapter

import "fmt"

func (m Measurements) validate() error {
	counts := []struct {
		name  string
		value int
	}{
		{"source span", m.SourceSpanCount},
		{"unrelated region", m.UnrelatedRegionCount},
		{"repeat", m.RepeatCount},
		{"canonical equal", m.CanonicalEqualCount},
		{"source equal", m.SourceEqualCount},
		{"semantic equal", m.SemanticEqualCount},
		{"region equal", m.RegionEqualCount},
		{"source map resolved", m.SourceMapResolvedCount},
		{"false acceptance", m.FalseAcceptanceCount},
		{"environment leak", m.EnvironmentLeakCount},
	}
	for _, count := range counts {
		if count.value < 0 {
			return fmt.Errorf("measurement %s cannot be negative", count.name)
		}
	}
	if m.CanonicalEqualCount > m.RepeatCount || m.SourceEqualCount > m.RepeatCount ||
		m.SemanticEqualCount > m.RepeatCount || m.RegionEqualCount > m.RepeatCount {
		return fmt.Errorf("equality counts cannot exceed repeat count")
	}
	return nil
}
