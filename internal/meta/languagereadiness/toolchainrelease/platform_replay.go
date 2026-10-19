package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

func replayDigest(first, second string) (string, int64, error) {
	firstDigest, firstBytes, err := digestFile(first)
	if err != nil {
		return "", 0, err
	}
	secondDigest, secondBytes, err := digestFile(second)
	if err != nil {
		return "", 0, err
	}
	if firstDigest != secondDigest || firstBytes != secondBytes {
		return "", 0, fmt.Errorf("TOOLCHAIN_RELEASE_REPLAY_MISMATCH")
	}
	return firstDigest, firstBytes, nil
}

func buildArchives(work string, input BuildInput, firstBinary, secondBinary string,
	payload releasePayload) (ReplayEvidence, PayloadEvidence, error) {
	first := filepath.Join(work, "a-"+archiveName(input.Target))
	second := filepath.Join(work, "b-"+archiveName(input.Target))
	firstPayload, err := archiveBinary(first, firstBinary, input.Target, payload, input.ExpectedHead)
	if err != nil {
		return ReplayEvidence{}, PayloadEvidence{}, err
	}
	secondPayload, err := archiveBinary(second, secondBinary, input.Target, payload, input.ExpectedHead)
	if err != nil || !samePayloadEvidence(firstPayload, secondPayload) {
		return ReplayEvidence{}, PayloadEvidence{}, fmt.Errorf("TOOLCHAIN_RELEASE_PAYLOAD_REPLAY_MISMATCH: %v", err)
	}
	digest, size, err := replayDigest(first, second)
	if err != nil {
		return ReplayEvidence{}, PayloadEvidence{}, err
	}
	if err := os.MkdirAll(input.OutputDir, 0o755); err != nil {
		return ReplayEvidence{}, PayloadEvidence{}, err
	}
	if err := copyFile(first, filepath.Join(input.OutputDir, archiveName(input.Target))); err != nil {
		return ReplayEvidence{}, PayloadEvidence{}, err
	}
	return ReplayEvidence{Name: archiveName(input.Target), Digest: digest, Bytes: size, Builds: 2, ReplayEqual: true}, firstPayload, nil
}
