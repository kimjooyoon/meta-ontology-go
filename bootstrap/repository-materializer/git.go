package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func gitBytes(root, index, work string, raw bool, input []byte, args ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	command.Env = os.Environ()
	if index != "" {
		command.Env = append(command.Env, "GIT_INDEX_FILE="+index)
	}
	if work != "" {
		command.Env = append(command.Env, "GIT_WORK_TREE="+work)
	}
	if raw {
		command.Env = append(command.Env, "GIT_NO_REPLACE_OBJECTS=1")
	}
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func gitText(root, index, work string, raw bool, input []byte, args ...string) (string, error) {
	output, err := gitBytes(root, index, work, raw, input, args...)
	return strings.TrimSpace(string(output)), err
}

func exactHead(root, expected string) (string, error) {
	head, err := gitText(root, "", "", true, nil, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if !validHex(expected, 40) || head != expected {
		return "", fmt.Errorf("checked-out head %q does not match expected %q", head, expected)
	}
	dirty, err := gitText(root, "", "", true, nil, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return "", err
	}
	if dirty != "" {
		return "", fmt.Errorf("physical Git checkout is not clean")
	}
	return head, nil
}

func trackedBlobs(root string) (map[string]string, error) {
	output, err := gitBytes(root, "", "", true, nil, "ls-files", "--stage", "-z")
	if err != nil {
		return nil, err
	}
	tracked := map[string]string{}
	for _, record := range bytes.Split(output, []byte{0}) {
		tab := bytes.IndexByte(record, '\t')
		if tab < 0 {
			continue
		}
		fields := strings.Fields(string(record[:tab]))
		if len(fields) == 3 {
			tracked[string(record[tab+1:])] = fields[1]
		}
	}
	return tracked, nil
}
