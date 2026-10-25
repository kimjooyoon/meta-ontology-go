package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type goooSnapshotMember struct {
	mode   fs.FileMode
	size   int64
	object string
}

// The Git tree, rather than archive attributes, owns the required population.
// export-ignore/export-subst must not silently change the execution inputs.
func goooSnapshotMembers(gitDir, head string) (map[string]goooSnapshotMember, error) {
	command := exec.Command("git", "--git-dir", gitDir, "ls-tree", "-r", "-l", "-z", "--full-tree", head)
	var raw []byte
	err := readGoooSnapshotCommand(gitDir, command, func(input io.Reader) error {
		var err error
		raw, err = io.ReadAll(io.LimitReader(input, (16<<20)+1))
		if len(raw) > 16<<20 {
			return fmt.Errorf("Git tree inventory exceeds 16 MiB")
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return parseGoooSnapshotMembers(raw)
}

func parseGoooSnapshotMembers(raw []byte) (map[string]goooSnapshotMember, error) {
	members := map[string]goooSnapshotMember{}
	var total int64
	for entry := range bytes.SplitSeq(raw, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		name, member, err := parseGoooSnapshotMember(entry)
		if err != nil {
			return nil, err
		}
		if len(members) >= 65536 || member.size > 1<<30-total {
			return nil, fmt.Errorf("Git snapshot exceeds file or byte limit")
		}
		if _, exists := members[name]; exists {
			return nil, fmt.Errorf("duplicate Git tree member")
		}
		members[name] = member
		total += member.size
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("Git snapshot is empty")
	}
	return members, nil
}

func parseGoooSnapshotMember(entry []byte) (string, goooSnapshotMember, error) {
	metadata, name, found := strings.Cut(string(entry), "\t")
	fields := strings.Fields(metadata)
	if !found || len(fields) != 4 || !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\\x00:") {
		return "", goooSnapshotMember{}, fmt.Errorf("invalid Git tree member")
	}
	if fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return "", goooSnapshotMember{}, fmt.Errorf("Git input requires regular files: %s", name)
	}
	size, err := strconv.ParseInt(fields[3], 10, 64)
	object, hashErr := hex.DecodeString(fields[2])
	if err != nil || size < 0 || hashErr != nil || (len(object) != 20 && len(object) != 32) {
		return "", goooSnapshotMember{}, fmt.Errorf("invalid Git object")
	}
	mode := fs.FileMode(0644)
	if fields[0] == "100755" {
		mode = 0755
	}
	return name, goooSnapshotMember{mode, size, fields[2]}, nil
}

func readGoooSnapshotCommand(gitDir string, command *exec.Cmd, consume func(io.Reader) error) error {
	// Git archive must address the complete bound tree, even when the caller is
	// inside a subdirectory of a different work tree (as in materialized CI).
	command.Dir = gitDir
	command.Env = os.Environ()
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"} {
		command.Env = replaceEnvironment(command.Env, key, "")
	}
	pipe, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err = command.Start(); err != nil {
		return err
	}
	err = consume(pipe)
	if err != nil {
		_ = command.Process.Kill()
	}
	waitErr := command.Wait()
	if err != nil {
		return err
	}
	return waitErr
}

func goooBlobHash(member goooSnapshotMember) hash.Hash {
	var result hash.Hash = sha1.New()
	if len(member.object) == 64 {
		result = sha256.New()
	}
	fmt.Fprintf(result, "blob %d\x00", member.size)
	return result
}
