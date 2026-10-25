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

// Restore the exact Git tree for each pass. The caller's generated reports,
// untracked inputs and previous pass outputs never become execution inputs.
func restoreGoooSplitSnapshot(target, gitDir, head string) error {
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	if err := os.Mkdir(target, 0700); err != nil {
		return err
	}
	command := exec.Command("git", "--git-dir", gitDir, "archive", "--format=tar", head)
	pipe, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}
	err = unpackGoooSplitSnapshot(target, tar.NewReader(pipe))
	if err != nil {
		_ = command.Process.Kill()
	}
	waitErr := command.Wait()
	if err != nil {
		return err
	}
	return waitErr
}

func unpackGoooSplitSnapshot(target string, archive *tar.Reader) error {
	var total int64
	for count := 0; ; count++ {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := header.Name
		if header.Typeflag == tar.TypeDir {
			name = filepath.ToSlash(filepath.Clean(name))
		}
		if count >= 65536 || !fs.ValidPath(name) || name == "." || header.Size < 0 {
			return fmt.Errorf("invalid or excessive Git snapshot member")
		}
		total += header.Size
		if total > 1<<30 {
			return fmt.Errorf("Git snapshot exceeds 1 GiB")
		}
		address := filepath.Join(target, filepath.FromSlash(name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(address, 0700); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeGoooSnapshotFile(address, header, archive); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Git snapshot requires regular files: %s", name)
		}
	}
}

func writeGoooSnapshotFile(address string, header *tar.Header, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(address), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(address, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fs.FileMode(header.Mode)&0777)
	if err != nil {
		return err
	}
	_, err = io.CopyN(file, reader, header.Size)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
