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

type historicalFileRule struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	ID     string `json:"id"`
}

type historicalFileColumns struct {
	Paths  []string `json:"paths"`
	SHA256 []string `json:"sha256"`
	IDs    []string `json:"rule_ids"`
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
	rules, err := readHistoricalFileRules(ctx, root, result)
	if err != nil {
		return err
	}
	result.HistoricalRules = rules
	byPath := make(map[string]historicalFileRule, len(rules))
	for _, rule := range rules {
		byPath[rule.Path] = rule
	}
	head, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil || len(strings.TrimSpace(string(head))) != 40 || !isHex(strings.TrimSpace(string(head))) {
		return fmt.Errorf("archive HEAD identity is unavailable: %v", err)
	}
	result.ArchiveHead = strings.TrimSpace(string(head))
	declared := make(map[string]bool, len(result.Outputs))
	for _, name := range result.Outputs {
		declared[name] = true
	}
	declared[directModuleSourcePath] = true
	seen := make(map[string]bool)
	for item := range bytes.SplitSeq(listing[:len(listing)-1], []byte{0}) {
		name := string(item)
		if name == "" || seen[name] || path.IsAbs(name) || path.Clean(name) != name ||
			name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return fmt.Errorf("invalid or repeated tracked path: %q", name)
		}
		seen[name] = true
		if err := observeTrackedReferenceFile(ctx, root, name, current, declared, roots, byPath, result); err != nil {
			return err
		}
	}
	for _, rule := range rules {
		if !seen[rule.Path] {
			result.Unresolved = append(result.Unresolved,
				"Gooo historical file is not tracked: "+rule.ID)
		}
	}
	sort.Strings(result.OutsideScope)
	sort.Strings(result.HistoricalScope)
	result.ScanElapsedUS = time.Since(started).Microseconds()
	return nil
}

func observeTrackedReferenceFile(ctx context.Context, root, name, current string,
	declared map[string]bool, roots historicalRoots, byPath map[string]historicalFileRule,
	result *DirectModuleImpact) error {
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
	exactHistorical := false
	if rule, declared := byPath[name]; declared {
		if payloadDigest(raw) != rule.SHA256 {
			result.Unresolved = append(result.Unresolved, "Gooo historical file digest changed: "+rule.ID)
		} else {
			committed, err := exec.CommandContext(ctx, "git", "-C", root, "show",
				result.ArchiveHead+":"+name).Output()
			if err != nil || !bytes.Equal(committed, raw) {
				result.Unresolved = append(result.Unresolved, "Gooo historical file differs from HEAD: "+rule.ID)
			} else {
				exactHistorical = true
			}
		}
	}
	if !bytes.Contains(raw, []byte(current)) {
		return nil
	}
	result.ReferenceFiles++
	if !declared[name] {
		result.Inputs[name] = payloadDigest(raw)
		if exactHistorical {
			result.HistoricalScope = append(result.HistoricalScope, name)
			return nil
		}
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

func readHistoricalFileRules(ctx context.Context, root string,
	result *DirectModuleImpact) ([]historicalFileRule, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil || payloadDigest(source) != result.Inputs[directModuleSourcePath] {
		return nil, fmt.Errorf("historical file rule source preimage changed")
	}
	input, _ := json.Marshal([]any{map[string]any{"module_path": result.Module}})
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"HistoricalReferenceFiles", input)
	if err != nil || value.ActivityID != "urn:gooo:direct-module-impact:historical-reference-files" ||
		value.SourceSHA256 != result.Inputs[directModuleSourcePath] {
		return nil, fmt.Errorf("Gooo historical file rules cannot be replayed: %v", err)
	}
	var columns historicalFileColumns
	if err := json.Unmarshal(value.Value, &columns); err != nil {
		return nil, err
	}
	if len(columns.Paths) != len(columns.SHA256) || len(columns.Paths) != len(columns.IDs) ||
		len(columns.Paths) > 32 {
		return nil, fmt.Errorf("Gooo historical file rule columns are inconsistent")
	}
	seenPaths, seenIDs := map[string]bool{}, map[string]bool{}
	for _, output := range result.Outputs {
		seenPaths[output] = true
	}
	rules := make([]historicalFileRule, 0, len(columns.Paths))
	for i, name := range columns.Paths {
		digest, id := columns.SHA256[i], columns.IDs[i]
		if (!strings.HasPrefix(name, "docs/") && !strings.HasPrefix(name, "examples/")) ||
			path.Clean(name) != name || strings.Contains(name, "\\") || seenPaths[name] ||
			len(digest) != len("sha256:")+64 || !strings.HasPrefix(digest, "sha256:") ||
			!isHex(strings.TrimPrefix(digest, "sha256:")) ||
			!strings.HasPrefix(id, "urn:gooo:direct-module-impact:historical:") ||
			strings.ContainsAny(id, " \t\r\n") || seenIDs[id] {
			return nil, fmt.Errorf("Gooo historical file rule has an invalid or repeated identity")
		}
		seenPaths[name], seenIDs[id] = true, true
		rules = append(rules, historicalFileRule{Path: name, SHA256: digest, ID: id})
	}
	return rules, nil
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
