package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func runtimePolicy() []Rule {
	return []Rule{
		{Action: "actions/checkout", MinimumMajor: 5, Runtime: "node24",
			Evidence: "https://github.com/actions/checkout#checkout-v5"},
		{Action: "actions/download-artifact", MinimumMajor: 7, Runtime: "node24",
			AllowedInputs: downloadInputs(),
			Evidence:      "https://github.com/actions/download-artifact/releases/tag/v7.0.0"},
		{Action: "actions/github-script", MinimumMajor: 8, Runtime: "node24",
			Evidence: "https://github.com/actions/github-script#v8"},
		{Action: "actions/setup-go", MinimumMajor: 6, Runtime: "node24",
			Evidence: "https://github.com/actions/setup-go#breaking-changes-in-v6"},
		{Action: "actions/upload-artifact", MinimumMajor: 6, Runtime: "node24",
			Evidence: "https://github.com/actions/upload-artifact/releases/tag/v6.0.0"},
	}
}

func downloadInputs() []string {
	return []string{
		"artifact-ids", "github-token", "merge-multiple", "name",
		"path", "pattern", "repository", "run-id",
	}
}

func policyDigest(policy []Rule) string {
	data, _ := json.Marshal(policy)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
