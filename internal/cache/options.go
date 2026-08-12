package cache

import (
	"fmt"
)

// DigestOptions establishes the typed boundary for options that affect a key.
// Callers must pass this digest to ProjectionKeySpec or KeySpec; opaque values
// embedded in either compatibility field are rejected by the constructors.
func DigestOptions(options any) (Digest, error) {
	if unknownFreshnessValue(options) {
		return "", fmt.Errorf("%w: options", ErrUnknownFreshness)
	}
	digest, err := DigestOf(options)
	if err != nil {
		return "", fmt.Errorf("hash options: %w", err)
	}
	if !digest.Known() {
		return "", fmt.Errorf("%w: options digest", ErrUnknownFreshness)
	}
	return digest, nil
}

func requireOptionsDigest(digest Digest, opaque any) (Digest, error) {
	if opaque != nil {
		return "", fmt.Errorf("%w: opaque Options requires OptionsDigest", ErrUnknownFreshness)
	}
	if !digest.Known() {
		return "", fmt.Errorf("%w: options digest", ErrUnknownFreshness)
	}
	return digest, nil
}
