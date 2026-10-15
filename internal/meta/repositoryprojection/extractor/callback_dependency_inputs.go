package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// These are observed selected source bytes, not an immutable execution sandbox.
// Standard-library, native-library and dynamic test-data inputs remain unbound.
type CallbackDependencyInputs struct {
	Scope    string                      `json:"scope"`
	Digest   string                      `json:"digest"`
	Packages []CallbackDependencyPackage `json:"packages"`
	Bytes    int64                       `json:"bytes"`
}

type CallbackDependencyPackage struct {
	ImportPath string                   `json:"import_path"`
	ForTest    string                   `json:"for_test,omitempty"`
	Module     CallbackDependencyModule `json:"module"`
	ModuleFile string                   `json:"module_declaration_digest"`
	Files      map[string]string        `json:"files"`
}

type CallbackDependencyModule struct {
	Path    string                    `json:"path"`
	Version string                    `json:"version,omitempty"`
	Sum     string                    `json:"sum,omitempty"`
	Replace *CallbackDependencyModule `json:"replace,omitempty"`
}

type callbackListedModule struct {
	Path, Version, Sum, GoMod string
	Main                      bool
	Replace                   *callbackListedModule
}

type callbackListedPackage struct {
	Dir, ImportPath, ForTest                                    string
	Goroot, Incomplete                                          bool
	Module                                                      *callbackListedModule
	GoFiles, CgoFiles, TestGoFiles, XTestGoFiles                []string
	CFiles, CXXFiles, MFiles, HFiles, FFiles, SFiles, SwigFiles []string
	SwigCXXFiles, SysoFiles, EmbedFiles                         []string
	TestEmbedFiles, XTestEmbedFiles                             []string
}

func callbackObservationDependencyInputs(ctx context.Context, directory string) (CallbackDependencyInputs, error) {
	command := exec.CommandContext(ctx, "go", "list", "-mod=readonly", "-deps", "-test", "-json", ".")
	command.Dir = directory
	command.Env = callbackObservationEnvironment()
	configureModuleListProcess(command)
	stdout, stderr := &boundedBuffer{limit: moduleListMax}, &boundedBuffer{limit: 32 << 10}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		return CallbackDependencyInputs{}, fmt.Errorf("observe dependency selection: %w; %s", err, stderr.Bytes())
	}
	if stdout.exceeded || stderr.exceeded {
		return CallbackDependencyInputs{}, fmt.Errorf("dependency selection exceeds output budget")
	}
	return callbackDependencyInputManifest(ctx, stdout.Bytes())
}

func callbackDependencyInputManifest(ctx context.Context, raw []byte) (CallbackDependencyInputs, error) {
	inputs := CallbackDependencyInputs{Scope: "SELECTED_EXTERNAL_PACKAGE_FILES", Packages: []CallbackDependencyPackage{}}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	seen := map[string]bool{}
	count := 0
	for ; ; count++ {
		if err := ctx.Err(); err != nil {
			return inputs, err
		}
		var listed callbackListedPackage
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return inputs, fmt.Errorf("decode dependency selection: %w", err)
		}
		if count >= moduleListPackagesMax || listed.ImportPath == "" || listed.Incomplete {
			return inputs, fmt.Errorf("dependency selection is incomplete or exceeds package budget")
		}
		if listed.Goroot || listed.Module != nil && listed.Module.Main {
			continue
		}
		if listed.Module == nil || listed.Dir == "" {
			return inputs, fmt.Errorf("external dependency has no module or source directory")
		}
		key := listed.ImportPath + "\x00" + listed.ForTest
		if seen[key] {
			return inputs, fmt.Errorf("duplicate external dependency selection")
		}
		seen[key] = true
		item := CallbackDependencyPackage{ImportPath: listed.ImportPath, ForTest: listed.ForTest, Files: map[string]string{}}
		var err error
		item.Module, err = callbackDependencyModuleIdentity(listed.Module, 0)
		if err != nil {
			return inputs, err
		}
		// Go's test variants already merge active test files and embeds into GoFiles/EmbedFiles.
		groups := [][]string{listed.GoFiles, listed.CgoFiles,
			listed.CFiles, listed.CXXFiles, listed.MFiles, listed.HFiles, listed.FFiles, listed.SFiles,
			listed.SwigFiles, listed.SwigCXXFiles, listed.SysoFiles, listed.EmbedFiles}
		for _, files := range groups {
			for _, name := range files {
				if _, exists := item.Files[name]; exists {
					continue
				}
				if !fs.ValidPath(name) || len(item.Files) >= moduleFilesMax {
					return inputs, fmt.Errorf("dependency file selection is invalid or exceeds file budget")
				}
				item.Files[name], err = callbackDependencyFileDigest(ctx, filepath.Join(listed.Dir, filepath.FromSlash(name)), &inputs.Bytes)
				if err != nil {
					return inputs, err
				}
			}
		}
		effective := listed.Module
		if effective.Replace != nil {
			effective = effective.Replace
		}
		if effective.GoMod == "" {
			return inputs, fmt.Errorf("dependency module declaration is unavailable")
		}
		item.ModuleFile, err = callbackDependencyFileDigest(ctx, effective.GoMod, &inputs.Bytes)
		if err != nil {
			return inputs, err
		}
		inputs.Packages = append(inputs.Packages, item)
	}
	if count == 0 {
		return inputs, fmt.Errorf("dependency selection is empty")
	}
	slices.SortFunc(inputs.Packages, func(left, right CallbackDependencyPackage) int {
		if order := strings.Compare(left.ImportPath, right.ImportPath); order != 0 {
			return order
		}
		return strings.Compare(left.ForTest, right.ForTest)
	})
	payload, err := json.Marshal(inputs.Packages)
	if err != nil {
		return inputs, err
	}
	inputs.Digest = proofDigest(payload)
	return inputs, nil
}

func callbackDependencyModuleIdentity(module *callbackListedModule, depth int) (CallbackDependencyModule, error) {
	if module == nil || module.Path == "" || depth > 1 {
		return CallbackDependencyModule{}, fmt.Errorf("dependency module identity is incomplete")
	}
	identity := CallbackDependencyModule{Path: module.Path, Version: module.Version, Sum: module.Sum}
	if module.Replace != nil {
		replacement, err := callbackDependencyModuleIdentity(module.Replace, depth+1)
		if err != nil {
			return identity, err
		}
		// Files bind a local replacement's contents; its checkout location is not an identity.
		if replacement.Version == "" {
			replacement.Path = "LOCAL_SOURCE_BYTES"
		}
		identity.Replace = &replacement
	}
	return identity, nil
}

func callbackDependencyFileDigest(ctx context.Context, name string, total *int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > moduleFileMax {
		return "", fmt.Errorf("dependency source is unavailable, non-regular or exceeds file budget")
	}
	raw, err := io.ReadAll(io.LimitReader(file, moduleFileMax+1))
	*total += int64(len(raw))
	if err != nil || len(raw) > moduleFileMax || *total > moduleListTotalMax {
		return "", fmt.Errorf("dependency source exceeds input byte budget")
	}
	return proofDigest(raw), nil
}

func compareCallbackDependencyInputs(observation *CallbackExtractionObservation, expected string,
	before, after CallbackDependencyInputs) error {
	if before.Digest == "" || after.Digest == "" {
		return fmt.Errorf("dependency input identity is unavailable")
	}
	if before.Digest == after.Digest && (expected == "" || before.Digest == expected) {
		return nil
	}
	observation.Decision = "REFUTED"
	observation.Frontier = CallbackExtractionClaim{State: "REFUTED", Stage: "CALLBACK_OBSERVATION",
		Step: "COMPARE_DEPENDENCY_INPUTS", Reason: "SELECTED_DEPENDENCY_INPUTS_CHANGED",
		NextOperation: "PRESERVE_DEPENDENCY_INPUT_COUNTEREXAMPLE", BlockedBy: []string{}}
	return fmt.Errorf("selected external package inputs changed during callback observation")
}
