package toolchainrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type historicalRoots struct {
	Releases     string `json:"releases"`
	Observations string `json:"observations"`
}

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
	roots, err := readHistoricalRoots(ctx, root, result)
	if err != nil {
		return err
	}
	result.ArchiveRoots = []string{roots.Releases, roots.Observations}
	head, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil || len(strings.TrimSpace(string(head))) != 40 || !isHex(strings.TrimSpace(string(head))) {
		return fmt.Errorf("archive HEAD identity is unavailable: %v", err)
	}
	result.ArchiveHead = strings.TrimSpace(string(head))
	declared := make(map[string]bool, len(result.Outputs))
	for _, name := range result.Outputs {
		declared[name] = true
	}
	seen := make(map[string]bool)
	for item := range bytes.SplitSeq(listing[:len(listing)-1], []byte{0}) {
		name := string(item)
		if name == "" || seen[name] || path.IsAbs(name) || path.Clean(name) != name ||
			name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return fmt.Errorf("invalid or repeated tracked path: %q", name)
		}
		seen[name] = true
		if err := observeTrackedReferenceFile(ctx, root, name, current, declared, roots, result); err != nil {
			return err
		}
	}
	sort.Strings(result.OutsideScope)
	sort.Strings(result.HistoricalScope)
	result.ScanElapsedUS = time.Since(started).Microseconds()
	return nil
}

func observeTrackedReferenceFile(ctx context.Context, root, name, current string,
	declared map[string]bool, roots historicalRoots, result *DirectModuleImpact) error {
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
		result.Inputs[name] = payloadDigest(raw)
		if strings.HasPrefix(name, roots.Releases) || strings.HasPrefix(name, roots.Observations) {
			committed, err := exec.CommandContext(ctx, "git", "-C", root, "show",
				result.ArchiveHead+":"+name).Output()
			if err == nil && bytes.Equal(committed, raw) {
				result.HistoricalScope = append(result.HistoricalScope, name)
				return nil
			}
		}
		result.OutsideScope = append(result.OutsideScope, name)
	}
	return nil
}

func readHistoricalRoots(ctx context.Context, root string, result *DirectModuleImpact) (historicalRoots, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil || payloadDigest(source) != result.Inputs[directModuleSourcePath] {
		return historicalRoots{}, fmt.Errorf("historical root source preimage changed")
	}
	input, _ := json.Marshal([]any{map[string]any{"module_path": result.Module}})
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"HistoricalRoots", input)
	if err != nil || value.ActivityID != "urn:gooo:direct-module-impact:historical-roots" ||
		value.SourceSHA256 != result.Inputs[directModuleSourcePath] {
		return historicalRoots{}, fmt.Errorf("Gooo historical roots cannot be replayed: %v", err)
	}
	var roots historicalRoots
	if err := json.Unmarshal(value.Value, &roots); err != nil {
		return historicalRoots{}, err
	}
	if roots.Releases == roots.Observations || !validArchiveRoot(roots.Releases) ||
		!validArchiveRoot(roots.Observations) {
		return historicalRoots{}, fmt.Errorf("Gooo historical roots are not safe paths")
	}
	return roots, nil
}

func validArchiveRoot(root string) bool {
	return strings.HasPrefix(root, "docs/") && strings.HasSuffix(root, "/") &&
		path.Clean(root) == strings.TrimSuffix(root, "/") &&
		!strings.Contains(root, "\\") && !strings.Contains(root, "../")
}
