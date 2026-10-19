package toolchainrelease

import (
	"debug/buildinfo"
	"fmt"
)

// The license policy must describe the modules linked into this executable,
// not merely the dependencies written in the release checkout's go.mod.
func verifyBinaryModules(path string, members requiredMembers) error {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return err
	}
	want := map[string]string{members.RuntimeModule: members.RuntimeVersion, members.JevModule: members.JevVersion}
	seen := map[string]bool{}
	for _, dependency := range info.Deps {
		version, required := want[dependency.Path]
		if !required {
			continue
		}
		if seen[dependency.Path] || dependency.Version != version || dependency.Replace != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_BINARY_MODULE_MISMATCH")
		}
		seen[dependency.Path] = true
	}
	if len(seen) != len(want) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_BINARY_MODULE_MISMATCH")
	}
	return nil
}
