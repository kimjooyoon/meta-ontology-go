package externalcapabilityexecution

import "strings"

func observeReference(root string) (Reference, error) {
	repository, err := runText(root, "git", "config", "--get", "remote.origin.url")
	if err != nil {
		return Reference{}, err
	}
	commit, err := runText(root, "git", "rev-parse", "HEAD")
	if err != nil {
		return Reference{}, err
	}
	tree, err := runText(root, "git", "rev-parse", "HEAD^{tree}")
	if err != nil {
		return Reference{}, err
	}
	version, err := runText("", "go", "env", "GOVERSION")
	if err != nil {
		return Reference{}, err
	}
	return Reference{
		RepositoryURL: strings.TrimSuffix(repository, ".git"),
		CommitSHA:     commit, TreeSHA: tree, GoVersion: version,
	}, nil
}

func dirtyPaths(root string) (int, error) {
	status, err := runText(root, "git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return 0, err
	}
	if status == "" {
		return 0, nil
	}
	return len(strings.Split(status, "\n")), nil
}
