package workspace

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Open(host, head string, source State) (*Sandbox, State, error) {
	parent, err := os.MkdirTemp("", "gooo-transformation-effect-*")
	if err != nil {
		return nil, State{}, err
	}
	box := &Sandbox{host: host, parent: parent, Root: filepath.Join(parent, "worktree")}
	if output, err := RunCombined(host, nil, "git", "worktree", "add", "--detach", box.Root, head); err != nil {
		_ = os.RemoveAll(parent)
		return nil, State{}, fmt.Errorf("open disposable worktree: %w: %s", err, output)
	}
	if err := clear(box.Root); err != nil {
		_ = box.Close()
		return nil, State{}, err
	}
	if err := copyTree(source, box.Root); err != nil {
		_ = box.Close()
		return nil, State{}, err
	}
	baseline, err := Scan(box.Root)
	if err != nil || baseline.Digest != source.Digest {
		_ = box.Close()
		return nil, State{}, fmt.Errorf("sandbox baseline does not match source workspace")
	}
	return box, baseline, nil
}

func clear(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != ".git" {
			if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (box *Sandbox) Close() error {
	_, worktreeErr := RunCombined(box.host, nil, "git", "worktree", "remove", "--force", box.Root)
	return errors.Join(worktreeErr, os.RemoveAll(box.parent))
}

func RunCombined(root string, env []string, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = root
	if env != nil {
		command.Env = env
	}
	return command.CombinedOutput()
}
