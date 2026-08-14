package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/scripts/ci-proof/manifest"
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

func hashGitFiles(root string, prefixes ...string) (string, error) {
	args := []string{"ls-files", "-z"}
	if len(prefixes) > 0 {
		args = append(args, "--")
		args = append(args, prefixes...)
	}
	output, err := commandIn(root, "git", args...)
	if err != nil {
		return "", err
	}
	names := strings.Split(strings.TrimSuffix(output, "\x00"), "\x00")
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		if name == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hashDirectory(root string) (string, error) {
	names := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			names = append(names, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(name, ".manifest.jsonl") {
			data, err = manifest.Canonicalize(root, data)
			if err != nil {
				return "", err
			}
		}
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hashGeneratedSourceMap(root string) (string, error) {
	names := make([]string, 0)
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			names = append(names, filepath.ToSlash(relative))
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "gooo:generated") || strings.Contains(line, "gooo:slot") {
				hash.Write([]byte(name + "\x00" + line + "\x00"))
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func command(name string, args ...string) (string, error) {
	return commandIn(".", name, args...)
}

func commandIn(root, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	command.Dir = root
	data, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(data), nil
}
