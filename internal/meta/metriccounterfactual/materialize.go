package metriccounterfactual

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Materialize(root string, manifest Manifest) error {
	if manifest.Schema != ManifestSchema || !ValidManifest(manifest) {
		return fmt.Errorf("invalid manifest")
	}
	seen := make(map[string]bool, len(manifest.Files))
	for _, file := range manifest.Files {
		if seen[file.Path] || languageForPath(file.Path) != file.Language {
			return fmt.Errorf("invalid manifest file %q", file.Path)
		}
		seen[file.Path] = true
		native, err := safeNative(root, file.Path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(native), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(native, []byte(file.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func ApplyPlan(root string, plan Plan) ([]Receipt, error) {
	if plan.Schema != PlanSchema || !ValidPlan(plan) {
		return nil, fmt.Errorf("invalid plan")
	}
	receipts := make([]Receipt, 0, len(plan.Mutations))
	for _, mutation := range plan.Mutations {
		receipt, err := applyMutation(root, mutation)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func applyMutation(root string, mutation Mutation) (Receipt, error) {
	native, err := safeNative(root, mutation.Path)
	if err != nil {
		return Receipt{}, err
	}
	before := []byte(nil)
	switch mutation.Kind {
	case "APPEND":
		before, err = os.ReadFile(native)
	case "CREATE":
		if _, statErr := os.Lstat(native); statErr == nil {
			err = fmt.Errorf("create target exists: %s", mutation.Path)
		} else if !os.IsNotExist(statErr) {
			err = statErr
		}
	default:
		err = fmt.Errorf("unsupported mutation %q", mutation.Kind)
	}
	if err != nil {
		return Receipt{}, err
	}
	after := append(append([]byte(nil), before...), []byte(mutation.Content)...)
	if err := os.MkdirAll(filepath.Dir(native), 0o755); err != nil {
		return Receipt{}, err
	}
	if err := os.WriteFile(native, after, 0o644); err != nil {
		return Receipt{}, err
	}
	beforeDigest := "ABSENT"
	if mutation.Kind != "CREATE" {
		beforeDigest = contentDigest(before)
	}
	return Receipt{
		Kind: mutation.Kind, Path: mutation.Path,
		BeforeDigest: beforeDigest, AfterDigest: contentDigest(after),
		BeforeLines: countLines(before), AfterLines: countLines(after),
	}, nil
}

func safeNative(root, relative string) (string, error) {
	if relative == "" || strings.Contains(relative, "\\") ||
		strings.HasPrefix(relative, "/") || path.Clean(relative) != relative {
		return "", fmt.Errorf("unsafe fixture path %q", relative)
	}
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func contentDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}
