package transformationeffect

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type sandbox struct {
	host   string
	parent string
	root   string
}

func openSandbox(host, head string, source treeState) (*sandbox, treeState, error) {
	parent, err := os.MkdirTemp("", "gooo-transformation-effect-*")
	if err != nil {
		return nil, treeState{}, err
	}
	box := &sandbox{host: host, parent: parent, root: filepath.Join(parent, "worktree")}
	if output, err := runCombined(host, nil, "git", "worktree", "add", "--detach", box.root, head); err != nil {
		_ = os.RemoveAll(parent)
		return nil, treeState{}, fmt.Errorf("open disposable worktree: %w: %s", err, output)
	}
	if err := clearSandbox(box.root); err != nil {
		_ = box.close()
		return nil, treeState{}, err
	}
	if err := copyTree(source, box.root); err != nil {
		_ = box.close()
		return nil, treeState{}, err
	}
	baseline, err := scanTree(box.root)
	if err != nil || baseline.Digest != source.Digest {
		_ = box.close()
		return nil, treeState{}, fmt.Errorf("sandbox baseline does not match source workspace")
	}
	return box, baseline, nil
}

func clearSandbox(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (box *sandbox) close() error {
	_, worktreeErr := runCombined(box.host, nil, "git", "worktree", "remove", "--force", box.root)
	return errors.Join(worktreeErr, os.RemoveAll(box.parent))
}

func runCombined(root string, env []string, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = root
	if env != nil {
		command.Env = env
	}
	return command.CombinedOutput()
}
