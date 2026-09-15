package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func validateFailureOwnerRegistryDocument(data []byte, branch string) error {
	var registry failureOwnerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("parse failure owner registry: %w", err)
	}
	if registry.Schema != "gooo/ci-governance/v2" || !validFailureOwnerBranch(branch) || branch == "dev" || branch == "main" {
		return fmt.Errorf("failure owner registry is invalid")
	}
	if err := validateFailureProtectedPushBranches(registry.ProtectedPushBranches); err != nil {
		return err
	}
	seen := make(map[string]bool, len(registry.Ownership))
	var matched bool
	for _, owner := range registry.Ownership {
		if !validFailureOwnerBranch(owner.Branch) || owner.Branch == "dev" || owner.Branch == "main" {
			return fmt.Errorf("failure owner registry contains an invalid protected or wildcard branch %q", owner.Branch)
		}
		if seen[owner.Branch] {
			return fmt.Errorf("failure owner registry contains duplicate branch %q", owner.Branch)
		}
		seen[owner.Branch] = true
		if len(owner.Paths) == 0 {
			return fmt.Errorf("failure owner registry branch %q has no paths", owner.Branch)
		}
		pathSeen := make(map[string]bool, len(owner.Paths))
		for _, path := range owner.Paths {
			if !validFailureOwnerPath(path) {
				return fmt.Errorf("failure owner registry branch %q has malformed path %q", owner.Branch, path)
			}
			if pathSeen[path] {
				return fmt.Errorf("failure owner registry branch %q has duplicate path %q", owner.Branch, path)
			}
			pathSeen[path] = true
		}
		if owner.Branch == branch {
			if matched {
				return fmt.Errorf("failure owner registry contains duplicate branch %q", branch)
			}
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("failure owner branch %q is not registered for dev pull requests", branch)
	}
	return nil
}
func validFailureOwnerBranch(branch string) bool {
	if branch == "" || branch != strings.TrimSpace(branch) || !strings.HasPrefix(branch, "agent/") || strings.ContainsAny(branch, " ~^:?*[]\\\r\n") || strings.Contains(branch, "//") || strings.Contains(branch, "..") || strings.Contains(branch, "@{") || strings.HasSuffix(branch, "/") || strings.HasSuffix(branch, ".") || strings.HasSuffix(branch, ".lock") {
		return false
	}
	return true
}
