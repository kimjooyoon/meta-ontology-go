package languagedelivery

import (
	"fmt"
	"reflect"
)

func ValidateContract(contract Contract) error {
	want := CanonicalContract()
	if !reflect.DeepEqual(contract, want) {
		return fmt.Errorf("LANGUAGE_DELIVERY_CONTRACT_V1_DRIFT")
	}
	counts := map[Audience]int{}
	ids := map[string]bool{}
	for _, item := range contract.Obligations {
		if ids[item.ID] {
			return fmt.Errorf("duplicate obligation %q", item.ID)
		}
		ids[item.ID] = true
		counts[item.Audience]++
		if item.MetaOperation == "" {
			return fmt.Errorf("obligation %q has no meta operation", item.ID)
		}
	}
	if len(contract.Obligations) != 36 {
		return fmt.Errorf("fixed denominator is %d, want 36", len(contract.Obligations))
	}
	for _, audience := range audienceOrder {
		if counts[audience] != 12 {
			return fmt.Errorf("audience %s denominator is %d, want 12", audience, counts[audience])
		}
	}
	return nil
}
