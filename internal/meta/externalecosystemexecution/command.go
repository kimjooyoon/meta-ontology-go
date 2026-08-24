package externalecosystemexecution

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
)

func controlledEnv() []string {
	blocked := []string{"GOTOOLCHAIN=", "GOWORK=", "GOFLAGS="}
	env := make([]string, 0, len(os.Environ())+3)
	for _, item := range os.Environ() {
		keep := true
		for _, prefix := range blocked {
			if strings.HasPrefix(item, prefix) { keep = false }
		}
		if keep { env = append(env, item) }
	}
	return append(env, "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=-mod=readonly")
}

func commandText(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir, cmd.Env = dir, controlledEnv()
	b, err := cmd.Output()
	return strings.TrimSpace(string(b)), err
}

func captureRepository(ctx context.Context, root string) (RepositoryState, error) {
	commit, err := commandText(ctx, root, "git", "rev-parse", "HEAD")
	if err != nil { return RepositoryState{}, err }
	tree, err := commandText(ctx, root, "git", "rev-parse", "HEAD^{tree}")
	if err != nil { return RepositoryState{}, err }
	state, err := commandText(ctx, root, "git", "status", "--porcelain", "--untracked-files=all")
	if err != nil { return RepositoryState{}, err }
	return RepositoryState{true, commit, tree, state != "", Digest(state)}, nil
}

func moduleGoVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil { return "", err }
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "go" { return fields[1], nil }
	}
	return "", scanner.Err()
}
