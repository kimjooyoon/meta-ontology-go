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
	StartByte  int    `json:"start_byte,omitempty"`
	EndByte    int    `json:"end_byte,omitempty"`
}

type projectionEdit struct {
	Path           string `json:"path"`
	Line           int    `json:"line"`
	StartByte      int    `json:"start_byte"`
	EndByte        int    `json:"end_byte"`
	Before         string `json:"before"`
	After          string `json:"after"`
	PreimageSHA256 string `json:"preimage_sha256"`
	RuleID         string `json:"rule_id"`
	RuleVersion    int    `json:"rule_version"`
}

type projectionPlan struct {
	Schema                string           `json:"schema"`
	ObservationSHA256     string           `json:"observation_sha256"`
	PlanCoreSHA256        string           `json:"plan_core_sha256"`
	CurrentVersion        string           `json:"current_version"`
	NextVersion           string           `json:"next_version"`
	ModuleSHA256          string           `json:"module_sha256"`
	ActiveOccurrences     int              `json:"active_occurrences"`
	ClassifiedOccurrences int              `json:"classified_occurrences"`
	UnresolvedOccurrences int              `json:"unresolved_occurrences"`
	StaleSelectors        int              `json:"stale_selectors"`
	HistoricalOccurrences int              `json:"historical_occurrences"`
	OutsideScope          int              `json:"outside_scope_occurrences"`
	OverlappingEdits      int              `json:"overlapping_edits"`
	HistoricalWrites      int              `json:"historical_writes"`
	OutsideScopeWrites    int              `json:"outside_scope_writes"`
	CandidateEdits        []projectionEdit `json:"candidate_edits"`
	Unresolved            []occurrence     `json:"unresolved"`
	Provenance            map[string]any   `json:"provenance,omitempty"`
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
	planVersion := flag.String("plan-version", "", "propose an exact next Go patch version without writing files")
	flag.Parse()
	if flag.NArg() != 0 {
		fail("unexpected positional argument")
	}
	result, err := scan(*root)
	if err != nil {
		fail(err.Error())
	}
	var value any = result
	if *planVersion != "" {
		plan, err := buildProjectionPlan(result, *planVersion)
		if err != nil {
			fail(err.Error())
		}
		value = plan
		if *inputs {
			value = projectionInputs(plan)
		}
	} else if *inputs {
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
	paths, err := trackedPaths(root)
	if err != nil {
		return report{}, err
	}
	return scanPaths(root, paths)
}

func scanPaths(root string, paths []string) (report, error) {
	requirement, err := toolchain.ReadRequirement(root)
	if err != nil {
		return report{}, err
	}
	version := strings.TrimPrefix(requirement.Language, "go")
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(version) {
		return report{}, fmt.Errorf("unsupported Go version %q", version)
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
		lineStart := 0
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
				lineStart += len(line) + 1
				continue
			}
			if !activePath(path) {
				result.OutsideScope += len(matches)
				lineStart += len(line) + 1
				continue
			}
			role, classified := classify(path, line, version, len(matches))
			for _, match := range matches {
				start := lineStart + match[0] + strings.Index(line[match[0]:match[1]], version)
				result.ActiveOccurrences++
				if classified {
					result.ClassifiedOccurrences++
				} else {
					result.UnresolvedOccurrences++
				}
				result.Occurrences = append(result.Occurrences, occurrence{
					Path: path, Line: lineNumber + 1, Role: role, Classified: classified, SHA256: fileHash,
					StartByte: start, EndByte: start + len(version),
				})
			}
			lineStart += len(line) + 1
		}
	}
	if result.ModuleSHA256 == "" || result.ActiveOccurrences == 0 {
		return report{}, fmt.Errorf("module or active version reference missing")
	}
	return result, nil
}

func buildProjectionPlan(observed report, next string) (projectionPlan, error) {
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(next) {
		return projectionPlan{}, fmt.Errorf("unsupported proposed Go version %q", next)
	}
	plan := projectionPlan{
		Schema: "gooo/toolchain-version-projection-plan/v1", CurrentVersion: observed.Version,
		NextVersion: next, ModuleSHA256: observed.ModuleSHA256,
		ActiveOccurrences: observed.ActiveOccurrences, ClassifiedOccurrences: observed.ClassifiedOccurrences,
		UnresolvedOccurrences: observed.UnresolvedOccurrences, StaleSelectors: observed.StaleSelectors,
		HistoricalOccurrences: observed.HistoricalOccurrences, OutsideScope: observed.OutsideScope,
		CandidateEdits: []projectionEdit{}, Unresolved: []occurrence{},
	}
	observationJSON, err := json.Marshal(observed)
	if err != nil {
		return projectionPlan{}, fmt.Errorf("encode observed impact: %w", err)
	}
	plan.ObservationSHA256 = digest(observationJSON)
	for _, site := range observed.Occurrences {
		if !site.Classified {
			plan.Unresolved = append(plan.Unresolved, site)
			continue
		}
		if !activePath(site.Path) || site.EndByte-site.StartByte != len(observed.Version) {
			return projectionPlan{}, fmt.Errorf("invalid classified site %s:%d", site.Path, site.Line)
		}
		plan.CandidateEdits = append(plan.CandidateEdits, projectionEdit{
			Path: site.Path, Line: site.Line, StartByte: site.StartByte, EndByte: site.EndByte,
			Before: observed.Version, After: next, PreimageSHA256: site.SHA256,
			RuleID: "urn:gooo:toolchain-projection:rule:" + site.Role, RuleVersion: 1,
		})
	}
	if len(plan.Unresolved) != observed.UnresolvedOccurrences ||
		len(plan.CandidateEdits) != observed.ClassifiedOccurrences {
		return projectionPlan{}, fmt.Errorf("projection observations disagree with counted sites")
	}
	sort.Slice(plan.CandidateEdits, func(i, j int) bool {
		left, right := plan.CandidateEdits[i], plan.CandidateEdits[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.StartByte < right.StartByte
	})
	sort.Slice(plan.Unresolved, func(i, j int) bool {
		left, right := plan.Unresolved[i], plan.Unresolved[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.StartByte != right.StartByte {
			return left.StartByte < right.StartByte
		}
		return left.Role < right.Role
	})
	for i := 1; i < len(plan.CandidateEdits); i++ {
		prior, current := plan.CandidateEdits[i-1], plan.CandidateEdits[i]
		if prior.Path == current.Path && prior.EndByte > current.StartByte {
			plan.OverlappingEdits++
		}
	}
	coreJSON, err := json.Marshal(plan)
	if err != nil {
		return projectionPlan{}, fmt.Errorf("encode projection plan: %w", err)
	}
	plan.PlanCoreSHA256 = digest(coreJSON)
	observationID := "urn:gooo:toolchain-projection:observation:" + plan.ObservationSHA256
	activityID := "urn:gooo:toolchain-projection:activity:" + plan.PlanCoreSHA256
	planID := "urn:gooo:toolchain-projection:plan:" + plan.PlanCoreSHA256
	plan.Provenance = map[string]any{
		"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#", "gooo": "urn:gooo:toolchain-projection:vocab:"},
		"@graph": []any{
			map[string]any{"@id": observationID, "@type": "prov:Entity", "gooo:content_sha256": plan.ObservationSHA256},
			map[string]any{"@id": activityID, "@type": "prov:Activity", "prov:used": map[string]string{"@id": observationID}},
			map[string]any{"@id": planID, "@type": "prov:Entity", "prov:wasGeneratedBy": map[string]string{"@id": activityID},
				"prov:wasDerivedFrom": map[string]string{"@id": observationID}, "gooo:core_sha256": plan.PlanCoreSHA256},
		},
	}
	return plan, nil
}

func projectionInputs(plan projectionPlan) map[string]any {
	return map[string]any{
		"schema": "gooo/body-composition-inputs/v1",
		"inputs": []any{map[string]any{
			"examples/toolchainimpact:AssessProjection": map[string]any{
				"module_sha256": plan.ModuleSHA256, "plan_core_sha256": plan.PlanCoreSHA256,
				"current_version": plan.CurrentVersion,
				"next_version":    plan.NextVersion, "active_occurrences": plan.ActiveOccurrences,
				"classified_occurrences": plan.ClassifiedOccurrences,
				"unresolved_occurrences": plan.UnresolvedOccurrences,
				"stale_selectors":        plan.StaleSelectors, "candidate_edits": len(plan.CandidateEdits),
				"overlapping_edits": plan.OverlappingEdits,
				"historical_writes": plan.HistoricalWrites, "outside_scope_writes": plan.OutsideScopeWrites,
			},
		}},
	}
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
