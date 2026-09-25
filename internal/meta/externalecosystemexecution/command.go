package externalecosystemexecution

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func controlledEnv() []string {
	blocked := []string{"GOTOOLCHAIN=", "GOWORK=", "GOFLAGS="}
	env := make([]string, 0, len(os.Environ())+3)
	for _, item := range os.Environ() {
		keep := true
		for _, prefix := range blocked {
			if strings.HasPrefix(item, prefix) {
				keep = false
			}
		}
		if keep {
			env = append(env, item)
		}
	}
	return append(env, "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=-mod=readonly")
}

func captureRepository(ctx context.Context, root string) (RepositoryState, error) {
	commit, err := commandText(ctx, root, "git", "rev-parse", "HEAD")
	if err != nil {
		return RepositoryState{}, err
	}
	tree, err := commandText(ctx, root, "git", "rev-parse", "HEAD^{tree}")
	if err != nil {
		return RepositoryState{}, err
	}
	state, err := commandText(ctx, root, "git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return RepositoryState{}, err
	}
	return RepositoryState{true, commit, tree, state != "", Digest(state)}, nil
}

func moduleGoVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "go" {
			return fields[1], nil
		}
	}
	return "", scanner.Err()
}

func loadReference(sourceRoot string, external RepositoryState, moduleGo string) ReferenceReceipt {
	path := filepath.Join(sourceRoot, referenceEvidencePath)
	b, err := os.ReadFile(path)
	available := err == nil
	exact := available && json.Valid(b) && bytes.Contains(b, []byte(ExpectedCommit)) &&
		bytes.Contains(b, []byte(ExpectedTree)) && bytes.Contains(b, []byte(ExpectedModuleGo))
	return ReferenceReceipt{
		Available: available, BindingExact: exact, ContractVersion: ReferenceContractVersion,
		Decision: ExpectedReferenceDecision, Resolution: "EXACT", URL: ExpectedReferenceURL,
		Commit: external.Commit, Tree: external.Tree, ModuleGo: moduleGo,
		EvidencePath: referenceEvidencePath, EvidenceSHA256: Digest(b),
	}
}
