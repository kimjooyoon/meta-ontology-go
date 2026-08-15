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

func computeProofDigests(root, generated string, evidence evidenceInput) (proofDigests, error) {
	source, err := hashGitFiles(root)
	if err != nil {
		return proofDigests{}, err
	}
	semantic, err := hashGitFiles(root, "internal/semantic", "internal/bidir")
	if err != nil {
		return proofDigests{}, err
	}
	provenance, err := hashGitFiles(root, "internal/provenance", "internal/semantic/prov.go")
	if err != nil {
		return proofDigests{}, err
	}
	projection, err := hashDirectory(generated)
	if err != nil {
		return proofDigests{}, err
	}
	build, err := hashGitFiles(root, "go.mod", "go.sum")
	if err != nil {
		return proofDigests{}, err
	}
	policy, err := hashGitFiles(root, ".github", "scripts", "internal/verify")
	if err != nil {
		return proofDigests{}, err
	}
	schema, err := hashGitFiles(root, ".github/ci-governance.json")
	if err != nil {
		return proofDigests{}, err
	}
	toolchain, err := toolchainIdentity()
	if err != nil {
		return proofDigests{}, err
	}
	toolchainDigest := digestBytes([]byte(toolchain))
	if mismatches := digestMismatchFields(source, semantic, projection, policy, toolchainDigest, evidence); len(mismatches) > 0 {
		return proofDigests{}, fmt.Errorf("independently recomputed CI evidence digest mismatch: %s", strings.Join(mismatches, ", "))
	}
	target := digestBytes([]byte(evidence.BaseRef + "\x00" + evidence.BaseSHA))
	return proofDigests{Source: source, Semantic: semantic, Provenance: provenance, Projection: projection, Build: build, Policy: policy, Schema: schema, Toolchain: toolchainDigest, Target: target}, nil
}

func digestMismatchFields(source, semantic, projection, policy, toolchain string, evidence evidenceInput) []string {
	checks := []struct {
		name, actual, expected string
	}{
		{"source", source, evidence.Digests.Source},
		{"semantic", semantic, evidence.Digests.IR},
		{"projection", projection, evidence.Digests.Generated},
		{"policy", policy, evidence.Digests.Policy},
		{"toolchain", toolchain, evidence.Digests.Toolchain},
	}
	mismatches := make([]string, 0, len(checks))
	for _, check := range checks {
		if check.actual != check.expected {
			mismatches = append(mismatches, check.name)
		}
	}
	return mismatches
}

func hashGitFiles(root string, prefixes ...string) (string, error) {
	args := []string{"ls-files", "-z"}
	if len(prefixes) > 0 {
		args = append(args, "--")
		args = append(args, prefixes...)
	}
	namesOutput, err := commandIn(root, "git", args...)
	if err != nil {
		return "", err
	}
	names := strings.Split(strings.TrimSuffix(namesOutput, "\x00"), "\x00")
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

func toolchainIdentity() (string, error) {
	goVersion, err := command("go", "version")
	if err != nil {
		return "", err
	}
	goEnv, err := command("go", "env", "GOVERSION", "GOROOT", "GOOS", "GOARCH")
	if err != nil {
		return "", err
	}
	if runtime.Version() != "go1.26.5" || !strings.Contains(goVersion, "go1.26.5") || !strings.HasPrefix(goEnv, "go1.26.5\n") {
		return "", fmt.Errorf("toolchain is not independently go1.26.5")
	}
	return runtime.Version() + "\n" + strings.TrimSpace(goVersion) + "\n" + strings.TrimSpace(goEnv), nil
}

func command(name string, args ...string) (string, error) {
	return commandIn(".", name, args...)
}

func commandIn(root, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(output), nil
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
