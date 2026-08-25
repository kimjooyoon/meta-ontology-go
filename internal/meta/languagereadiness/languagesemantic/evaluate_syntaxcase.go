package languagesemantic

import (
	"fmt"
	"path/filepath"
)

type syntaxCase struct {
	Definition struct {
		ID   string `json:"id"`
		Path string `json:"path"`
		Kind string `json:"kind"`
	} `json:"definition"`
	Evidence struct {
		ObservedDecision string   `json:"observed_decision"`
		Diagnostics      []string `json:"diagnostics"`
	} `json:"evidence"`
	Status string `json:"status"`
}

func validateSyntaxCases(cases []syntaxCase) error {
	if len(cases) != expectedSyntaxCases {
		return fmt.Errorf("syntax evidence contains %d cases, want %d", len(cases), expectedSyntaxCases)
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	valid, invalid := 0, 0
	for _, item := range cases {
		path := filepath.ToSlash(filepath.Clean(item.Definition.Path))
		if item.Definition.ID == "" || path == "." || ids[item.Definition.ID] || paths[path] {
			return fmt.Errorf("syntax evidence case identity is incomplete or duplicated")
		}
		ids[item.Definition.ID], paths[path] = true, true
		switch item.Definition.Kind {
		case "VALID":
			valid++
		case "INVALID":
			invalid++
		default:
			return fmt.Errorf("syntax evidence case kind is unknown")
		}
	}
	if valid != expectedSyntaxValid || invalid != expectedSyntaxInvalid {
		return fmt.Errorf("syntax evidence case partition is %d/%d, want %d/%d",
			valid, invalid, expectedSyntaxValid, expectedSyntaxInvalid)
	}
	return nil
}
