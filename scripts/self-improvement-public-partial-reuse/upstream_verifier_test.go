package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestPinUpstreamVerifierDigestMismatchIsRefuted(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "input-verifier")
	if err := os.WriteFile(filename, []byte("untrusted verifier bytes"), 0o444); err != nil {
		t.Fatal(err)
	}
	input := runInput{OrchestrationVerifier: filename,
		OrchestrationVerifierDigest: cache.HashBytes([]byte("different trusted bytes")).String()}
	pinned, err := pinUpstreamVerifier(input, directory)
	var failure *upstreamFailure
	if !errors.As(err, &failure) || failure.Decision != "REFUTED" ||
		failure.Reason != "UPSTREAM_VERIFIER_DIGEST_MISMATCH" || failure.Unknown != nil || pinned != "" {
		t.Fatalf("digest contradiction was not retained: path=%q error=%v", pinned, err)
	}
	if _, err := os.Stat(filepath.Join(directory, "orchestration-verifier")); !os.IsNotExist(err) {
		t.Fatal("contradictory verifier bytes reached the pinned executable path")
	}
}

func TestPinUpstreamVerifierMissingMaterialIsUnknown(t *testing.T) {
	directory := t.TempDir()
	input := runInput{OrchestrationVerifier: filepath.Join(directory, "missing")}
	pinned, err := pinUpstreamVerifier(input, directory)
	var failure *upstreamFailure
	if !errors.As(err, &failure) || failure.Decision != "UNKNOWN" ||
		failure.Reason != "UPSTREAM_VERIFIER_NOT_ESTABLISHED" || failure.Unknown == nil || pinned != "" {
		t.Fatalf("missing material classification: path=%q error=%v", pinned, err)
	}
	unknown := failure.Unknown
	if unknown.Stage != "UPSTREAM_ORCHESTRATION" || unknown.Step != "PIN_VERIFIER" ||
		unknown.Reason != failure.Reason || unknown.UnknownClass != "DEPENDENCY_BLOCKED" ||
		unknown.NextOperation == "" || len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != "orchestration.verification" {
		t.Fatalf("missing material lost its causal frontier: %+v", unknown)
	}
}

func TestUpstreamVerifierBuildIdentityClassifications(t *testing.T) {
	for _, field := range []string{"matching", "package", "toolchain", "revision", "dirty"} {
		t.Run(field, func(t *testing.T) {
			info := debug.BuildInfo{
				Path: "github.com/kimjooyoon/meta-ontology-go/scripts/self-improvement-public-orchestration",
				GoVersion: "go1.27.0",
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "test-source-head"},
					{Key: "vcs.modified", Value: "false"},
				},
			}
			switch field {
			case "package":
				info.Path = "different/verifier"
			case "toolchain":
				info.GoVersion = "go1.26.0"
			case "revision":
				info.Settings[0].Value = "different-source-head"
			case "dirty":
				info.Settings[1].Value = "true"
			}
			err := validateUpstreamVerifierBuild(&info, "test-source-head")
			if field == "matching" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *upstreamFailure
			if !errors.As(err, &failure) || failure.Decision != "REFUTED" ||
				failure.Reason != "UPSTREAM_VERIFIER_BUILD_IDENTITY_MISMATCH" || failure.Unknown != nil {
				t.Fatalf("build identity contradiction was not retained: %v", err)
			}
		})
	}
}
