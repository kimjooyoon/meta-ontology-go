package main

import (
	"fmt"
	"runtime"
	"strings"
)

func computeDigests(root, generated string) (digests, error) {
	source, err := hashGitFiles(root)
	if err != nil {
		return digests{}, err
	}
	ir, err := hashGitFiles(root, "internal/semantic", "internal/bidir")
	if err != nil {
		return digests{}, err
	}
	fixtures, err := hashGitFiles(root, "internal/generator", "examples")
	if err != nil {
		return digests{}, err
	}
	policy, err := hashGitFiles(root, ".github", "scripts", "internal/verify")
	if err != nil {
		return digests{}, err
	}
	sourceMap, err := hashGeneratedSourceMap(generated)
	if err != nil {
		return digests{}, err
	}
	generatedDigest, err := hashDirectory(generated)
	if err != nil {
		return digests{}, err
	}
	toolchain, err := toolchainIdentity()
	if err != nil {
		return digests{}, err
	}
	return digests{SourceSHA256: source, IRSHA256: ir, GeneratorFixtureSHA256: fixtures, GeneratedOutputSHA256: generatedDigest, SourceMapSHA256: sourceMap, PolicySHA256: policy, ToolchainSHA256: digestBytes([]byte(toolchain))}, nil
}
func toolchainIdentity() (string, error) {
	goVersion, err := command("go", "version")
	if err != nil {
		return "", err
	}
	goEnv, err := command("go", "env", "GOVERSION", "GOROOT", "GOOS", "GOARCH")
	if err != nil {
		return "", err
	}
	runtimeVersion := runtime.Version()
	if runtimeVersion != "go1.26.5" || !strings.Contains(goVersion, "go1.26.5") || !strings.HasPrefix(goEnv, "go1.26.5\n") {
		return "", fmt.Errorf("toolchain identity is not independently go1.26.5: runtime=%q go=%q env=%q", runtimeVersion, strings.TrimSpace(goVersion), strings.TrimSpace(goEnv))
	}
	return runtimeVersion + "\n" + strings.TrimSpace(goVersion) + "\n" + strings.TrimSpace(goEnv), nil
}
