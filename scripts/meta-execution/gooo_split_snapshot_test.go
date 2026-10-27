package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func splitSnapshotGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	command.Env = replaceEnvironment(childEnvironment("", root), "GIT_WORK_TREE", "")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Git fixture: %v: %s", err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestGoooSplitSnapshotPreservesCommittedInputAndRejectsArchiveProjection(t *testing.T) {
	caller, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_DIR", filepath.Join(caller, ".git"))
	t.Setenv("GIT_WORK_TREE", caller)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(t.TempDir(), "caller-index"))
	root, target := t.TempDir(), filepath.Join(t.TempDir(), "restored")
	splitSnapshotGit(t, root, "init", "-q")
	for _, attributes := range []string{"", "source.gooo export-ignore\n", "source.gooo export-subst\n"} {
		const source = "committed $Format:%H$\n"
		for name, raw := range map[string]string{".gitattributes": attributes, "source.gooo": source} {
			if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
		}
		splitSnapshotGit(t, root, "add", ".")
		splitSnapshotGit(t, root, "-c", "user.name=Snapshot fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "input")
		head := splitSnapshotGit(t, root, "rev-parse", "HEAD")
		if err := os.WriteFile(filepath.Join(root, "source.gooo"), []byte("uncommitted\n"), 0600); err != nil {
			t.Fatal(err)
		}
		err := restoreGoooSplitSnapshot(target, filepath.Join(root, ".git"), head)
		if attributes != "" {
			if err == nil {
				t.Fatal("archive attributes silently changed committed input", attributes)
			}
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(target, "source.gooo"))
		if err != nil || readErr != nil || string(raw) != source {
			t.Fatal("snapshot included caller changes or lost Git PAX metadata", string(raw), err, readErr)
		}
	}
}

func TestGoooSplitSnapshotInventoryRejectsIncompleteMembers(t *testing.T) {
	object := strings.Repeat("a", 40)
	valid := "100644 blob " + object + " 4\tsource.gooo\x00"
	for _, raw := range []string{"", valid + valid, strings.Replace(valid, "100644", "120000", 1),
		strings.Replace(valid, " 4\t", " -1\t", 1), strings.Replace(valid, " 4\t", " 1073741825\t", 1),
		strings.Replace(valid, "source.gooo", "../outside.gooo", 1)} {
		if _, err := parseGoooSnapshotMembers([]byte(raw)); err == nil {
			t.Fatal("invalid snapshot inventory accepted", raw)
		}
	}
}
