package languageassurance

import (
	"encoding/hex"
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID {
		return fmt.Errorf("language assurance report identity is malformed")
	}
	if len(report.Denominator) != 12 || len(report.Obligations) != 12 || len(report.MetaOperations) != 5 || len(report.Indicators) != 5 {
		return fmt.Errorf("language assurance report cardinality is malformed")
	}
	if report.ReportDigest == "" {
		return fmt.Errorf("language assurance report digest is missing")
	}
	expected, err := Evaluate(report.SubjectSHA, report.Transaction)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(report, expected) {
		return fmt.Errorf("language assurance report does not replay")
	}
	return nil
}

func ValidateForSubject(report Report, subjectSHA string) error {
	if err := Validate(report); err != nil {
		return err
	}
	if report.SubjectSHA != subjectSHA {
		return fmt.Errorf("language assurance report does not bind the exact subject")
	}
	return nil
}

func validateInput(subjectSHA string, transaction Transaction) error {
	if !validSHA(subjectSHA) {
		return fmt.Errorf("language assurance subject sha is malformed")
	}
	if transaction.Schema != TransactionSchema || transaction.TransactionID == "" {
		return fmt.Errorf("language assurance transaction identity is malformed")
	}
	rules := make(map[string]bool, len(transaction.AuthorityRoutes))
	for _, route := range transaction.AuthorityRoutes {
		if route.RuleID == "" || route.AuthoredBy == "" || route.PromotedBy == "" || rules[route.RuleID] {
			return fmt.Errorf("language assurance authority route is malformed")
		}
		rules[route.RuleID] = true
	}
	bindings := make(map[string]map[Role]bool, len(transaction.RoleBindings))
	for _, binding := range transaction.RoleBindings {
		if binding.Principal == "" || len(binding.Roles) == 0 || bindings[binding.Principal] != nil {
			return fmt.Errorf("language assurance role binding is malformed")
		}
		roles := make(map[Role]bool, len(binding.Roles))
		for _, role := range binding.Roles {
			if !validRole(role) || roles[role] {
				return fmt.Errorf("language assurance role binding is malformed")
			}
			roles[role] = true
		}
		bindings[binding.Principal] = roles
	}
	if len(transaction.AuthorityRoutes) > 0 && len(transaction.RoleBindings) > 0 {
		for _, route := range transaction.AuthorityRoutes {
			if !bindings[route.AuthoredBy][RoleContractAuthor] || !bindings[route.PromotedBy][RolePromoter] {
				return fmt.Errorf("language assurance authority route lacks role evidence")
			}
		}
	}
	decisions := make(map[string]bool, len(transaction.DecisionTransitions))
	for _, transition := range transaction.DecisionTransitions {
		if transition.ID == "" || decisions[transition.ID] || !validDecision(transition.Input) || !validDecision(transition.Output) {
			return fmt.Errorf("language assurance decision transition is malformed")
		}
		decisions[transition.ID] = true
	}
	return nil
}

func roleSet(roles []Role) map[Role]bool {
	result := make(map[Role]bool, len(roles))
	for _, role := range roles {
		result[role] = true
	}
	return result
}

func validRole(role Role) bool {
	switch role {
	case RoleContractAuthor, RoleImplementer, RoleEvaluatorAuthor, RoleAdapterAuthor, RolePolicyAdopter, RolePromoter, RoleAuditor:
		return true
	default:
		return false
	}
}

func validDecision(decision Decision) bool {
	switch decision {
	case DecisionUnknown, DecisionPass, DecisionFail, DecisionFixedPoint, DecisionAuthorized, DecisionAllow, DecisionBlock:
		return true
	default:
		return false
	}
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
