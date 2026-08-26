package languageassurance

import (
	"encoding/hex"
	"fmt"
	"slices"
)

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
	if !validSnapshotBindings(transaction.SnapshotBindings) {
		return fmt.Errorf("language assurance snapshot binding is malformed")
	}
	if !validRawReconstructions(transaction.RawReconstructions) {
		return fmt.Errorf("language assurance raw reconstruction is malformed")
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
	return slices.Contains([]Role{RoleContractAuthor, RoleImplementer, RoleEvaluatorAuthor, RoleAdapterAuthor, RolePolicyAdopter, RolePromoter, RoleAuditor}, role)
}

func validSHA(value string) bool {
	_, err := hex.DecodeString(value)
	return len(value) == 40 && err == nil
}
