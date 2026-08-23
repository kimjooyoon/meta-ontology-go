package toolchainrelease

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/toolchain"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type BuildInput struct {
	Root         string
	OutputDir    string
	ExpectedHead string
	Target       Target
}

func validateBuildInput(input BuildInput) error {
	expected, ok := expectedTarget(input.Target.ID)
	if !ok || expected != input.Target {
		return fmt.Errorf("TOOLCHAIN_RELEASE_TARGET_UNKNOWN")
	}
	if !headPattern.MatchString(input.ExpectedHead) {
		return fmt.Errorf("TOOLCHAIN_RELEASE_HEAD_INVALID")
	}
	if runtime.GOOS != expected.GOOS || runtime.GOARCH != expected.GOARCH {
		return fmt.Errorf("TOOLCHAIN_RELEASE_NATIVE_PLATFORM_MISMATCH")
	}
	root, _ := filepath.Abs(input.Root)
	output, _ := filepath.Abs(input.OutputDir)
	relative, err := filepath.Rel(root, output)
	if err != nil || relative == "." || (!strings.HasPrefix(relative, ".."+string(filepath.Separator)) && relative != "..") {
		return fmt.Errorf("TOOLCHAIN_RELEASE_OUTPUT_INSIDE_REPOSITORY")
	}
	requirement, err := toolchain.ReadRequirement(input.Root)
	if err != nil {
		return err
	}
	if requirement.Toolchain != ExpectedToolchain || runtime.Version() != ExpectedToolchain {
		return fmt.Errorf("TOOLCHAIN_RELEASE_GO127_MISMATCH")
	}
	return nil
}
