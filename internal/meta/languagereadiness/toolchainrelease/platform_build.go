package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

func BuildPlatform(input BuildInput) (PlatformReceipt, error) {
	if err := validateBuildInput(input); err != nil {
		return PlatformReceipt{}, err
	}
	if err := requireRepositoryHead(input.Root, input.ExpectedHead); err != nil {
		return PlatformReceipt{}, err
	}
	payload, err := buildReleasePayload(input.Root, input.Target, input.ExpectedHead)
	if err != nil {
		return PlatformReceipt{}, err
	}
	before, err := repositoryState(input.Root)
	if err != nil || before != "" {
		return PlatformReceipt{}, fmt.Errorf("TOOLCHAIN_RELEASE_REPOSITORY_NOT_CLEAN")
	}
	work, err := os.MkdirTemp("", "gooo-platform-release-")
	if err != nil {
		return PlatformReceipt{}, err
	}
	defer os.RemoveAll(work)
	firstBinary := filepath.Join(work, "a-"+binaryName(input.Target))
	secondBinary := filepath.Join(work, "b-"+binaryName(input.Target))
	if err := buildBinary(input.Root, firstBinary); err != nil {
		return PlatformReceipt{}, err
	}
	if err := buildBinary(input.Root, secondBinary); err != nil {
		return PlatformReceipt{}, err
	}
	binaryDigest, binaryBytes, err := replayDigest(firstBinary, secondBinary)
	if err != nil {
		return PlatformReceipt{}, err
	}
	build, err := inspectBuild(firstBinary, input)
	if err != nil {
		return PlatformReceipt{}, err
	}
	if err := verifyBinaryModules(firstBinary, payload.policy.Members); err != nil {
		return PlatformReceipt{}, err
	}
	smoke, err := smokeBinary(firstBinary, input.Root)
	if err != nil {
		return PlatformReceipt{}, err
	}
	if err := smokeLanguageExamples(firstBinary, work, input); err != nil {
		return PlatformReceipt{}, err
	}
	archive, memberEvidence, err := buildArchives(work, input, firstBinary, secondBinary, payload)
	if err != nil {
		return PlatformReceipt{}, err
	}
	after, err := repositoryState(input.Root)
	if err != nil || after != before {
		return PlatformReceipt{}, fmt.Errorf("TOOLCHAIN_RELEASE_REPOSITORY_WRITE")
	}
	receipt := newPlatformReceipt(input, build, smoke, archive, memberEvidence, binaryDigest, binaryBytes)
	return FinalizeReceipt(receipt)
}
