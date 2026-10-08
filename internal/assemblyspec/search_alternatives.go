package assemblyspec

import "fmt"

func (s Spec) validateSearchAlternatives() error {
	if len(s.SearchAlternatives) == 0 {
		return nil
	}
	if s.Search == nil || s.FillPlan != nil || len(s.SearchAlternatives) > 4 {
		return fmt.Errorf("search alternatives require an integer search and at most four alternatives")
	}
	ids, settings := map[string]bool{}, map[string]bool{}
	for _, alternative := range s.SearchAlternatives {
		key := fmt.Sprintf("%s/%d", alternative.Grammar, alternative.MaxCandidates)
		if !identifier(alternative.ID) || ids[alternative.ID] || settings[key] ||
			!supportedSearchGrammar(alternative.Grammar) ||
			alternative.MaxCandidates < 2 || alternative.MaxCandidates > 16 {
			return fmt.Errorf("search alternative requires a unique ID and grammar/bound pair, supported grammar and 2..16 candidates")
		}
		ids[alternative.ID], settings[key] = true, true
	}
	return nil
}

func supportedSearchGrammar(grammar string) bool {
	return grammar == "integer-offset-constant/v1" || grammar == "integer-hole-residual/v1" ||
		grammar == "integer-hole-quadratic/v1"
}
