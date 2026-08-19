package main

import (
	"fmt"
	"strings"
)

func validateFailureProtectedPushBranches(branches []string) error {
	if len(branches) == 0 || !sameStrings(branches, []string{"dev", "main"}) {
		return fmt.Errorf("protected push owner registry is invalid")
	}
	for _, branch := range branches {
		if branch == "" || strings.ContainsAny(branch, "/*?[]") {
			return fmt.Errorf("protected push owner registry contains an invalid branch %q", branch)
		}
	}
	return nil
}
