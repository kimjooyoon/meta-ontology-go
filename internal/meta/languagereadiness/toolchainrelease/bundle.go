package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteBundle(directory string, evidence []PlatformEvidence) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	items := expectedEvidence(evidence)
	lines := make([]string, 0, len(items))
	for _, item := range items {
		name := item.Receipt.Archive.Name
		if err := copyFile(item.ArchivePath, filepath.Join(directory, name)); err != nil {
			return err
		}
		lines = append(lines, strings.TrimPrefix(item.Receipt.Archive.Digest, "sha256:")+"  "+name)
	}
	sort.Strings(lines)
	if len(lines) != TargetCount {
		return fmt.Errorf("TOOLCHAIN_RELEASE_BUNDLE_INCOMPLETE")
	}
	return os.WriteFile(filepath.Join(directory, "SHA256SUMS"),
		[]byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func expectedEvidence(evidence []PlatformEvidence) []PlatformEvidence {
	result := []PlatformEvidence{}
	for _, target := range targetRegistry {
		for _, item := range evidence {
			if item.Receipt.Platform.ID == target.ID {
				result = append(result, item)
				break
			}
		}
	}
	return result
}
