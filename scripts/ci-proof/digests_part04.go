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
)

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
