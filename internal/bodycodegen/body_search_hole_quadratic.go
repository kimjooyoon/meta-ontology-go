package bodycodegen

import (
	"context"
	"math/big"
	"slices"
	"strconv"
)

func observeQuadraticHoleProbe(ctx context.Context, activity string, negative []byte,
	test IRBodyFillTestCase, calls *int, probe *IRBodyHoleProbe) {
	probe.Status = "PROBE_UNAVAILABLE"
	if len(negative) > 0 && ctx.Err() == nil {
		*calls += 1
		results, _, err := evaluateIntegerCasesContext(ctx, negative, activity, []IRBodyFillTestCase{test})
		if err == nil && len(results) == 1 {
			value := results[0].Actual
			probe.MinusOne = &value
		}
	}
	if probe.MinusOne != nil && probe.Zero != nil && probe.One != nil {
		probe.Roots, probe.Status = exactQuadraticHoleRoots(*probe.MinusOne, *probe.Zero, *probe.One, test.Expected)
	}
}

// The coefficients are doubled to keep all arithmetic integral. A three-point
// fit proposes values; execution of every candidate checks the original body.
func exactQuadraticHoleRoots(minusOne, zero, one, expected int64) ([]int64, string) {
	two := big.NewInt(2)
	a := new(big.Int).Sub(new(big.Int).Add(big.NewInt(minusOne), big.NewInt(one)), new(big.Int).Mul(two, big.NewInt(zero)))
	b := new(big.Int).Sub(big.NewInt(one), big.NewInt(minusOne))
	c := new(big.Int).Mul(two, new(big.Int).Sub(big.NewInt(zero), big.NewInt(expected)))
	if a.Sign() == 0 {
		if b.Sign() == 0 {
			return nil, "NO_OBSERVED_SENSITIVITY"
		}
		if root, ok := exactInt64Quotient(new(big.Int).Neg(c), b); ok {
			return []int64{root}, "LINEAR_PROPOSAL"
		}
		return nil, "NO_INT64_ROOT_PROPOSAL"
	}
	discriminant := new(big.Int).Sub(new(big.Int).Mul(b, b), new(big.Int).Mul(big.NewInt(4), new(big.Int).Mul(a, c)))
	if discriminant.Sign() < 0 {
		return nil, "NO_REAL_ROOT_PROPOSAL"
	}
	squareRoot := new(big.Int).Sqrt(discriminant)
	if new(big.Int).Mul(squareRoot, squareRoot).Cmp(discriminant) != 0 {
		return nil, "NO_INTEGER_ROOT_PROPOSAL"
	}
	denominator := new(big.Int).Mul(two, a)
	var roots []int64
	for _, sign := range []int64{-1, 1} {
		numerator := new(big.Int).Add(new(big.Int).Neg(b), new(big.Int).Mul(big.NewInt(sign), squareRoot))
		if root, ok := exactInt64Quotient(numerator, denominator); ok && !slices.Contains(roots, root) {
			roots = append(roots, root)
		}
	}
	if len(roots) == 0 {
		return nil, "NO_INT64_ROOT_PROPOSAL"
	}
	slices.Sort(roots)
	return roots, "QUADRATIC_PROPOSAL"
}

func exactInt64Quotient(numerator, denominator *big.Int) (int64, bool) {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	return quotient.Int64(), remainder.Sign() == 0 && quotient.IsInt64()
}

func quadraticHoleExpressions(probes []IRBodyHoleProbe, fallback []string) []string {
	seen := map[string]bool{}
	var expressions []string
	add := func(expression string) {
		if expression != "" && !seen[expression] {
			seen[expression] = true
			expressions = append(expressions, expression)
		}
	}
	for _, probe := range probes {
		for _, root := range probe.Roots {
			add(strconv.FormatInt(root, 10))
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
				add(holeAffineExpression(IRBodyHoleProbe{Input: anchor.Input, Derived: &first}, IRBodyHoleProbe{Input: probe.Input, Derived: &root}))
			}
			if offset, ok := bodySearchInt64Delta(probe.Input, root); ok {
				add(holeLinearExpression(1, offset))
			}
		}
	}
	for _, expression := range fallback {
		add(expression)
	}
	return expressions
}
