package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func exportCandidate(root, output string, candidate syntaxregistration.Candidate, elapsed int64) error {
	if output == "" {
		return fmt.Errorf("a new external output directory is required")
	}
	input, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	input, err = filepath.Abs(input)
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return err
	}
	absolute = filepath.Join(parent, filepath.Base(absolute))
	relative, err := filepath.Rel(input, absolute)
	if err != nil {
		return err
	}
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return fmt.Errorf("candidate output must be outside the input repository")
	}
	if err := os.Mkdir(absolute, 0700); err != nil {
		return fmt.Errorf("candidate output must be new: %w", err)
	}
	for _, member := range candidate.Members {
		path := filepath.Join(absolute, "members", filepath.FromSlash(member.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		if err := os.WriteFile(path, member.Content, 0600); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(absolute, "candidate.json"), raw, 0600); err != nil {
		return err
	}
	observation := map[string]any{"operation": syntaxregistration.Operation,
		"emitted_members": candidate.Emitted, "required_members": candidate.Required,
		"wall_ms": elapsed, "repository_writes": 0, "semantic_admission": "UNASSESSED",
		"global_planner_admission": "NOT_IMPLEMENTED", "manual_follow_up_edits": nil}
	raw, err = json.MarshalIndent(observation, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(absolute, "execution.json"), raw, 0600)
}
