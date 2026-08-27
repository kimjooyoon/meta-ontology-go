package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

const (
	fixedCaseContractTotal         = 5
	fixedCaseContractDenominatorID = "gooo://semantic-delta-receipt-denominator/v2"
	caseContractStage              = "suite-contract"
	caseContractStep               = "validate-case-inventory"
	caseContractExactReason        = "FIXED_CASE_ID_INVENTORY_EXACT"
	caseContractMismatchReason     = "FIXED_CASE_ID_INVENTORY_MISMATCH"
	caseContractMetaErrorReason    = "FIXED_CASE_META_RECIPE_UNAVAILABLE"
)

type fixedCaseRecipe struct {
	ID         string
	BeforePath string
	AfterPath  string
}

var fixedCaseRecipes = []fixedCaseRecipe{
	{ID: "equivalent", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/equivalent-after.gooo"},
	{ID: "semantic-change", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/semantic-after.gooo"},
	{ID: "value-program-change", BeforePath: "examples/semantic-delta-receipt/value-program-before.gooo", AfterPath: "examples/semantic-delta-receipt/value-program-after.gooo"},
	{ID: "indeterminate", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/indeterminate-after.gooo"},
	{ID: "ambiguous-match", BeforePath: "examples/semantic-delta-receipt/ambiguous-before.gooo", AfterPath: "examples/semantic-delta-receipt/ambiguous-after.gooo"},
}

func fixedCaseIDs() []string {
	ids := make([]string, 0, len(fixedCaseRecipes))
	for _, recipe := range fixedCaseRecipes {
		ids = append(ids, recipe.ID)
	}
	return ids
}

func observedCaseIDs(definitions []producer.CaseDefinition) []string {
	ids := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		ids = append(ids, definition.ID)
	}
	return ids
}

func observedRecipeIDs(meta producer.MetaContract) []string {
	ids := make([]string, 0, len(meta.CaseRecipes))
	for _, recipe := range meta.CaseRecipes {
		ids = append(ids, recipe.ID)
	}
	return ids
}

func exactCaseIDInventory(expected, observed []string) bool {
	if len(expected) != fixedCaseContractTotal || len(observed) != fixedCaseContractTotal || len(expected) != len(observed) {
		return false
	}
	counts := make(map[string]int, len(observed))
	for _, id := range observed {
		counts[id]++
	}
	for _, id := range expected {
		if counts[id] != 1 {
			return false
		}
		counts[id] = 0
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func exactCaseRecipeInventory(meta producer.MetaContract) bool {
	if len(meta.CaseRecipes) != fixedCaseContractTotal {
		return false
	}
	seen := make(map[string]bool, len(meta.CaseRecipes))
	for _, recipe := range meta.CaseRecipes {
		if seen[recipe.ID] {
			return false
		}
		seen[recipe.ID] = true
		matched := false
		for _, expected := range fixedCaseRecipes {
			if recipe.ID == expected.ID {
				matched = recipe.BeforePath == expected.BeforePath && recipe.AfterPath == expected.AfterPath
				break
			}
		}
		if !matched {
			return false
		}
	}
	ids := make([]string, 0, len(meta.CaseRecipes))
	for _, recipe := range meta.CaseRecipes {
		ids = append(ids, recipe.ID)
	}
	return exactCaseIDInventory(fixedCaseIDs(), ids)
}

func caseContractValid(definitions []producer.CaseDefinition, meta producer.MetaContract, metaErr error) bool {
	return metaErr == nil && producer.DenominatorID == fixedCaseContractDenominatorID && exactCaseDefinitionInventory(definitions) && exactCaseRecipeInventory(meta)
}

func exactCaseDefinitionInventory(definitions []producer.CaseDefinition) bool {
	if !exactCaseIDInventory(fixedCaseIDs(), observedCaseIDs(definitions)) {
		return false
	}
	for _, definition := range definitions {
		for _, expected := range fixedCaseRecipes {
			if definition.ID == expected.ID && (definition.BeforePath != expected.BeforePath || definition.AfterPath != expected.AfterPath) {
				return false
			}
		}
	}
	return true
}
