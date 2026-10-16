package main

import (
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/recordbridge"
)

type targetField struct{ name, kind string }

func targetFields(path, name string) (map[string]targetField, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fields, err := recordbridge.Fields(path, raw, name)
	if err != nil {
		return nil, err
	}
	result := make(map[string]targetField, len(fields))
	for id, field := range fields {
		result[id] = targetField{name: field.Name, kind: field.Kind}
	}
	return result, nil
}
