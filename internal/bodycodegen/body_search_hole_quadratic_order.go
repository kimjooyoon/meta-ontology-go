package bodycodegen

import (
	"math/big"
	"slices"
)

// v2 retains the v1 candidate universe and promotes root-derived expressions
// compatible with all available integral root observations. Whole-body scoring
// remains separate: a three-point fit may disagree with the original program.
func orderQuadraticHoleExpressions(probes []IRBodyHoleProbe, original []string) []string {
	available, seen := map[string]bool{}, map[string]bool{}
	for _, expression := range original {
		available[expression] = true
	}
	ordered := make([]string, 0, len(original))
	add := func(coefficient, offset int64) {
		expression := holeLinearExpression(coefficient, offset)
		if available[expression] && !seen[expression] && fitsQuadraticRootObservations(probes, coefficient, offset) {
			ordered = append(ordered, expression)
			seen[expression] = true
		}
	}
	for _, probe := range probes {
		for _, root := range probe.Roots {
			add(0, root)
		}
	}
	var anchor *IRBodyHoleProbe
	for i, probe := range probes {
		if len(probe.Roots) == 0 {
			continue
		}
		if anchor == nil {
			anchor = &probes[i]
		}
		for _, root := range probe.Roots {
			for _, first := range anchor.Roots {
				coefficient, offset, ok := holeAffineCoefficients(
					IRBodyHoleProbe{Input: anchor.Input, Derived: &first}, IRBodyHoleProbe{Input: probe.Input, Derived: &root})
				if ok {
					add(coefficient, offset)
				}
			}
		}
	}
	for _, expression := range original {
		if !seen[expression] {
			ordered = append(ordered, expression)
			seen[expression] = true
		}
	}
	return ordered
}

func fitsQuadraticRootObservations(probes []IRBodyHoleProbe, coefficient, offset int64) bool {
	observed := false
	for _, probe := range probes {
		if len(probe.Roots) == 0 {
			continue
		}
		observed = true
		value := new(big.Int).Add(new(big.Int).Mul(big.NewInt(coefficient), big.NewInt(probe.Input)), big.NewInt(offset))
		if !value.IsInt64() || !slices.Contains(probe.Roots, value.Int64()) {
			return false
		}
	}
	return observed
}
