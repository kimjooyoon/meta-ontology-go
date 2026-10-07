package main

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

func readRecordAssemblyPolicy(flags map[string]string) (*bodycodegen.RecordAssemblyPolicy, error) {
	if flags["--assembly-policy"] == "" {
		return nil, nil
	}
	source, err := readBodyExecutionFile(flags["--assembly-policy"], 128<<10)
	if err != nil {
		return nil, err
	}
	return &bodycodegen.RecordAssemblyPolicy{Source: string(source), Activity: flags["--policy-activity"]}, nil
}
