package bodycodegen

import (
	"math/big"
	"strconv"
)

func exactHoleResidual(zero, one, expected int64) (*int64, string) {
	coefficient := new(big.Int).Sub(big.NewInt(one), big.NewInt(zero))
	if coefficient.Sign() == 0 {
		return nil, "NO_OBSERVED_SENSITIVITY"
	}
	residual := new(big.Int).Sub(big.NewInt(expected), big.NewInt(zero))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(residual, coefficient, remainder)
	if remainder.Sign() != 0 {
		return nil, "NON_INTEGRAL_RESIDUAL"
	}
	if !quotient.IsInt64() {
		return nil, "RESIDUAL_OUT_OF_RANGE"
	}
	value := quotient.Int64()
	return &value, "DERIVED_PROPOSAL"
}

func holeContextExpressions(probes []IRBodyHoleProbe, fallback []string) []string {
	seen := make(map[string]bool)
	expressions := make([]string, 0, len(probes)*3+len(fallback))
	add := func(expression string) {
		if expression != "" && !seen[expression] {
			seen[expression] = true
			expressions = append(expressions, expression)
		}
	}
	for _, probe := range probes {
		if probe.Derived != nil {
			add(strconv.FormatInt(*probe.Derived, 10))
		}
	}
	var anchor *IRBodyHoleProbe
	for i, probe := range probes {
		if probe.Derived == nil {
			continue
		}
		if anchor == nil {
			anchor = &probes[i]
		} else {
			add(holeAffineExpression(*anchor, probe))
		}
		if offset, ok := bodySearchInt64Delta(probe.Input, *probe.Derived); ok {
			add(holeLinearExpression(1, offset))
		}
	}
	for _, expression := range fallback {
		add(expression)
	}
	return expressions
}

func holeAffineExpression(first, next IRBodyHoleProbe) string {
	if first.Input == next.Input || first.Derived == nil || next.Derived == nil {
		return ""
	}
	dx := new(big.Int).Sub(big.NewInt(next.Input), big.NewInt(first.Input))
	dy := new(big.Int).Sub(big.NewInt(*next.Derived), big.NewInt(*first.Derived))
	coefficient, remainder := new(big.Int), new(big.Int)
	coefficient.QuoRem(dy, dx, remainder)
	if remainder.Sign() != 0 || !coefficient.IsInt64() {
		return ""
	}
	offset := new(big.Int).Sub(big.NewInt(*first.Derived), new(big.Int).Mul(coefficient, big.NewInt(first.Input)))
	if !offset.IsInt64() {
		return ""
	}
	return holeLinearExpression(coefficient.Int64(), offset.Int64())
}

func holeLinearExpression(coefficient, offset int64) string {
	if coefficient == 0 {
		return strconv.FormatInt(offset, 10)
	}
	expression := "input"
	if coefficient == -1 {
		expression = "-input"
	} else if coefficient != 1 {
		expression += " * " + strconv.FormatInt(coefficient, 10)
	}
	if offset != 0 {
		expression += " + " + strconv.FormatInt(offset, 10)
	}
	return expression
}
