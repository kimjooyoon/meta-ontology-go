package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

type entry struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

func main() {
	root := flag.String("root", "", "repository root")
	headSHA := flag.String("head-sha", "", "exact observed HEAD")
	executionID := flag.String("execution-id", "", "witness execution identifier")
	entriesPath := flag.String("entries", "", "raw entry artifact path")
	outputPath := flag.String("output", "", "snapshot metadata path")
	flag.Parse()
	if *root == "" || *headSHA == "" || *executionID == "" || *entriesPath == "" || *outputPath == "" {
		fail("-root, -head-sha, -execution-id, -entries, and -output are required")
	}
	if !model.ValidHead(*headSHA) || !model.ValidExecutionID(*headSHA, *executionID) {
		fail("invalid expected head sha or execution id")
	}
	rootAbsolute, err := filepath.Abs(*root)
	if err != nil {
		fail(fmt.Sprintf("resolve repository root: %v", err))
	}
	observed, err := gitHead(rootAbsolute)
	if err != nil {
		fail(err.Error())
	}
	if observed != *headSHA {
		fail(fmt.Sprintf("SNAPSHOT_BINDING/observe-head/HEAD_SHA_MISMATCH expected=%s actual=%s", *headSHA, observed))
	}
	entries, err := repositoryEntries(rootAbsolute)
	if err != nil {
		fail(err.Error())
	}
	raw, err := json.Marshal(entries)
	if err != nil {
		fail(fmt.Sprintf("encode raw repository entries: %v", err))
	}
	raw = append(raw, '\n')
	entriesAbsolute, err := filepath.Abs(*entriesPath)
	if err != nil {
		fail(fmt.Sprintf("resolve entries path: %v", err))
	}
	if err := os.MkdirAll(filepath.Dir(entriesAbsolute), 0o755); err != nil {
		fail(fmt.Sprintf("create entries directory: %v", err))
	}
	if err := os.WriteFile(entriesAbsolute, raw, 0o644); err != nil {
		fail(fmt.Sprintf("write raw repository entries: %v", err))
	}
	snapshot := model.RepositorySnapshot{
		Schema: model.RepositorySnapshotSchema, HeadSHA: *headSHA, ExecutionID: *executionID,
		EntriesPath: entriesAbsolute, EntriesDigest: model.DigestBytes(raw), PathDigest: model.Digest(entries), EntryCount: len(entries),
	}
	metadata, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		fail(fmt.Sprintf("encode repository snapshot: %v", err))
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(fmt.Sprintf("create snapshot directory: %v", err))
	}
	if err := os.WriteFile(*outputPath, append(metadata, '\n'), 0o644); err != nil {
		fail(fmt.Sprintf("write repository snapshot: %v", err))
	}
	fmt.Printf("repository content snapshot: execution=%s entries=%d digest=%s\n", *executionID, len(entries), snapshot.EntriesDigest)
}

func gitHead(root string) (string, error) {
	command := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("SNAPSHOT_BINDING/observe-head/GIT_HEAD_READ_FAILED: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func repositoryEntries(root string) ([]entry, error) {
	command := exec.Command("git", "-C", root, "ls-files", "-co", "--exclude-standard", "-z")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("REPOSITORY_SNAPSHOT/collect/LS_FILES_FAILED: %w", err)
	}
	paths := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	if len(paths) == 1 && paths[0] == "" {
		return []entry{}, nil
	}
	entries := make([]entry, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		if path == "" || seen[path] {
			return nil, fmt.Errorf("REPOSITORY_SNAPSHOT/collect/DUPLICATE_OR_EMPTY_PATH")
		}
		seen[path] = true
		absolute := filepath.Join(root, filepath.FromSlash(path))
		data, err := os.ReadFile(absolute)
		if err != nil {
			return nil, fmt.Errorf("REPOSITORY_SNAPSHOT/read-content/%s: %w", path, err)
		}
		entries = append(entries, entry{Path: filepath.ToSlash(path), Digest: model.DigestBytes(data)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
