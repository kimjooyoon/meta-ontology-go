package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	recordFieldPredicateGrammar             = "record-field-predicate/v1"
	recordFieldPredicateV2Grammar           = "record-field-predicate/v2"
	recordFieldRelationGrammar              = "record-field-relation/v1"
	recordFieldRelationCompositionGrammar   = "record-field-relation-composition/v1"
	recordFieldRelationCompositionV2Grammar = "record-field-relation-composition/v2"
	recordPredicateCompositionGrammar       = "record-field-predicate-composition/v1"
	recordPredicateCompositionV2Grammar     = "record-field-predicate-composition/v2"
	recordStringLiteralGrammar              = "record-string-literal/v1"
	recordIntegerLiteralGrammar             = "record-integer-literal/v1"
	recordBooleanLiteralGrammar             = "record-boolean-literal/v1"
)

// generateSourceRecordFillCandidates derives a finite, source-bounded expression
// grammar from the typed activity signature and its declared record examples.
func generateSourceRecordFillCandidates(filename string, source []byte, activityName string, plan *assemblyspec.FillPlan, valueCases []assemblyspec.ValueCase) ([]IRBodyFillCandidate, *IRBodyFillCandidateGenerationReceipt, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return nil, nil, fmt.Errorf("parse Gooo record source-fill declaration: %v", diagnostics)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		candidate, ok := declaration.(*syntax.ActivityDecl)
		if ok && candidate.Name == activityName {
			activity = candidate
			break
		}
	}
	if activity == nil {
		return nil, nil, fmt.Errorf("record source-fill activity %q was not found", activityName)
	}
	_, records, err := resolveBodyModel(file)
	if err != nil {
		return nil, nil, err
	}
	output := recordTypeByName(records, activity.Output)
	if output == nil {
		return nil, nil, fmt.Errorf("record source-fill derivation requires a declared record output")
	}
	context := recordFillGrammarContext{activity: activity, records: records, output: output, cases: valueCases}
	return generateSourceFillCandidatesWith(plan, func(grammar assemblyspec.FillHoleGrammar) ([]string, int, bool, error) {
		return context.expressions(grammar)
	})
}

type recordFillGrammarContext struct {
	activity *syntax.ActivityDecl
	records  []RecordType
	output   *RecordType
	cases    []assemblyspec.ValueCase
}

func (c recordFillGrammarContext) expressions(grammar assemblyspec.FillHoleGrammar) ([]string, int, bool, error) {
	switch grammar.Grammar {
	case recordFieldPredicateGrammar:
		return c.predicateExpressions(grammar.MaxExpressions, false)
	case recordFieldPredicateV2Grammar:
		return c.predicateExpressions(grammar.MaxExpressions, true)
	case recordFieldRelationGrammar:
		return c.fieldRelationExpressions(grammar.MaxExpressions)
	case recordFieldRelationCompositionGrammar:
		return c.composedFieldRelationExpressions(grammar.MaxExpressions)
	case recordFieldRelationCompositionV2Grammar:
		return c.composedFieldRelationExpressionsV2(grammar.MaxExpressions)
	case recordPredicateCompositionGrammar:
		return c.composedPredicateExpressions(grammar.MaxExpressions, false)
	case recordPredicateCompositionV2Grammar:
		return c.composedPredicateExpressions(grammar.MaxExpressions, true)
	case recordStringLiteralGrammar:
		return c.outputLiteralExpressions(grammar.MaxExpressions, semantic.BuiltinStringTypeID)
	case recordIntegerLiteralGrammar:
		return c.outputLiteralExpressions(grammar.MaxExpressions, semantic.BuiltinIntegerTypeID)
	case recordBooleanLiteralGrammar:
		return c.outputLiteralExpressions(grammar.MaxExpressions, semantic.BuiltinBooleanTypeID)
	default:
		return nil, 0, false, fmt.Errorf("unsupported record source-fill grammar %q", grammar.Grammar)
	}
}

// fieldRelationExpressions derives comparisons between distinct fields on
// declared record inputs. Unlike literal predicates, these candidates express
// relationships between values supplied by the caller rather than memorizing
// individual observed values.
func (c recordFillGrammarContext) fieldRelationExpressions(maxExpressions int) ([]string, int, bool, error) {
	selectors, err := c.fieldRelationSelectors()
	if err != nil {
		return nil, 0, false, err
	}
	enumerated := forEachRecordFieldRelation(selectors, func(_ int, _ recordFieldRelationAtom) bool { return true })
	if enumerated < 2 {
		return nil, enumerated, false, fmt.Errorf("record relation grammar produced fewer than two distinct expressions")
	}
	expressions := make([]string, 0, min(maxExpressions, enumerated))
	forEachRecordFieldRelation(selectors, func(_ int, atom recordFieldRelationAtom) bool {
		if len(expressions) < maxExpressions {
			expressions = append(expressions, atom.expression)
		}
		return true
	})
	return expressions, enumerated, len(expressions) == enumerated, nil
}

type recordFieldRelationSelector struct {
	expression string
	typeID     string
}

type recordFieldRelationAtom struct {
	expression string
	left       string
	right      string
}

func (c recordFillGrammarContext) fieldRelationSelectors() ([]recordFieldRelationSelector, error) {
	selectors := make([]recordFieldRelationSelector, 0)
	for inputIndex, input := range c.activity.Inputs {
		if scalarTypeID(input.Name) != "" {
			continue
		}
		record := recordTypeByName(c.records, input.Name)
		if record == nil {
			return nil, fmt.Errorf("record relation grammar cannot resolve input type %q", input.Name)
		}
		root := "input"
		if len(c.activity.Inputs) > 1 {
			root = fmt.Sprintf("input%d", inputIndex)
		}
		for _, field := range record.Fields {
			if field.TypeID != string(semantic.BuiltinStringTypeID) &&
				field.TypeID != string(semantic.BuiltinBooleanTypeID) &&
				field.TypeID != string(semantic.BuiltinIntegerTypeID) {
				continue
			}
			selectors = append(selectors, recordFieldRelationSelector{expression: root + "." + field.Name, typeID: field.TypeID})
		}
	}
	return selectors, nil
}

func forEachRecordFieldRelation(selectors []recordFieldRelationSelector, visit func(int, recordFieldRelationAtom) bool) int {
	count := 0
	for leftIndex, left := range selectors {
		for _, right := range selectors[leftIndex+1:] {
			if left.typeID != right.typeID {
				continue
			}
			operators := []string{"==", "!="}
			if left.typeID == string(semantic.BuiltinIntegerTypeID) {
				operators = []string{"<", "<=", ">", ">=", "==", "!="}
			}
			for _, operator := range operators {
				atom := recordFieldRelationAtom{
					expression: left.expression + " " + operator + " " + right.expression,
					left:       left.expression, right: right.expression,
				}
				if !visit(count, atom) {
					return count + 1
				}
				count++
			}
		}
	}
	return count
}

func (c recordFillGrammarContext) composedFieldRelationExpressions(maxExpressions int) ([]string, int, bool, error) {
	selectors, err := c.fieldRelationSelectors()
	if err != nil {
		return nil, 0, false, err
	}
	atomCount := forEachRecordFieldRelation(selectors, func(_ int, _ recordFieldRelationAtom) bool { return true })
	maxInt := int(^uint(0) >> 1)
	if atomCount < 2 {
		return nil, atomCount, false, fmt.Errorf("record relation composition grammar produced fewer than two distinct expressions")
	}
	if atomCount > maxInt/atomCount {
		return nil, 0, false, fmt.Errorf("record relation composition space exceeds the supported counter range")
	}
	enumerated := atomCount * atomCount
	expressions := make([]string, 0, min(maxExpressions, enumerated))
	appendPair := func(left, right recordFieldRelationAtom) {
		for _, operator := range []string{"&&", "||"} {
			if len(expressions) == maxExpressions {
				return
			}
			expressions = append(expressions, "("+left.expression+") "+operator+" ("+right.expression+")")
		}
	}
	for _, preferDisjoint := range []bool{true, false} {
		forEachRecordFieldRelation(selectors, func(leftIndex int, left recordFieldRelationAtom) bool {
			forEachRecordFieldRelation(selectors, func(rightIndex int, right recordFieldRelationAtom) bool {
				if rightIndex <= leftIndex || recordFieldRelationsDisjoint(left, right) != preferDisjoint {
					return true
				}
				appendPair(left, right)
				return len(expressions) < maxExpressions
			})
			return len(expressions) < maxExpressions
		})
		if len(expressions) == maxExpressions {
			break
		}
	}
	if len(expressions) < maxExpressions {
		forEachRecordFieldRelation(selectors, func(_ int, atom recordFieldRelationAtom) bool {
			expressions = append(expressions, atom.expression)
			return len(expressions) < maxExpressions
		})
	}
	return expressions, enumerated, len(expressions) == enumerated, nil
}

// composedFieldRelationExpressionsV2 extends the pair grammar with all
// three-atom parenthesizations. Triple candidates come first, ordered by the
// number of distinct selector pairs and then selector fields they use. Pair
// candidates and atoms remain in the grammar as simpler fallbacks. The exact
// denominator counts 8 expressions per distinct atom triple, both operators
// for each distinct atom pair, and every atom once.
func (c recordFillGrammarContext) composedFieldRelationExpressionsV2(maxExpressions int) ([]string, int, bool, error) {
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
		return nil, atomCount, false, fmt.Errorf("record relation composition v2 grammar produced fewer than two distinct relations")
	}
	enumerated, err := recordFieldRelationCompositionV2Count(atomCount)
	if err != nil {
		return nil, 0, false, err
	}
	expressions := make([]string, 0, min(maxExpressions, enumerated))

	for distinctPairs := 3; distinctPairs >= 1 && len(expressions) < maxExpressions; distinctPairs-- {
		for distinctSelectors := 6; distinctSelectors >= 2 && len(expressions) < maxExpressions; distinctSelectors-- {
			for first := 0; first < len(atoms) && len(expressions) < maxExpressions; first++ {
				for second := first + 1; second < len(atoms) && len(expressions) < maxExpressions; second++ {
					for third := second + 1; third < len(atoms) && len(expressions) < maxExpressions; third++ {
						triple := [3]recordFieldRelationAtom{atoms[first], atoms[second], atoms[third]}
						if recordFieldRelationTriplePairCount(triple) != distinctPairs ||
							recordFieldRelationTripleSelectorCount(triple) != distinctSelectors {
							continue
						}
						for _, leftOperator := range []string{"&&", "||"} {
							for _, rightOperator := range []string{"&&", "||"} {
								left := "((" + triple[0].expression + ") " + leftOperator + " (" + triple[1].expression + ")) " + rightOperator + " (" + triple[2].expression + ")"
								right := "(" + triple[0].expression + ") " + leftOperator + " ((" + triple[1].expression + ") " + rightOperator + " (" + triple[2].expression + "))"
								expressions = append(expressions, left, right)
								if len(expressions) >= maxExpressions {
									expressions = expressions[:maxExpressions]
									break
								}
							}
							if len(expressions) >= maxExpressions {
								break
							}
						}
					}
				}
			}
		}
	}

	appendPair := func(left, right recordFieldRelationAtom) {
		for _, operator := range []string{"&&", "||"} {
			if len(expressions) == maxExpressions {
				return
			}
			expressions = append(expressions, "("+left.expression+") "+operator+" ("+right.expression+")")
		}
	}
	for _, preferDisjoint := range []bool{true, false} {
		for leftIndex, left := range atoms {
			for rightIndex := leftIndex + 1; rightIndex < len(atoms); rightIndex++ {
				right := atoms[rightIndex]
				if recordFieldRelationsDisjoint(left, right) != preferDisjoint {
					continue
				}
				appendPair(left, right)
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
	for _, atom := range atoms {
		if len(expressions) == maxExpressions {
			break
		}
		expressions = append(expressions, atom.expression)
	}
	return expressions, enumerated, len(expressions) == enumerated, nil
}

func recordFieldRelationCompositionV2Count(atomCount int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	if atomCount < 0 || atomCount != 0 && atomCount > maxInt/atomCount {
		return 0, fmt.Errorf("record relation composition v2 space exceeds the supported counter range")
	}
	count := atomCount * atomCount
	if atomCount < 3 {
		return count, nil
	}
	factors := [3]int{atomCount, atomCount - 1, atomCount - 2}
	for _, divisor := range []int{2, 3} {
		for index, factor := range factors {
			if factor%divisor == 0 {
				factors[index] /= divisor
				break
			}
		}
	}
	triples := factors[0]
	for _, factor := range factors[1:] {
		if factor != 0 && triples > maxInt/factor {
			return 0, fmt.Errorf("record relation composition v2 space exceeds the supported counter range")
		}
		triples *= factor
	}
	if triples > maxInt/8 {
		return 0, fmt.Errorf("record relation composition v2 space exceeds the supported counter range")
	}
	triples *= 8
	if count > maxInt-triples {
		return 0, fmt.Errorf("record relation composition v2 space exceeds the supported counter range")
	}
	return count + triples, nil
}

func recordFieldRelationTriplePairCount(triple [3]recordFieldRelationAtom) int {
	count := 0
	seen := make(map[[2]string]bool, 3)
	for _, atom := range triple {
		pair := [2]string{atom.left, atom.right}
		if !seen[pair] {
			seen[pair] = true
			count++
		}
	}
	return count
}

func recordFieldRelationTripleSelectorCount(triple [3]recordFieldRelationAtom) int {
	selectors := make(map[string]bool, 6)
	for _, atom := range triple {
		selectors[atom.left] = true
		selectors[atom.right] = true
	}
	return len(selectors)
}

func recordFieldRelationsDisjoint(left, right recordFieldRelationAtom) bool {
	return left.left != right.left && left.left != right.right &&
		left.right != right.left && left.right != right.right
}

func (c recordFillGrammarContext) predicateExpressions(maxExpressions int, integerOrder bool) ([]string, int, bool, error) {
	atoms, err := c.predicateAtoms(integerOrder)
	if err != nil {
		return nil, 0, false, err
	}
	return retainRecordFillExpressions(atoms, maxExpressions)
}

func (c recordFillGrammarContext) composedPredicateExpressions(maxExpressions int, integerOrder bool) ([]string, int, bool, error) {
	atoms, err := c.predicateAtoms(integerOrder)
	if err != nil {
		return nil, 0, false, err
	}
	count := len(atoms)
	maxInt := int(^uint(0) >> 1)
	if count != 0 && count > maxInt/count {
		return nil, 0, false, fmt.Errorf("record predicate composition space exceeds the supported counter range")
	}
	enumerated := count * count
	if enumerated < 2 {
		return nil, enumerated, false, fmt.Errorf("record predicate composition grammar produced fewer than two distinct expressions")
	}
	expressions := make([]string, 0, min(maxExpressions, enumerated))
	seen := make(map[string]bool, maxExpressions)
	appendExpression := func(expression string) {
		if len(expressions) < maxExpressions && !seen[expression] {
			seen[expression] = true
			expressions = append(expressions, expression)
		}
	}
	if integerOrder {
		// The v2 cap alternates cross-field conditions with same-field integer
		// intervals. This keeps a small grammar useful for both record routing
		// and numeric range decisions.
		crossField := recordPredicatePairExpressions(atoms, true, maxExpressions)
		integerRanges := recordIntegerRangeExpressions(atoms, maxExpressions)
		for index := 0; len(expressions) < maxExpressions && (index < len(crossField) || index < len(integerRanges)); index++ {
			if index < len(crossField) {
				appendExpression(crossField[index])
			}
			if index < len(integerRanges) {
				appendExpression(integerRanges[index])
			}
		}
		for _, expression := range crossField {
			appendExpression(expression)
		}
		for _, expression := range integerRanges {
			appendExpression(expression)
		}
		for _, expression := range recordPredicatePairExpressions(atoms, false, maxExpressions) {
			appendExpression(expression)
		}
	} else {
		// Keep the v1 cross-field-first candidate sequence stable.
		for _, expression := range recordPredicatePairExpressions(atoms, true, maxExpressions) {
			appendExpression(expression)
		}
		for _, expression := range recordPredicatePairExpressions(atoms, false, maxExpressions) {
			appendExpression(expression)
		}
	}
	for _, atom := range atoms {
		if len(expressions) == maxExpressions {
			break
		}
		appendExpression(atom)
	}
	return expressions, enumerated, len(expressions) == enumerated, nil
}

func recordPredicatePairExpressions(atoms []string, distinctSelectors bool, maxExpressions int) []string {
	expressions := make([]string, 0, min(maxExpressions, len(atoms)))
	for left := range atoms {
		for right := left + 1; right < len(atoms) && len(expressions) < maxExpressions; right++ {
			leftSelector, rightSelector := recordPredicateSelector(atoms[left]), recordPredicateSelector(atoms[right])
			if (leftSelector != rightSelector) != distinctSelectors {
				continue
			}
			for _, operator := range []string{"&&", "||"} {
				if len(expressions) == maxExpressions {
					break
				}
				expressions = append(expressions, "("+atoms[left]+") "+operator+" ("+atoms[right]+")")
			}
		}
	}
	return expressions
}

func recordIntegerRangeExpressions(atoms []string, maxExpressions int) []string {
	type selectorThresholds map[int64]map[string]string
	bySelector := make(map[string]selectorThresholds)
	for _, atom := range atoms {
		selector, operator, threshold, ok := recordIntegerPredicateParts(atom)
		if !ok || (operator != "<" && operator != "<=" && operator != ">" && operator != ">=") {
			continue
		}
		if bySelector[selector] == nil {
			bySelector[selector] = make(selectorThresholds)
		}
		if bySelector[selector][threshold] == nil {
			bySelector[selector][threshold] = make(map[string]string)
		}
		bySelector[selector][threshold][operator] = atom
	}

	selectors := make([]string, 0, len(bySelector))
	for selector := range bySelector {
		selectors = append(selectors, selector)
	}
	slices.Sort(selectors)
	expressions := make([]string, 0, maxExpressions)
	for _, selector := range selectors {
		thresholds := make([]int64, 0, len(bySelector[selector]))
		for threshold := range bySelector[selector] {
			thresholds = append(thresholds, threshold)
		}
		slices.Sort(thresholds)
		for lowerIndex, lower := range thresholds {
			for upperIndex := lowerIndex + 1; upperIndex < len(thresholds); upperIndex++ {
				upper := thresholds[upperIndex]
				for _, lowerOperator := range []string{">=", ">"} {
					lowerExpression := bySelector[selector][lower][lowerOperator]
					if lowerExpression == "" {
						continue
					}
					for _, upperOperator := range []string{"<=", "<"} {
						upperExpression := bySelector[selector][upper][upperOperator]
						if upperExpression == "" {
							continue
						}
						expressions = append(expressions, "("+lowerExpression+") && ("+upperExpression+")")
						if len(expressions) == maxExpressions {
							return expressions
						}
					}
				}
			}
		}
	}
	return expressions
}

func recordIntegerPredicateParts(expression string) (selector, operator string, threshold int64, ok bool) {
	for _, candidate := range []string{" <= ", " >= ", " < ", " > "} {
		left, right, found := strings.Cut(expression, candidate)
		if !found {
			continue
		}
		value, err := strconv.ParseInt(strings.TrimSpace(right), 10, 64)
		if err != nil || strings.TrimSpace(left) == "" {
			return "", "", 0, false
		}
		return strings.TrimSpace(left), strings.TrimSpace(candidate), value, true
	}
	return "", "", 0, false
}

func recordPredicateSelector(expression string) string {
	for _, operator := range []string{" <= ", " >= ", " == ", " != ", " < ", " > "} {
		if selector, _, ok := strings.Cut(expression, operator); ok {
			return selector
		}
	}
	return expression
}

func (c recordFillGrammarContext) predicateAtoms(integerOrder bool) ([]string, error) {
	expressions := make([]string, 0)
	seen := map[string]bool{}
	for _, valueCase := range c.cases {
		inputs, err := decodeRecordFillInputTuple(valueCase.Inputs, len(c.activity.Inputs))
		if err != nil {
			return nil, err
		}
		for inputIndex, declaration := range c.activity.Inputs {
			root := "input"
			if len(c.activity.Inputs) > 1 {
				root = fmt.Sprintf("input%d", inputIndex)
			}
			typeName := declaration.Name
			if scalarTypeID(typeName) != "" {
				literal, ok := recordFillLiteral(inputs[inputIndex], scalarTypeID(typeName))
				if ok {
					ordered := integerOrder && scalarTypeID(typeName) == string(semantic.BuiltinIntegerTypeID)
					appendRecordPredicates(&expressions, seen, root, literal, ordered)
				}
				continue
			}
			record := recordTypeByName(c.records, typeName)
			if record == nil {
				return nil, fmt.Errorf("record predicate grammar cannot resolve input type %q", typeName)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(inputs[inputIndex], &fields); err != nil || fields == nil {
				return nil, fmt.Errorf("record predicate case input %d must be a JSON object", inputIndex)
			}
			for _, field := range record.Fields {
				raw, exists := fields[field.Name]
				if !exists {
					return nil, fmt.Errorf("record predicate case is missing input field %q", field.Name)
				}
				literal, ok := recordFillLiteral(raw, field.TypeID)
				if ok {
					ordered := integerOrder && field.TypeID == string(semantic.BuiltinIntegerTypeID)
					appendRecordPredicates(&expressions, seen, root+"."+field.Name, literal, ordered)
				}
			}
		}
	}
	return expressions, nil
}

func appendRecordPredicates(expressions *[]string, seen map[string]bool, selector, literal string, integerOrder bool) {
	operators := []string{"==", "!="}
	if integerOrder {
		operators = []string{"<", "<=", ">", ">=", "==", "!="}
	}
	for _, operator := range operators {
		expression := selector + " " + operator + " " + literal
		if !seen[expression] {
			seen[expression] = true
			*expressions = append(*expressions, expression)
		}
	}
}

func (c recordFillGrammarContext) outputLiteralExpressions(maxExpressions int, typeID semantic.ID) ([]string, int, bool, error) {
	expressions := make([]string, 0)
	seen := map[string]bool{}
	for _, valueCase := range c.cases {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(valueCase.Expected), &fields); err != nil || fields == nil {
			return nil, 0, false, fmt.Errorf("record output literal case must be a JSON object")
		}
		for _, field := range c.output.Fields {
			if field.TypeID != string(typeID) {
				continue
			}
			raw, exists := fields[field.Name]
			if !exists {
				return nil, 0, false, fmt.Errorf("record output literal case is missing field %q", field.Name)
			}
			literal, ok := recordFillLiteral(raw, field.TypeID)
			if ok && !seen[literal] {
				seen[literal] = true
				expressions = append(expressions, literal)
			}
		}
	}
	return retainRecordFillExpressions(expressions, maxExpressions)
}

func decodeRecordFillInputTuple(raw string, expected int) ([]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var inputs []json.RawMessage
	if err := decoder.Decode(&inputs); err != nil || len(inputs) != expected {
		return nil, fmt.Errorf("record source-fill case requires %d positional inputs", expected)
	}
	return inputs, nil
}

func scalarTypeID(name string) string {
	switch name {
	case "Integer":
		return string(semantic.BuiltinIntegerTypeID)
	case "Boolean":
		return string(semantic.BuiltinBooleanTypeID)
	case "Text":
		return string(semantic.BuiltinStringTypeID)
	default:
		return ""
	}
}

func recordFillLiteral(raw json.RawMessage, typeID string) (string, bool) {
	switch typeID {
	case string(semantic.BuiltinStringTypeID):
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return "", false
		}
		return strconv.Quote(value), true
	case string(semantic.BuiltinIntegerTypeID):
		var value json.Number
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if decoder.Decode(&value) != nil {
			return "", false
		}
		if _, err := value.Int64(); err != nil {
			return "", false
		}
		return value.String(), true
	case string(semantic.BuiltinBooleanTypeID):
		if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
			return string(raw), true
		}
	}
	return "", false
}

func retainRecordFillExpressions(expressions []string, maxExpressions int) ([]string, int, bool, error) {
	enumerated := len(expressions)
	if enumerated < 2 {
		return nil, enumerated, false, fmt.Errorf("record source-fill grammar produced fewer than two distinct expressions")
	}
	if enumerated > maxExpressions {
		expressions = expressions[:maxExpressions]
	}
	return expressions, enumerated, len(expressions) == enumerated, nil
}
