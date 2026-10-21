// Command toolchain-version-impact inventories current Go version references.
// It observes files and never edits them. The Gooo source makes the decision.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/toolchain"
)

type occurrence struct {
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Role       string `json:"role"`
	Classified bool   `json:"classified"`
	SHA256     string `json:"file_sha256"`
}

type report struct {
	Schema                string       `json:"schema"`
	Version               string       `json:"version"`
	ModuleSHA256          string       `json:"module_sha256"`
	ActiveOccurrences     int          `json:"active_occurrences"`
	ClassifiedOccurrences int          `json:"classified_occurrences"`
	UnresolvedOccurrences int          `json:"unresolved_occurrences"`
	StaleSelectors        int          `json:"stale_selectors"`
	HistoricalOccurrences int          `json:"historical_occurrences"`
	OutsideScope          int          `json:"outside_scope_occurrences"`
	Occurrences           []occurrence `json:"occurrences"`
}

func main() {
	root := flag.String("root", ".", "repository root")
	inputs := flag.Bool("gooo-inputs", false, "emit inputs for AssessImpact")
	flag.Parse()
	if flag.NArg() != 0 {
		fail("unexpected positional argument")
	}
	result, err := scan(*root)
	if err != nil {
		fail(err.Error())
	}
	var value any = result
	if *inputs {
		value = map[string]any{
			"schema": "gooo/body-composition-inputs/v1",
			"inputs": []any{map[string]any{
				"examples/toolchainimpact:AssessImpact": map[string]any{
					"module_sha256":          result.ModuleSHA256,
					"version":                result.Version,
					"active_occurrences":     result.ActiveOccurrences,
					"classified_occurrences": result.ClassifiedOccurrences,
					"unresolved_occurrences": result.UnresolvedOccurrences,
					"stale_selectors":        result.StaleSelectors,
					"historical_occurrences": result.HistoricalOccurrences,
				},
			}},
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func scan(root string) (report, error) {
	requirement, err := toolchain.ReadRequirement(root)
	if err != nil {
		return report{}, err
	}
	version := strings.TrimPrefix(requirement.Language, "go")
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(version) {
		return report{}, fmt.Errorf("unsupported Go version %q", version)
	}
	paths, err := trackedPaths(root)
	if err != nil {
		return report{}, err
	}
	current := regexp.MustCompile(`(^|[^0-9.])` + regexp.QuoteMeta(version) + `([^0-9.]|$)`)
	historicalGo := regexp.MustCompile(`\bGo\s+([0-9]+\.[0-9]+\.[0-9]+)|\bgo([0-9]+\.[0-9]+\.[0-9]+)`)
	result := report{Schema: "gooo/toolchain-version-impact/v1", Version: version, Occurrences: []occurrence{}}
	for _, path := range paths {
		full := filepath.Join(root, filepath.FromSlash(path))
		info, err := os.Lstat(full)
		if err != nil {
			return report{}, fmt.Errorf("stat tracked path %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			if activePath(path) {
				return report{}, fmt.Errorf("active tracked path %s is not a regular file", path)
			}
			continue
		}
		data, err := os.ReadFile(full)
		if err != nil {
			return report{}, err
		}
		if path == "go.mod" {
			result.ModuleSHA256 = digest(data)
		}
		if !strings.Contains(string(data), version) && !strings.HasPrefix(path, "docs/releases/") && !activePath(path) {
			continue
		}
		fileHash := digest(data)
		for lineNumber, line := range strings.Split(string(data), "\n") {
			if activePath(path) {
				if selected, ok := selectorVersion(path, line); ok && selected != version {
					result.StaleSelectors++
					result.ActiveOccurrences++
					result.UnresolvedOccurrences++
					result.Occurrences = append(result.Occurrences, occurrence{
						Path: path, Line: lineNumber + 1, Role: "stale_selector", SHA256: fileHash,
					})
				}
			}
			matches := current.FindAllStringIndex(line, -1)
			if strings.HasPrefix(path, "docs/releases/") {
				for _, match := range historicalGo.FindAllStringSubmatch(line, -1) {
					observed := match[1]
					if observed == "" {
						observed = match[2]
					}
					if observed != version {
						result.HistoricalOccurrences++
					}
				}
			}
			if len(matches) == 0 {
				continue
			}
			if !activePath(path) {
				result.OutsideScope += len(matches)
				continue
			}
			role, classified := classify(path, line, version, len(matches))
			for range matches {
				result.ActiveOccurrences++
				if classified {
					result.ClassifiedOccurrences++
				} else {
					result.UnresolvedOccurrences++
				}
				result.Occurrences = append(result.Occurrences, occurrence{
					Path: path, Line: lineNumber + 1, Role: role, Classified: classified, SHA256: fileHash,
				})
			}
		}
	}
	if result.ModuleSHA256 == "" || result.ActiveOccurrences == 0 {
		return report{}, fmt.Errorf("module or active version reference missing")
	}
	return result, nil
}

func trackedPaths(root string) ([]string, error) {
	command := exec.Command("git", "ls-files", "-z")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked files: %w", err)
	}
	parts := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	sort.Strings(parts)
	return parts, nil
}

func activePath(path string) bool {
	return path == "go.mod" || strings.HasPrefix(path, ".github/workflows/") &&
		(strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"))
}

func classify(path, line, version string, matches int) (string, bool) {
	if matches != 1 {
		return "multiple_values", false
	}
	if selected, ok := selectorVersion(path, line); ok && selected == version {
		if path == "go.mod" {
			return "module_directive", true
		}
		if strings.HasPrefix(strings.TrimPrefix(strings.TrimSpace(line), "run: "), `test "$(go env GOVERSION)"`) {
			return "workflow_toolchain_assertion", true
		}
		return "workflow_selector", true
	}
	if path == "go.mod" {
		return "module_other", false
	}
	return "workflow_unclassified", false
}

func selectorVersion(path, line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "run: ")
	var raw string
	switch {
	case path == "go.mod" && strings.HasPrefix(trimmed, "go "):
		raw = strings.TrimSpace(strings.TrimPrefix(trimmed, "go "))
	case path == "go.mod" && strings.HasPrefix(trimmed, "toolchain go"):
		raw = strings.TrimPrefix(trimmed, "toolchain go")
	case strings.HasPrefix(trimmed, "go-version:"):
		raw = scalar(strings.TrimPrefix(trimmed, "go-version:"))
	case strings.HasPrefix(trimmed, "GOTOOLCHAIN:"):
		raw = strings.TrimPrefix(scalar(strings.TrimPrefix(trimmed, "GOTOOLCHAIN:")), "go")
	case strings.HasPrefix(trimmed, `test "$(go env GOVERSION)" = "go`) && strings.HasSuffix(trimmed, `"`):
		raw = strings.TrimSuffix(strings.TrimPrefix(trimmed, `test "$(go env GOVERSION)" = "go`), `"`)
	default:
		return "", false
	}
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(raw) {
		return "", false
	}
	return raw, true
}

func scalar(raw string) string {
	value := strings.TrimSpace(strings.SplitN(raw, "#", 2)[0])
	return strings.Trim(value, "\"'")
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
