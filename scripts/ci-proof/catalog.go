package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const failureOwnerRegistryPath = ".github/ci-governance.json"
const promotionOwnerBindingCode = "CI-PROMOTION-OWNER-BINDING-001"

type catalogDocumentEntry struct {
	Code            string
	Class           string
	Severity        string
	BlockingScope   string
	Parallelizable  bool
	HandoffRequired bool
	Owner           string
}

type failureOwnerRegistry struct {
	Schema                string   `json:"schema"`
	ProtectedPushBranches []string `json:"protected_push_branches"`
	Ownership             []struct {
		Branch string   `json:"branch"`
		Paths  []string `json:"paths"`
	} `json:"ownership"`
}

var failureCatalogDigest, failureCatalogDigestErr = loadFailureCatalogDigest()

func loadFailureCatalogDigest() (string, error) {
	data, err := readFailureFile(failureCatalogPath)
	if err != nil {
		return "", fmt.Errorf("read failure catalog: %w", err)
	}
	return "sha256:" + digestBytes(data), nil
}

func readFailureFile(name string) ([]byte, error) {
	candidates := []string{name, filepath.Join("..", name), filepath.Join("..", "..", name)}
	var lastErr error
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func validateFailureCatalog() error {
	if failureCatalogDigestErr != nil {
		return failureCatalogDigestErr
	}
	data, err := readFailureFile(failureCatalogPath)
	if err != nil {
		return fmt.Errorf("read failure catalog: %w", err)
	}
	if want := "sha256:" + digestBytes(data); want != failureCatalogDigest {
		return fmt.Errorf("failure catalog digest changed while validating")
	}
	entries, err := parseCatalogDocument(string(data))
	if err != nil {
		return err
	}
	if len(entries) != len(failureCatalogRecords) {
		return fmt.Errorf("failure catalog metadata count mismatch")
	}
	for _, record := range failureCatalogRecords {
		entry, ok := entries[record.Code]
		if !ok || entry.Class != record.Entry.Class || entry.Severity != record.Entry.Severity || entry.BlockingScope != record.Entry.BlockingScope || entry.Parallelizable != record.Entry.Parallelizable || entry.HandoffRequired != record.Entry.HandoffRequired || entry.Owner != record.Entry.Owner {
			return fmt.Errorf("failure catalog metadata mismatch for %s", record.Code)
		}
	}
	return nil
}

func parseCatalogDocument(document string) (map[string]catalogDocumentEntry, error) {
	const prefix = "<!-- machine-catalog: "
	const suffix = " -->"
	entries := make(map[string]catalogDocumentEntry)
	for _, line := range strings.Split(document, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		if !strings.HasSuffix(line, suffix) {
			return nil, fmt.Errorf("malformed failure catalog metadata")
		}
		fields := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix), "|")
		if len(fields) != 7 || fields[0] == "" {
			return nil, fmt.Errorf("failure catalog metadata must contain seven fields")
		}
		parallelizable, err := strconv.ParseBool(fields[4])
		if err != nil {
			return nil, fmt.Errorf("invalid catalog parallelization for %s", fields[0])
		}
		handoffRequired, err := strconv.ParseBool(fields[5])
		if err != nil {
			return nil, fmt.Errorf("invalid catalog handoff flag for %s", fields[0])
		}
		if _, exists := entries[fields[0]]; exists {
			return nil, fmt.Errorf("duplicate failure catalog metadata for %s", fields[0])
		}
		entries[fields[0]] = catalogDocumentEntry{Code: fields[0], Class: fields[1], Severity: fields[2], BlockingScope: fields[3], Parallelizable: parallelizable, HandoffRequired: handoffRequired, Owner: fields[6]}
	}
	return entries, nil
}

func validateFailureOwnerRegistry(branch string) error {
	data, err := readFailureFile(failureOwnerRegistryPath)
	if err != nil {
		return fmt.Errorf("read failure owner registry: %w", err)
	}
	var registry failureOwnerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("parse failure owner registry: %w", err)
	}
	if registry.Schema != "gooo/ci-governance/v1" || branch == "" || strings.ContainsAny(branch, "*?[]") {
		return fmt.Errorf("failure owner registry is invalid")
	}
	var matched bool
	for _, owner := range registry.Ownership {
		if owner.Branch != branch {
			continue
		}
		if matched || !sameCatalogPaths(owner.Paths, []string{".github/**", "scripts/**", "internal/verify/**"}) {
			return fmt.Errorf("failure owner registry is duplicated or outside CI scope")
		}
		matched = true
	}
	if !matched {
		return fmt.Errorf("failure owner branch %q is not registered for CI scope", branch)
	}
	return nil
}

func validateFailureOwnerBinding(binding failureBinding) error {
	if binding.Event == "pull_request" {
		if binding.BaseRef != "dev" && binding.BaseRef != "main" {
			return fmt.Errorf("%s: pull request base branch is unsupported", promotionOwnerBindingCode)
		}
		if binding.BaseRef == "main" {
			if binding.OwnerBranch != "dev" {
				return fmt.Errorf("%s: main promotion owner must be exact dev", promotionOwnerBindingCode)
			}
			if err := validateProtectedPushOwnerRegistry(binding.OwnerBranch); err != nil {
				return fmt.Errorf("%s: %w", promotionOwnerBindingCode, err)
			}
			return nil
		}
		if binding.OwnerBranch == "dev" {
			return fmt.Errorf("%s: dev-to-dev pull request is not a feature or promotion route", promotionOwnerBindingCode)
		}
		return validateFailureOwnerRegistry(binding.OwnerBranch)
	}
	if binding.Event != "push" || binding.PRNumber != 0 || binding.OwnerBranch != binding.BaseRef || binding.EventRef != "refs/heads/"+binding.BaseRef {
		return fmt.Errorf("protected push owner must equal the exact protected base branch")
	}
	return validateProtectedPushOwnerRegistry(binding.BaseRef)
}

func validateProtectedPushOwnerRegistry(branch string) error {
	data, err := readFailureFile(failureOwnerRegistryPath)
	if err != nil {
		return fmt.Errorf("read protected push owner registry: %w", err)
	}
	var registry failureOwnerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("parse protected push owner registry: %w", err)
	}
	if registry.Schema != "gooo/ci-governance/v1" || len(registry.ProtectedPushBranches) == 0 || !sameStrings(registry.ProtectedPushBranches, []string{"dev", "main"}) {
		return fmt.Errorf("protected push owner registry is invalid")
	}
	for _, protectedBranch := range registry.ProtectedPushBranches {
		if protectedBranch == branch {
			return nil
		}
	}
	return fmt.Errorf("protected push branch %q is not registered", branch)
}

func sameCatalogPaths(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]bool, len(left))
	for _, path := range left {
		if seen[path] {
			return false
		}
		seen[path] = true
	}
	for _, path := range right {
		if !seen[path] {
			return false
		}
	}
	return true
}
