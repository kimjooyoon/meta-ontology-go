package query

import (
	"fmt"
)

type derivedRule struct {
	id         DerivedRuleID
	predicate  DerivedRelation
	base       Relation
	inverse    bool
	transitive bool
}
type derivedKey struct {
	subject   ID
	predicate DerivedRelation
	object    ID
}

func ParseDerivedRule(raw DerivedRuleID) (DerivedRuleID, error) {
	switch raw {
	case RuleUsedBy, RuleGeneratedBy, RuleDerivedTo, RuleDependsOn:
		return raw, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedDerivedRule, raw)
	}
}
func ruleDefinition(ruleID DerivedRuleID) (derivedRule, error) {
	switch ruleID {
	case RuleUsedBy:
		return derivedRule{ruleID, DerivedUsedBy, Used, true, false}, nil
	case RuleGeneratedBy:
		return derivedRule{ruleID, DerivedGeneratedBy, WasGeneratedBy, true, false}, nil
	case RuleDerivedTo:
		return derivedRule{ruleID, DerivedTo, WasDerivedFrom, true, false}, nil
	case RuleDependsOn:
		return derivedRule{ruleID, DerivedDependsOn, WasDerivedFrom, false, true}, nil
	default:
		return derivedRule{}, fmt.Errorf("%w: %q", ErrUnsupportedDerivedRule, ruleID)
	}
}
