package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	metatoolchain "github.com/kimjooyoon/meta-ontology-go/internal/meta/toolchain"
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
		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(line, "gooo:generated") || strings.Contains(line, "gooo:slot") {
				hash.Write([]byte(name + "\x00" + line + "\x00"))
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func toolchainIdentity() (string, error) {
	return metatoolchain.Identity(".")
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
