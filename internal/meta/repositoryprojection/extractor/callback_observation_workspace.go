package extractor

import (
	"context"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func callbackObservationSources(root, logical string, proposal CallbackExtractionProposal) (map[string][]byte, map[string][]byte, error) {
	packagePath := path.Dir(logical)
	baseline := map[string][]byte{}
	tree := os.DirFS(filepath.Join(root, filepath.FromSlash(packagePath)))
	err := fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("callback observation refuses non-regular input %s", name)
		}
		raw, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		baseline[name] = raw
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if callbackPreviewDigest(baseline[path.Base(logical)]) != proposal.SourceDigest {
		return nil, nil, fmt.Errorf("callback observation source changed after planning")
	}
	final := maps.Clone(baseline)
	for _, artifact := range proposal.Artifacts {
		if path.Dir(artifact.Path) != packagePath || !fs.ValidPath(artifact.Path) {
			return nil, nil, fmt.Errorf("callback observation artifact escapes source package")
		}
		raw := []byte(artifact.Source)
		if proofDigest(raw) != artifact.Digest || physicalLines(raw) != artifact.Lines {
			return nil, nil, fmt.Errorf("callback observation artifact identity differs")
		}
		final[path.Base(artifact.Path)] = raw
	}
	return baseline, final, nil
}

func callbackObservationToolchain(ctx context.Context, root string) (string, string, error) {
	query := func(args ...string) (string, error) {
		command := exec.CommandContext(ctx, "go", args...)
		command.Dir = root
		command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
		output, err := command.Output()
		return strings.TrimSpace(string(output)), err
	}
	module, err := query("list", "-mod=readonly", "-m", "-f", "{{.Path}}")
	if err != nil {
		return "", "", fmt.Errorf("observe source module: %w", err)
	}
	version, err := query("env", "GOVERSION")
	if err != nil {
		return "", "", fmt.Errorf("observe Go toolchain: %w", err)
	}
	if module == "" || !strings.HasPrefix(version, "go1.27.") {
		return "", "", fmt.Errorf("callback observation requires a source module and Go 1.27 toolchain")
	}
	return module, version, nil
}

func materializeCallbackObservation(parent, variant, root, logical, module, version string, files map[string][]byte) (string, error) {
	directory := filepath.Join(parent, variant)
	for name, raw := range files {
		if !fs.ValidPath(name) {
			return "", fmt.Errorf("callback observation input path is not relative")
		}
		target := filepath.Join(directory, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, raw, 0o600); err != nil {
			return "", err
		}
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	packageModule := module + "/" + path.Dir(logical)
	if path.Dir(logical) == "." {
		packageModule = module + "/callback-observation"
	}
	requiredVersion := "v0.0.0"
	if major, ok := strings.CutPrefix(path.Base(module), "v"); ok {
		if number, err := strconv.Atoi(major); err == nil && number >= 2 {
			requiredVersion = "v" + major + ".0.0"
		}
	}
	goMod := "module " + packageModule + "\n\ngo " + strings.TrimPrefix(version, "go") + "\n\nrequire " + module + " " + requiredVersion + "\nreplace " + module + " => " + strconv.Quote(absoluteRoot) + "\n"
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte(goMod), 0o600); err != nil {
		return "", err
	}
	return directory, nil
}
