package toolchainrelease

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// observeTrackedModuleReferences discovers exact current-version bytes in the
// Git index's working-tree files. Unknown mentions stay unresolved; the scanner
// does not infer that a release note or test fixture may safely be rewritten.
func observeTrackedModuleReferences(ctx context.Context, root, current string,
	result *DirectModuleImpact) error {
	started := time.Now()
	if current == "" {
		return fmt.Errorf("current module version is empty")
	}
	command := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z", "--cached")
	listing, err := command.Output()
	if err != nil || len(listing) == 0 || listing[len(listing)-1] != 0 {
		return fmt.Errorf("complete Git tracked-path listing is unavailable: %v", err)
	}
	result.ScanScope = "git-tracked-working-tree-regular-files/v1"
	result.TrackedPathsSHA = payloadDigest(listing)
	declared := make(map[string]bool, len(result.Outputs))
	for _, name := range result.Outputs {
		declared[name] = true
	}
	seen := make(map[string]bool)
	for _, item := range bytes.Split(listing[:len(listing)-1], []byte{0}) {
		name := string(item)
		if name == "" || seen[name] || path.IsAbs(name) || path.Clean(name) != name ||
			name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return fmt.Errorf("invalid or repeated tracked path: %q", name)
		}
		seen[name] = true
		if err := observeTrackedReferenceFile(ctx, root, name, current, declared, result); err != nil {
			return err
		}
	}
	sort.Strings(result.OutsideScope)
	result.ScanElapsedUS = time.Since(started).Microseconds()
	return nil
}

func observeTrackedReferenceFile(ctx context.Context, root, name, current string,
	declared map[string]bool, result *DirectModuleImpact) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("tracked path is not a readable regular file: %s", name)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read tracked path %s: %w", name, err)
	}
	result.TrackedFiles++
	if before, exists := result.Inputs[name]; exists && before != payloadDigest(raw) {
		return fmt.Errorf("tracked preimage changed during observation: %s", name)
	}
	if !bytes.Contains(raw, []byte(current)) {
		return nil
	}
	result.ReferenceFiles++
	if !declared[name] {
		result.OutsideScope = append(result.OutsideScope, name)
		result.Inputs[name] = payloadDigest(raw)
	}
	return nil
}
