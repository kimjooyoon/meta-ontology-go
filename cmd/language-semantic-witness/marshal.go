package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic"
)

func marshal(report languagesemantic.Report) ([]byte, error) {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
