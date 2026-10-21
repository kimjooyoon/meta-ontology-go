package toolchainrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type preservedVersionSpanRule struct {
	Path    string `json:"path"`
	Prefix  string `json:"prefix"`
	Version string `json:"version"`
	ID      string `json:"rule_id"`
}

type preservedVersionSpanColumns struct {
	Paths    []string `json:"paths"`
	Prefixes []string `json:"prefixes"`
	Versions []string `json:"versions"`
	IDs      []string `json:"rule_ids"`
}

func readPreservedVersionSpans(ctx context.Context, root string, result *DirectModuleImpact) error {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil || payloadDigest(source) != result.Inputs[directModuleSourcePath] {
		return fmt.Errorf("Gooo preserved span source preimage changed")
	}
	input, _ := json.Marshal([]any{map[string]any{"module_path": result.Module}})
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"RequiredPreservedVersionSpans", input)
	if err != nil || value.ActivityID != "urn:gooo:direct-module-impact:required-preserved-spans" ||
		value.SourceSHA256 != result.Inputs[directModuleSourcePath] {
		return fmt.Errorf("Gooo preserved span rules cannot be replayed: %v", err)
	}
	var columns preservedVersionSpanColumns
	if err := json.Unmarshal(value.Value, &columns); err != nil {
		return err
	}
	if len(columns.Paths) != len(columns.Prefixes) || len(columns.Paths) != len(columns.Versions) ||
		len(columns.Paths) != len(columns.IDs) || len(columns.Paths) > 16 {
		return fmt.Errorf("Gooo preserved span rule columns are inconsistent")
	}
	activePaths, seenIDs := map[string]bool{}, map[string]bool{}
	for _, rule := range result.ActiveRules {
		activePaths[rule.Path] = true
		seenIDs[rule.ID] = true
	}
	for i, name := range columns.Paths {
		prefix, version, id := columns.Prefixes[i], columns.Versions[i], columns.IDs[i]
		if !activePaths[name] || prefix == "" || len(prefix) > 256 ||
			version == "" || len(version) > 128 || strings.ContainsAny(version, " \t\r\n") ||
			!strings.HasPrefix(id, "urn:gooo:direct-module-impact:rule:") ||
			strings.ContainsAny(id, " \t\r\n") || seenIDs[id] {
			return fmt.Errorf("Gooo preserved span has an invalid or repeated source identity")
		}
		seenIDs[id] = true
		result.PreservedSpans = append(result.PreservedSpans,
			preservedVersionSpanRule{Path: name, Prefix: prefix, Version: version, ID: id})
	}
	return nil
}

func observePlanVersionMentions(root string, result *DirectModuleImpact) error {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil || payloadDigest(source) != result.Inputs[directModuleSourcePath] {
		return fmt.Errorf("Gooo version role source preimage changed")
	}
	declared := 0
	for _, rule := range result.PreservedSpans {
		if rule.Version == result.CurrentVersion {
			declared++
		}
	}
	if bytes.Count(source, []byte(result.CurrentVersion)) != declared {
		result.Unresolved = append(result.Unresolved,
			"Gooo version role source contains an unowned current-version mention")
	}
	return nil
}

// An output document is closed only when every current-version occurrence has
// one source-owned role. A preserved span remains fixed across future upgrades.
func observeDocumentSpans(root string, result *DirectModuleImpact) error {
	byPath := map[string][]activeVersionRule{}
	for _, active := range result.ActiveRules {
		byPath[active.Path] = append(byPath[active.Path], active)
	}
	for _, name := range result.Outputs {
		activeRules := byPath[name]
		if len(activeRules) == 0 {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil || payloadDigest(raw) != result.Inputs[name] {
			return fmt.Errorf("active document preimage changed: %s", name)
		}
		owned := map[int]bool{}
		claim := func(prefix, version, id string) {
			span := []byte(prefix + version)
			if bytes.Count(raw, span) != 1 {
				result.Unresolved = append(result.Unresolved, "version span is not unique: "+id)
				return
			}
			if version != result.CurrentVersion {
				return
			}
			offset := bytes.Index(raw, span) + len(prefix)
			if owned[offset] {
				result.Unresolved = append(result.Unresolved, "version span has duplicate roles: "+id)
			}
			owned[offset] = true
		}
		for _, active := range activeRules {
			claim(active.Prefix, result.CurrentVersion, active.ID)
		}
		for _, preserved := range result.PreservedSpans {
			if preserved.Path == name {
				claim(preserved.Prefix, preserved.Version, preserved.ID)
			}
		}
		version := []byte(result.CurrentVersion)
		for offset := 0; offset < len(raw); {
			index := bytes.Index(raw[offset:], version)
			if index < 0 {
				break
			}
			offset += index
			if !owned[offset] {
				result.Unresolved = append(result.Unresolved,
					fmt.Sprintf("current version span has no Gooo role: %s:%d", name, offset))
			}
			offset += len(version)
		}
	}
	return nil
}
