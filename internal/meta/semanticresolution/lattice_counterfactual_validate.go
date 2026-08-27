package semanticresolution

import "fmt"

func validateCounterfactuals(counterfactuals []LatticeCounterfactual) error {
	if len(counterfactuals) != LatticeCounterfactualDenominator {
		return fmt.Errorf("lattice counterfactual denominator is not fixed")
	}
	seen := map[string]bool{}
	for _, item := range counterfactuals {
		if seen[item.ID] || item.ID == "" || !item.SourceDigestChanged {
			return fmt.Errorf("counterfactual identity or source digest is invalid")
		}
		seen[item.ID] = true
		switch item.ID {
		case "observed-2-to-3":
			if item.Kind != "SEMANTIC" || !item.SemanticDigestChanged || !item.DecisionChanged || !item.ClaimTransitionChanged || item.BaselineDecision != DecisionUnknown || item.VariantDecision != DecisionPass || item.BaselineClaim.AfterState != ClaimOpen || item.VariantClaim.AfterState != ClaimDischarged {
				return fmt.Errorf("semantic counterfactual did not change decision and claim transition")
			}
		case "comment-only":
			if item.Kind != "NON_SEMANTIC" || item.SemanticDigestChanged || item.DecisionChanged || item.ClaimTransitionChanged {
				return fmt.Errorf("non-semantic counterfactual changed semantic output")
			}
		default:
			return fmt.Errorf("unknown counterfactual %q", item.ID)
		}
	}
	if !seen["observed-2-to-3"] || !seen["comment-only"] {
		return fmt.Errorf("counterfactual pair is incomplete")
	}
	return nil
}
