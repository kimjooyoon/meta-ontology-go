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
	recordFieldPredicateGrammar         = "record-field-predicate/v1"
	recordFieldPredicateV2Grammar       = "record-field-predicate/v2"
	recordPredicateCompositionGrammar   = "record-field-predicate-composition/v1"
	recordPredicateCompositionV2Grammar = "record-field-predicate-composition/v2"
	recordStringLiteralGrammar          = "record-string-literal/v1"
	recordIntegerLiteralGrammar         = "record-integer-literal/v1"
	recordBooleanLiteralGrammar         = "record-boolean-literal/v1"
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
