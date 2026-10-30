package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
)

// Resolve each variant through Go rather than copying a fixed list of roots.
// Dynamic test data is still an explicit caller input; external modules retain
// the ordinary module-cache behavior of the existing readonly test command.
func copyCallbackPreviewModuleInputs(ctx context.Context, root, destination, logical, candidate string) error {
	if !fs.ValidPath(logical) || path.Ext(logical) != ".go" {
		return fmt.Errorf("callback fixture source is not a relative Go path")
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	overlay, err := os.MkdirTemp("", "callback-fixture-overlay-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(overlay)
	args, err := callbackFixtureListArguments(root, logical, candidate, overlay)
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir, command.Env = root, callbackObservationEnvironment()
	configureModuleListProcess(command)
	stdout, stderr := &boundedBuffer{limit: moduleListMax}, &boundedBuffer{limit: 32 << 10}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("select callback fixture inputs: %w; %s", err, stderr.Bytes())
	}
	if stdout.exceeded || stderr.exceeded {
		return fmt.Errorf("callback fixture selection exceeds existing output budget")
	}
	if err := copyCallbackFixtureSelection(ctx, root, destination, logical, stdout.Bytes()); err != nil {
		return err
	}
	if candidate != "" {
		return os.WriteFile(filepath.Join(destination, filepath.FromSlash(logical)), []byte(candidate), 0o644)
	}
	return nil
}

func callbackFixtureListArguments(root, logical, candidate, overlay string) ([]string, error) {
	args := []string{"list", "-mod=readonly", "-deps", "-test", "-json"}
	if candidate != "" {
		file := filepath.Join(overlay, "candidate.go")
		if err := os.WriteFile(file, []byte(candidate), 0o600); err != nil {
			return nil, err
		}
		raw, err := json.Marshal(struct{ Replace map[string]string }{
			Replace: map[string]string{filepath.Join(root, filepath.FromSlash(logical)): file},
		})
		if err != nil {
			return nil, err
		}
		file = filepath.Join(overlay, "overlay.json")
		if err := os.WriteFile(file, raw, 0o600); err != nil {
			return nil, err
		}
		args = append(args, "-overlay="+file)
	}
	return append(args, "./"+path.Dir(logical)), nil
}

func copyCallbackFixtureSelection(ctx context.Context, root, destination, logical string, raw []byte) error {
	files := map[string]bool{"go.mod": true}
	if _, err := os.Stat(filepath.Join(root, "go.sum")); err == nil {
		files["go.sum"] = true
	} else if !os.IsNotExist(err) {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for count := 0; ; count++ {
		var listed callbackListedPackage
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if count >= moduleListPackagesMax || listed.ImportPath == "" || listed.Incomplete {
			return fmt.Errorf("callback fixture selection is incomplete or exceeds package budget")
		}
		if listed.Goroot || listed.Module == nil || !listed.Module.Main {
			continue
		}
		if err := addCallbackFixtureFiles(root, logical, listed, files); err != nil {
			return err
		}
	}
	if !files[logical] {
		return fmt.Errorf("callback fixture selection omits its authored source")
	}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := copyCallbackPreviewPath(root, destination, name); err != nil {
			return err
		}
	}
	return nil
}

func addCallbackFixtureFiles(root, logical string, listed callbackListedPackage, files map[string]bool) error {
	if filepath.Clean(listed.Module.GoMod) != filepath.Join(root, "go.mod") {
		return fmt.Errorf("callback fixture selected another main module")
	}
	target := listed.Module.Path
	if directory := path.Dir(logical); directory != "." {
		target += "/" + directory
	}
	// Go supplies this synthetic binary's source from its build cache.
	if listed.Name == "main" && listed.ImportPath == target+".test" {
		return nil
	}
	directory, err := filepath.Rel(root, listed.Dir)
	if err != nil || !fs.ValidPath(filepath.ToSlash(directory)) {
		return fmt.Errorf("callback fixture package is outside its module")
	}
	for _, group := range callbackDependencySourceGroups(listed) {
		for _, name := range group {
			if !fs.ValidPath(name) {
				return fmt.Errorf("callback fixture selected a non-relative source file")
			}
			files[path.Join(filepath.ToSlash(directory), name)] = true
		}
	}
	return nil
}
