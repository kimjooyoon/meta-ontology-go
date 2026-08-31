package toolchainformatfix

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
)

type cycleEvidence struct {
	applications, fixedPoints, directWrites int
	passed                                  bool
}

func evaluateCycle() cycleEvidence {
	change := formatfix.Build(unformattedPath, unformatted)
	if formatfix.Validate(change) != nil || change.Decision != formatfix.DecisionChangePlanned {
		return cycleEvidence{}
	}
	result, err := formatfix.Apply(unformatted, change)
	if err != nil {
		return cycleEvidence{}
	}
	fixedAfter := formatfix.Build(unformattedPath, result)
	fixedCanonical := formatfix.Build(canonicalPath, canonical)
	if formatfix.Validate(fixedAfter) != nil || formatfix.Validate(fixedCanonical) != nil ||
		fixedAfter.Decision != formatfix.DecisionFixedPoint ||
		fixedCanonical.Decision != formatfix.DecisionFixedPoint {
		return cycleEvidence{}
	}
	return cycleEvidence{applications: 1, fixedPoints: 2,
		directWrites: change.DirectWrites + fixedAfter.DirectWrites + fixedCanonical.DirectWrites,
		passed:       true}
}
