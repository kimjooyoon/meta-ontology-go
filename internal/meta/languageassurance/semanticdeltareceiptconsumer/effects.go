package semanticdeltareceiptconsumer

import (
	"os"
	"path/filepath"
	"strings"
)

func observeEffects(input Input) EffectsObservation {
	result := EffectsObservation{Status: effectsUnknown, TransientWriteObservation: effectsUnknown, MutationAuthority: effectsUnknown, OutputLocation: outputLocation(input.OutputPath)}
	if input.EffectsBeforePath == "" || input.EffectsAfterPath == "" {
		return result
	}
	beforeRaw, beforeErr := os.ReadFile(input.EffectsBeforePath)
	afterRaw, afterErr := os.ReadFile(input.EffectsAfterPath)
	if beforeErr != nil || afterErr != nil {
		return result
	}
	before, beforeOK := snapshotEntries(beforeRaw)
	after, afterOK := snapshotEntries(afterRaw)
	if !beforeOK || !afterOK {
		return result
	}
	changed := 0
	for path, digest := range before {
		if after[path] != digest {
			changed++
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			changed++
		}
	}
	result.Status = effectsNetStateUnchanged
	if changed != 0 {
		result.Status = "NET_REPOSITORY_STATE_CHANGED"
	}
	result.BeforeSnapshotPath, result.AfterSnapshotPath = input.EffectsBeforePath, input.EffectsAfterPath
	result.BeforeSnapshotDigest, result.AfterSnapshotDigest = digestBytes(beforeRaw), digestBytes(afterRaw)
	result.BeforePathCount, result.AfterPathCount, result.ChangedPathOrContentCount = len(before), len(after), changed
	return result
}

func snapshotEntries(raw []byte) (map[string]string, bool) {
	entries := make(map[string]string)
	if len(raw) == 0 {
		return entries, true
	}
	for line := range strings.SplitSeq(strings.TrimSuffix(string(raw), "\n"), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, false
		}
		if _, duplicate := entries[parts[0]]; duplicate {
			return nil, false
		}
		entries[parts[0]] = parts[1]
	}
	return entries, true
}

func outputLocation(path string) string {
	if path == "" {
		return effectsUnknown
	}
	output, err := filepath.Abs(path)
	if err != nil {
		return effectsUnknown
	}
	root, err := filepath.Abs(".")
	if err != nil {
		return effectsUnknown
	}
	relative, err := filepath.Rel(root, output)
	if err != nil {
		return effectsUnknown
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "OUTSIDE_REPOSITORY"
	}
	return "INSIDE_REPOSITORY"
}
