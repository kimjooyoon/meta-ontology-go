package metarecognition

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func validateReplay(document ReplayDocument) error {
	if document.Schema != SchemaVersion {
		return fmt.Errorf("replay schema %q is not %q", document.Schema, SchemaVersion)
	}
	caseIDs := make(map[string]struct{}, len(document.Cases))
	for _, value := range document.Cases {
		if value.ID == "" {
			return fmt.Errorf("replay case has empty id")
		}
		if _, exists := caseIDs[value.ID]; exists {
			return fmt.Errorf("duplicate replay case %q", value.ID)
		}
		caseIDs[value.ID] = struct{}{}
		if !value.Subject.Valid() {
			return fmt.Errorf("replay case %q has invalid subject %q", value.ID, value.Subject)
		}
		if value.Source == "" || path.IsAbs(value.Source) || value.Source == "." || strings.Contains(value.Source, "\\") ||
			path.Clean(value.Source) != value.Source || strings.HasPrefix(value.Source, "../") || value.Source == ".." {
			return fmt.Errorf("replay case %q has invalid source %q", value.ID, value.Source)
		}
		if err := uniqueValues(value.ID, "root", value.Roots); err != nil {
			return err
		}
		if err := uniqueValues(value.ID, "command", value.Commands); err != nil {
			return err
		}
		if err := uniqueValues(value.ID, "path", value.Paths); err != nil {
			return err
		}
	}
	return nil
}
func canonicalRootRelativePath(workspaceRoot, sourcePath string) (string, error) {
	root, err := canonicalPhysicalPath(workspaceRoot, "workspace root")
	if err != nil {
		return "", err
	}
	source, err := canonicalPhysicalPath(sourcePath, "source path")
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(filepath.FromSlash(root), filepath.FromSlash(source))
	if err != nil {
		return "", fmt.Errorf("source path %q escapes workspace root %q", sourcePath, workspaceRoot)
	}
	relative = filepath.ToSlash(relative)
	if relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
		return "", fmt.Errorf("source path %q escapes workspace root %q", sourcePath, workspaceRoot)
	}
	return relative, nil
}
func canonicalPhysicalPath(value, label string) (string, error) {
	cleaned := strings.ReplaceAll(value, "\\", "/")
	if cleaned == "" || !path.IsAbs(cleaned) || strings.Contains(cleaned, "//") {
		return "", fmt.Errorf("%s %q is not an unambiguous absolute path", label, value)
	}
	parts := strings.Split(cleaned, "/")
	for _, part := range parts[1:] {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("%s %q contains an ambiguous component", label, value)
		}
	}
	return path.Clean(cleaned), nil
}
