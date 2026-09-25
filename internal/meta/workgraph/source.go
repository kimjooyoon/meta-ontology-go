package workgraph

import "strings"

func sourceBinding(contract Contract, source string) (bool, string) {
	required := []string{
		"package workgraph",
		"namespace workgraph",
		"entity " + contract.Claim.Entity + " id \"" + contract.Claim.ID + "\"",
		"activity LowerProjectResolution(",
	}
	for _, gate := range contract.Gates {
		required = append(required, "activity "+gate.Activity+"(")
	}
	for _, declaration := range required {
		if !strings.Contains(source, declaration) {
			return false, declaration
		}
	}
	return true, ""
}
