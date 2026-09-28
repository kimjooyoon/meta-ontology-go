package main

import (
	"fmt"
	"strings"
)

func validFailureHeadBranch(branch string) bool {
	if branch == "" || branch != strings.TrimSpace(branch) || !strings.HasPrefix(branch, "agent/") || strings.ContainsAny(branch, " ~^:?*[]\\\r\n") || strings.Contains(branch, "//") || strings.Contains(branch, "..") || strings.Contains(branch, "@{") || strings.HasSuffix(branch, "/") || strings.HasSuffix(branch, ".") || strings.HasSuffix(branch, ".lock") {
		return false
	}
	return true
}

func validateFailureBranchBinding(binding failureBinding) error {
	if binding.Event == "pull_request" {
		if binding.BaseRef != "dev" && binding.BaseRef != "main" {
			return fmt.Errorf("%s: pull request base branch is unsupported", promotionBranchBindingCode)
		}
		if binding.BaseRef == "main" && binding.HeadBranch != "dev" {
			return fmt.Errorf("%s: main promotion head must be exact dev", promotionBranchBindingCode)
		}
		if binding.BaseRef == "dev" && (binding.HeadBranch == "dev" || !validFailureHeadBranch(binding.HeadBranch)) {
			return fmt.Errorf("%s: feature pull request head is invalid", promotionBranchBindingCode)
		}
		return nil
	}
	if binding.Event != "push" || binding.PRNumber != 0 || binding.HeadBranch != binding.BaseRef || binding.EventRef != "refs/heads/"+binding.BaseRef {
		return fmt.Errorf("protected push identity must match its exact branch ref")
	}
	if binding.BaseRef != "dev" && binding.BaseRef != "main" {
		return fmt.Errorf("protected push branch is unsupported")
	}
	return nil
}
