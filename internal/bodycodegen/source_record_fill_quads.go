package bodycodegen

import (
	"fmt"
)

// composedFieldRelationExpressionsV3 adds all four-atom Boolean trees before
// the complete v2 triple, pair, and atomic fallback space. Every v3 expression
// uses four distinct relation atoms. The finite denominator counts five binary
// tree shapes and eight operator assignments for each distinct atom quartet.
func (c recordFillGrammarContext) composedFieldRelationExpressionsV3(maxExpressions int) ([]string, int, bool, error) {
	selectors, err := c.fieldRelationSelectors()
	if err != nil {
		return nil, 0, false, err
	}
	atoms := make([]recordFieldRelationAtom, 0)
	atomCount := forEachRecordFieldRelation(selectors, func(_ int, atom recordFieldRelationAtom) bool {
		atoms = append(atoms, atom)
		return true
	})
	if atomCount < 2 {
		return nil, atomCount, false, fmt.Errorf("record relation composition v3 grammar produced fewer than two distinct relations")
	}
	enumerated, err := recordFieldRelationCompositionV3Count(atomCount)
	if err != nil {
		return nil, 0, false, err
	}
	expressions := make([]string, 0, min(maxExpressions, enumerated))

	for distinctPairs := 4; distinctPairs >= 1 && len(expressions) < maxExpressions; distinctPairs-- {
		for distinctSelectors := 8; distinctSelectors >= 2 && len(expressions) < maxExpressions; distinctSelectors-- {
			for first := 0; first < len(atoms) && len(expressions) < maxExpressions; first++ {
				for second := first + 1; second < len(atoms) && len(expressions) < maxExpressions; second++ {
					for third := second + 1; third < len(atoms) && len(expressions) < maxExpressions; third++ {
						for fourth := third + 1; fourth < len(atoms) && len(expressions) < maxExpressions; fourth++ {
							quadruple := [4]recordFieldRelationAtom{atoms[first], atoms[second], atoms[third], atoms[fourth]}
							if recordFieldRelationDistinctPairCount(quadruple[:]) != distinctPairs ||
								recordFieldRelationDistinctSelectorCount(quadruple[:]) != distinctSelectors {
								continue
							}
							for _, shape := range recordFieldRelationQuadrupleShapes(quadruple) {
								for _, firstOperator := range []string{"&&", "||"} {
									for _, secondOperator := range []string{"&&", "||"} {
										for _, thirdOperator := range []string{"&&", "||"} {
											if len(expressions) == maxExpressions {
												break
											}
											expressions = append(expressions, shape(firstOperator, secondOperator, thirdOperator))
										}
										if len(expressions) == maxExpressions {
											break
										}
									}
									if len(expressions) == maxExpressions {
										break
									}
								}
								if len(expressions) == maxExpressions {
									break
								}
							}
						}
					}
				}
			}
		}
	}

	if len(expressions) < maxExpressions {
		fallback, _, _, err := c.composedFieldRelationExpressionsV2(maxExpressions - len(expressions))
		if err != nil {
			return nil, 0, false, err
		}
		expressions = append(expressions, fallback...)
	}
	return expressions, enumerated, len(expressions) == enumerated, nil
}

func recordFieldRelationQuadrupleShapes(quadruple [4]recordFieldRelationAtom) []func(string, string, string) string {
	a, b, c, d := quadruple[0].expression, quadruple[1].expression, quadruple[2].expression, quadruple[3].expression
	return []func(string, string, string) string{
		func(first, second, third string) string {
			ab := "(" + a + ") " + first + " (" + b + ")"
			abc := "(" + ab + ") " + second + " (" + c + ")"
			return "(" + abc + ") " + third + " (" + d + ")"
		},
		func(first, second, third string) string {
			bc := "(" + b + ") " + second + " (" + c + ")"
			abc := "(" + a + ") " + first + " (" + bc + ")"
			return "(" + abc + ") " + third + " (" + d + ")"
		},
		func(first, second, third string) string {
			ab := "(" + a + ") " + first + " (" + b + ")"
			cd := "(" + c + ") " + third + " (" + d + ")"
			return "(" + ab + ") " + second + " (" + cd + ")"
		},
		func(first, second, third string) string {
			bc := "(" + b + ") " + second + " (" + c + ")"
			bcd := "(" + bc + ") " + third + " (" + d + ")"
			return "(" + a + ") " + first + " (" + bcd + ")"
		},
		func(first, second, third string) string {
			cd := "(" + c + ") " + third + " (" + d + ")"
			bcd := "(" + b + ") " + second + " (" + cd + ")"
			return "(" + a + ") " + first + " (" + bcd + ")"
		},
	}
}

func recordFieldRelationCompositionV3Count(atomCount int) (int, error) {
	if atomCount < 0 {
		return 0, fmt.Errorf("record relation composition v3 space exceeds the supported counter range")
	}
	quadrupleCount, err := recordFieldRelationChooseFour(atomCount)
	if err != nil {
		return 0, fmt.Errorf("record relation composition v3 space exceeds the supported counter range")
	}
	maxInt := int(^uint(0) >> 1)
	if quadrupleCount > maxInt/40 {
		return 0, fmt.Errorf("record relation composition v3 space exceeds the supported counter range")
	}
	quadrupleCount *= 40
	fallbackCount, err := recordFieldRelationCompositionV2Count(atomCount)
	if err != nil || fallbackCount > maxInt-quadrupleCount {
		return 0, fmt.Errorf("record relation composition v3 space exceeds the supported counter range")
	}
	return quadrupleCount + fallbackCount, nil
}

func recordFieldRelationChooseFour(atomCount int) (int, error) {
	if atomCount < 0 {
		return 0, fmt.Errorf("negative atom count")
	}
	if atomCount < 4 {
		return 0, nil
	}
	factors := []int{atomCount, atomCount - 1, atomCount - 2, atomCount - 3}
	denominator := 24
	for index := range factors {
		divisor := recordFieldRelationGreatestCommonDivisor(factors[index], denominator)
		factors[index] /= divisor
		denominator /= divisor
	}
	if denominator != 1 {
		return 0, fmt.Errorf("choose-four denominator did not reduce")
	}
	maxInt := int(^uint(0) >> 1)
	count := 1
	for _, factor := range factors {
		if factor != 0 && count > maxInt/factor {
			return 0, fmt.Errorf("choose-four count overflows int")
		}
		count *= factor
	}
	return count, nil
}

func recordFieldRelationGreatestCommonDivisor(left, right int) int {
	for right != 0 {
		left, right = right, left%right
	}
	return left
}

func recordFieldRelationDistinctPairCount(atoms []recordFieldRelationAtom) int {
	seen := make(map[[2]string]bool, len(atoms))
	for _, atom := range atoms {
		seen[[2]string{atom.left, atom.right}] = true
	}
	return len(seen)
}

func recordFieldRelationDistinctSelectorCount(atoms []recordFieldRelationAtom) int {
	seen := make(map[string]bool, len(atoms)*2)
	for _, atom := range atoms {
		seen[atom.left] = true
		seen[atom.right] = true
	}
	return len(seen)
}
