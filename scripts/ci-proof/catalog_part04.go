package main

import (
	"encoding/json"
	"fmt"
	pathpkg "path"
	"slices"
	"strings"
)

func validFailureOwnerPath(path string) bool {
	if path == "" || path != strings.TrimSpace(path) || strings.Contains(path, "\\") || strings.HasPrefix(path, "/") {
		return false
	}
	suffix := ""
	base := path
	if strings.HasSuffix(path, "/**") {
		suffix = "/**"
		base = strings.TrimSuffix(path, suffix)
	}
	if base == "" || strings.ContainsAny(base, "*?[]") || (suffix == "" && strings.ContainsAny(path, "*?[]")) {
		return false
	}
	cleaned := pathpkg.Clean(base)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != base {
		return false
	}
	return path == base+suffix
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
	if registry.Schema != "gooo/ci-governance/v2" {
		return fmt.Errorf("protected push owner registry is invalid")
	}
	if err := validateFailureProtectedPushBranches(registry.ProtectedPushBranches); err != nil {
		return err
	}
	if slices.Contains(registry.ProtectedPushBranches, branch) {
		return nil
	}
	return fmt.Errorf("protected push branch %q is not registered", branch)
}
