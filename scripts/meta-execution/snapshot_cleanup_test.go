package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDisposableSnapshotCleanup8557(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "owned-snapshot")
	child := filepath.Join(root, "snapshot.git", "objects", "pack")
	if err := os.MkdirAll(child, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "fresh-8557.pack"), []byte("temporary source identity 8557"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := makeReadOnlySnapshot(root); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "retained-origin")
	if err := os.Mkdir(outside, 0o500); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		// Add a link after sealing the fixture: cleanup must not chmod its target.
		if err := os.Chmod(root, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "external-origin")); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(root, 0o500); err != nil {
			t.Fatal(err)
		}
	}
	if err := removeReadOnlySnapshot(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("temporary snapshot remains", err)
	}
	info, err := os.Stat(outside)
	if err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o500) {
		t.Fatal("cleanup changed an external origin", err)
	}
	if err := removeReadOnlySnapshot(root); err != nil {
		t.Fatal("already removed snapshot", err)
	}
	t.Log("8557: read-only nested snapshot removed; external origin permission retained; repeated cleanup is empty")
}
