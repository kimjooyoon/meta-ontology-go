package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func profileOutputDirectory(root, output string) (string, error) {
	if output == "" {
		return "", fmt.Errorf("a new external output directory is required")
	}
	input, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	input, err = filepath.Abs(input)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(output)
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", err
	}
	absolute = filepath.Join(parent, filepath.Base(absolute))
	relative, err := filepath.Rel(input, absolute)
	if err != nil {
		return "", err
	}
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return "", fmt.Errorf("candidate output must be outside the input repository")
	}
	if err := os.Mkdir(absolute, 0700); err != nil {
		return "", fmt.Errorf("candidate output must be new: %w", err)
	}
	return absolute, nil
}

func exportProfileCandidate(root, directory string, candidate syntaxregistration.ProfileMigrationCandidate,
	output io.Writer) error {
	for _, member := range candidate.Members {
		if !fs.ValidPath(member.Path) {
			return fmt.Errorf("invalid source member path: %s", member.Path)
		}
	}
	absolute, err := profileOutputDirectory(root, directory)
	if err != nil {
		return err
	}
	for _, member := range candidate.Members {
		for _, item := range []struct {
			name string
			raw  []byte
		}{{"before", member.BeforeContent}, {"members", member.Content}} {
			path := filepath.Join(absolute, item.name, filepath.FromSlash(member.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				return err
			}
			if err := os.WriteFile(path, item.raw, 0600); err != nil {
				return err
			}
		}
	}
	raw, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(absolute, "candidate.json"), append(raw, '\n'), 0600); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(map[string]any{"status": "PROPOSAL_EXPORTED",
		"output": absolute, "required_members": candidate.RequiredMembers, "emitted_members": len(candidate.Members),
		"case_population": candidate.CasePopulation, "repository_writes": 0, "semantic_admission": candidate.Admission,
		"apply_authorized": false, "promotion_authorized": false})
}
