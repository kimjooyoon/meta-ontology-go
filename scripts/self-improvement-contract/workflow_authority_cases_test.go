package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContractWorkflowAuthority(t *testing.T) {
	source, branches := loadContractWorkflowAuthority(t)
	if err := validateContractWorkflowAuthority(source, branches); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(branches)
	if err != nil {
		t.Fatal(err)
	}
	branchLine := "    branches: " + string(encoded) + "\n"
	cases := []struct {
		name   string
		source string
	}{
		{"unrestricted push", strings.Replace(source, branchLine, "", 1)},
		{"different branches", strings.Replace(source, branchLine, "    branches: [feature]\n", 1)},
		{"missing tags", strings.Replace(source, "    tags: ['**']\n", "", 1)},
		{"restricted tags", strings.Replace(source, "    tags: ['**']", "    tags: ['v*']", 1)},
		{"missing PR", strings.Replace(source, "  pull_request:\n", "", 1)},
		{"PR path filter", strings.Replace(source, "  pull_request:\n", "  pull_request:\n    paths: [README.md]\n", 1)},
		{"write permissions", strings.Replace(source, "  contents: read", "  contents: write", 1)},
		{"second event root", source + "\non:\n  push:\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := validateContractWorkflowAuthority(test.source, branches); err == nil {
				t.Fatal("workflow authority drift was accepted")
			}
		})
	}
	for _, invalid := range [][]string{nil, {"dev", "dev"}, {"*"}} {
		if _, err := contractWorkflowHeader(invalid); err == nil {
			t.Fatal("invalid protected branch authority was accepted")
		}
	}
}

func loadContractWorkflowAuthority(t *testing.T) (string, []string) {
	t.Helper()
	root := filepath.Join("..", "..", ".github")
	source, err := os.ReadFile(filepath.Join(root, "workflows", "self-improvement-contract.yml"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := os.ReadFile(filepath.Join(root, "ci-governance.json"))
	if err != nil {
		t.Fatal(err)
	}
	var governance struct {
		ProtectedPushBranches []string `json:"protected_push_branches"`
	}
	if err := json.Unmarshal(policy, &governance); err != nil {
		t.Fatal(err)
	}
	return string(source), governance.ProtectedPushBranches
}
