package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// This checks the workflow's deliberately literal header, not arbitrary YAML.
// Protected branch authority comes from the existing meta-governance registry.
func contractWorkflowHeader(branches []string) (string, error) {
	if len(branches) == 0 {
		return "", fmt.Errorf("protected push branches are missing")
	}
	seen := map[string]bool{}
	for _, branch := range branches {
		if branch == "" || strings.ContainsAny(branch, "!*?[]\\\r\n") || seen[branch] {
			return "", fmt.Errorf("protected push branches must be distinct literal branch names")
		}
		seen[branch] = true
	}
	encoded, err := json.Marshal(branches)
	if err != nil {
		return "", err
	}
	header := "name: Self-improvement contract\n\non:\n  push:\n"
	header += "    branches: " + string(encoded) + "\n    tags: ['**']\n  pull_request:\n"
	header += "\npermissions:\n  contents: read\n\njobs:\n"
	return header, nil
}

func validateContractWorkflowAuthority(source string, branches []string) error {
	header, err := contractWorkflowHeader(branches)
	if err != nil {
		return err
	}
	body, matched := strings.CutPrefix(source, header)
	if !matched {
		return fmt.Errorf("workflow must preserve all PRs, protected branch pushes, tag pushes and read-only permissions")
	}
	// Do not let a later root declaration replace the checked event contract.
	for line := range strings.SplitSeq(body, "\n") {
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			return fmt.Errorf("unexpected root declaration after the workflow contract")
		}
	}
	return nil
}
