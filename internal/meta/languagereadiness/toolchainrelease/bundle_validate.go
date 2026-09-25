package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ValidateBundle(directory string, evidence []PlatformEvidence) error {
	items := expectedEvidence(evidence)
	lines := make([]string, 0, len(items))
	for _, item := range items {
		name := item.Receipt.Archive.Name
		digest, size, err := digestFile(filepath.Join(directory, name))
		if err != nil || digest != item.Receipt.Archive.Digest || size != item.Receipt.Archive.Bytes {
			return fmt.Errorf("TOOLCHAIN_RELEASE_BUNDLE_DIGEST_MISMATCH")
		}
		lines = append(lines, strings.TrimPrefix(digest, "sha256:")+"  "+name)
	}
	sort.Strings(lines)
	raw, err := os.ReadFile(filepath.Join(directory, "SHA256SUMS"))
	if err != nil || string(raw) != strings.Join(lines, "\n")+"\n" {
		return fmt.Errorf("TOOLCHAIN_RELEASE_CHECKSUM_MANIFEST_MISMATCH")
	}
	return nil
}
