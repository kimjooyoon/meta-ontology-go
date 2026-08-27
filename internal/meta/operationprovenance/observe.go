package operationprovenance

import (
	"os/exec"
	"sort"
	"strings"
)

type repositorySnapshot struct {
	digest string
	status map[string]string
}

func readRepositorySnapshot(root string) (repositorySnapshot, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return repositorySnapshot{}, err
	}
	status := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		path := strings.TrimSpace(line)
		if len(path) > 3 {
			path = strings.TrimSpace(path[3:])
		}
		status[path] = line
	}
	return repositorySnapshot{digest: digestBytes(output), status: status}, nil
}

func deriveObservation(before, after repositorySnapshot) WorkspaceObservation {
	paths := make([]string, 0)
	seen := make(map[string]bool)
	for path := range before.status {
		seen[path] = true
	}
	for path := range after.status {
		seen[path] = true
	}
	for path := range seen {
		if before.status[path] != after.status[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	writes := len(paths) > 0
	return WorkspaceObservation{
		BeforeDigest: before.digest, AfterDigest: after.digest,
		ChangedPaths: paths, RepositoryWorkspaceWrites: writes,
		MutationAuthority: writes,
	}
}
