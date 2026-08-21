package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func stageCutover(settings cutoverConfig, oldPaths, newPaths []string) (int, error) {
	set := map[string]bool{}
	for _, paths := range [][]string{oldPaths, newPaths} {
		for _, name := range paths {
			set[name] = true
		}
	}
	paths := make([]string, 0, len(set))
	for name := range set {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	input := append([]byte(strings.Join(paths, "\x00")), 0)
	if _, err := cutoverGit(settings.root, input, "add", "--pathspec-from-file=-", "--pathspec-file-nul"); err != nil {
		return 1, err
	}
	unstaged, err := cutoverGit(settings.root, nil, "diff", "--name-only")
	if err != nil || len(bytes.TrimSpace(unstaged)) != 0 {
		return 1, fmt.Errorf("cutover left unstaged paths: %s: %w", unstaged, err)
	}
	untracked, err := cutoverGit(settings.root, nil, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil || len(untracked) != 0 {
		return 1, fmt.Errorf("cutover left untracked paths: %w", err)
	}
	tree, err := cutoverGit(settings.root, nil, "write-tree")
	if err != nil {
		return 1, err
	}
	staged, err := cutoverGit(settings.root, nil, "ls-tree", "-r", "--name-only", "-z", strings.TrimSpace(string(tree)))
	if err != nil {
		return 1, err
	}
	return mismatch(newPaths, zeroStrings(staged)), nil
}
