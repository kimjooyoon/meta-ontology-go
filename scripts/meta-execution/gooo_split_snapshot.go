package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// Each pass restores the same committed file population and Git blob bytes.
func restoreGoooSplitSnapshot(target, gitDir, head string) error {
	members, err := goooSnapshotMembers(gitDir, head)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	if err := os.Mkdir(target, 0700); err != nil {
		return err
	}
	command := exec.Command("git", "--git-dir", gitDir, "archive", "--format=tar", head)
	return readGoooSnapshotCommand(gitDir, command, func(input io.Reader) error {
		return unpackGoooSplitSnapshot(target, tar.NewReader(input), head, members)
	})
}

func unpackGoooSplitSnapshot(target string, archive *tar.Reader, head string, members map[string]goooSnapshotMember) error {
	for count := 0; ; count++ {
		header, err := archive.Next()
		if err == io.EOF {
			if len(members) != 0 {
				return fmt.Errorf("Git archive omitted %d committed files", len(members))
			}
			return nil
		}
		if err != nil {
			return err
		}
		if count >= 65536 {
			return fmt.Errorf("excessive Git snapshot member count")
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			if count != 0 || len(header.PAXRecords) != 1 || header.PAXRecords["comment"] != head {
				return fmt.Errorf("Git archive metadata differs from committed input")
			}
			continue
		}
		name := header.Name
		if header.Typeflag == tar.TypeDir {
			name = filepath.ToSlash(filepath.Clean(name))
		}
		if !fs.ValidPath(name) || name == "." || header.Size < 0 {
			return fmt.Errorf("invalid Git snapshot member")
		}
		address := filepath.Join(target, filepath.FromSlash(name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(address, 0700); err != nil {
				return err
			}
		case tar.TypeReg:
			member, found := members[name]
			if !found || member.size != header.Size || (member.mode&0111 != 0) != (fs.FileMode(header.Mode)&0111 != 0) {
				return fmt.Errorf("Git archive member differs: %s", name)
			}
			if err := writeGoooSnapshotFile(address, member, archive); err != nil {
				return err
			}
			delete(members, name)
		default:
			return fmt.Errorf("Git snapshot requires regular files: %s", name)
		}
	}
}

func writeGoooSnapshotFile(address string, member goooSnapshotMember, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(address), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(address, os.O_WRONLY|os.O_CREATE|os.O_EXCL, member.mode)
	if err != nil {
		return err
	}
	digest := goooBlobHash(member)
	_, err = io.CopyN(io.MultiWriter(file, digest), reader, member.size)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if fmt.Sprintf("%x", digest.Sum(nil)) != member.object {
		return fmt.Errorf("Git archive changed committed bytes: %s", address)
	}
	return nil
}
