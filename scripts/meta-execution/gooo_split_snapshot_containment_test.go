package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func unpackContainmentFixture(t *testing.T, root, name string, kind byte) error {
	t.Helper()
	var raw bytes.Buffer
	writer := tar.NewWriter(&raw)
	payload := []byte("original source\n")
	header := &tar.Header{Name: name, Typeflag: kind, Mode: 0644}
	members := map[string]goooSnapshotMember{}
	if kind == tar.TypeReg {
		header.Size = int64(len(payload))
		member := goooSnapshotMember{mode: 0644, size: header.Size}
		hash := goooBlobHash(member)
		hash.Write(payload)
		member.object = fmt.Sprintf("%x", hash.Sum(nil))
		members[name] = member
	}
	if err := writer.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if kind == tar.TypeReg {
		if _, err := writer.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return unpackGoooSplitSnapshot(root, tar.NewReader(&raw), "bound-head", members)
}

func TestGoooSplitSnapshotConfinesExistingParentLinks(t *testing.T) {
	for _, kind := range []byte{tar.TypeDir, tar.TypeReg} {
		t.Run(string(kind), func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
				t.Skipf("symlink fixture unavailable: %v", err)
			}
			if err := unpackContainmentFixture(t, root, "linked/escaped", kind); err == nil {
				t.Error("archive followed a parent link outside its root")
			}
			entries, err := os.ReadDir(outside)
			if err != nil || len(entries) != 0 {
				t.Fatal("archive changed the outside directory", err, len(entries))
			}
		})
	}
}

func TestGoooSplitSnapshotRejectsNoncanonicalDirectories(t *testing.T) {
	for _, name := range []string{"child/../kept/", "kept//", "../escaped/", "/absolute/", "parent\\child/", "volume:child/"} {
		if err := unpackContainmentFixture(t, t.TempDir(), name, tar.TypeDir); err == nil {
			t.Error("directory spelling normalized instead of rejected", name)
		}
	}
}

func TestGoooSplitSnapshotRestoresNestedCommittedBytes(t *testing.T) {
	root := t.TempDir()
	if err := unpackContainmentFixture(t, root, "nested/source.gooo", tar.TypeReg); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "nested", "source.gooo"))
	if err != nil || string(raw) != "original source\n" {
		t.Fatal("confined extraction changed committed bytes", err)
	}
}
