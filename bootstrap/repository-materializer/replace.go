package main

import (
	"bytes"
	"fmt"
)

func bindLogicalHead(root, index, work, head, logicalTree string) (string, error) {
	physicalTree, err := gitText(root, index, work, true, nil, "rev-parse", head+"^{tree}")
	if err != nil {
		return "", err
	}
	if physicalTree == logicalTree {
		return "", nil
	}
	commit, err := gitBytes(root, index, work, true, nil, "cat-file", "commit", head)
	if err != nil {
		return "", err
	}
	newline := bytes.IndexByte(commit, '\n')
	if newline < 0 || !bytes.HasPrefix(commit, []byte("tree ")) {
		return "", fmt.Errorf("malformed physical commit %s", head)
	}
	replacementBody := append([]byte("tree "+logicalTree), commit[newline:]...)
	replacement, err := gitText(root, index, work, true, replacementBody, "hash-object", "-t", "commit", "-w", "--stdin")
	if err != nil {
		return "", err
	}
	_, err = gitBytes(root, index, work, true, nil, "update-ref", "refs/replace/"+head, replacement)
	return replacement, err
}
