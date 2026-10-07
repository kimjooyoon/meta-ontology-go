package bodycodegen

import "testing"

func TestHoleResidualArithmeticPreservesIntegerBoundaries(t *testing.T) {
	for _, test := range []struct {
		zero, one, expected, value int64
		status                     string
	}{
		{2, 3, 3, 1, "DERIVED_PROPOSAL"},
		{4, 6, 10, 3, "DERIVED_PROPOSAL"},
		{-2, 1, 7, 3, "DERIVED_PROPOSAL"},
		{0, 0, 1, 0, "NO_OBSERVED_SENSITIVITY"},
		{0, 2, 3, 0, "NON_INTEGRAL_RESIDUAL"},
		{-9223372036854775808, -9223372036854775807, 9223372036854775807, 0, "RESIDUAL_OUT_OF_RANGE"},
		{9223372036854775807, -9223372036854775808, -9223372036854775808, 1, "DERIVED_PROPOSAL"},
	} {
		value, status := exactHoleResidual(test.zero, test.one, test.expected)
		if status != test.status || (value != nil) != (status == "DERIVED_PROPOSAL") || value != nil && *value != test.value {
			t.Fatal(test, value, status)
		}
	}
	min, max := int64(-9223372036854775808), int64(9223372036854775807)
	if got := holeAffineExpression(IRBodyHoleProbe{Input: min, Derived: &min}, IRBodyHoleProbe{Input: max, Derived: &max}); got != "input" {
		t.Fatal("affine arithmetic overflowed", got)
	}
}
