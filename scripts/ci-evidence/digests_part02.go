package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
