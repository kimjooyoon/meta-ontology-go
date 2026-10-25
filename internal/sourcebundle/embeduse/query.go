package embeduse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
)

const maxFiles = 1024
const maxConsumers = 64

type listedPackage struct {
	Dir, ImportPath                                                                                   string
	GoFiles, CgoFiles, TestGoFiles, XTestGoFiles, IgnoredGoFiles                                      []string
	EmbedPatterns, EmbedFiles, TestEmbedPatterns, TestEmbedFiles, XTestEmbedPatterns, XTestEmbedFiles []string
	Incomplete                                                                                        bool
	Error                                                                                             *struct{ Err string }
}

type observer struct {
	projectedNames []string
	ctx            context.Context
	root, work     string
	importPath     string
	env            []string
	projections    map[string]string
	inputs         map[string]int
	bytes, serial  int
	report         Report
}

func (o *observer) list(replace map[string]string) (listedPackage, error) {
	var selected listedPackage
	raw, err := json.Marshal(struct{ Replace map[string]string }{replace})
	if err != nil {
		return selected, err
	}
	path, err := o.backing(raw)
	if err != nil {
		return selected, err
	}
	args := []string{"list", "-mod=readonly", "-overlay", path,
		"-json=Dir,ImportPath,GoFiles,CgoFiles,TestGoFiles,XTestGoFiles,IgnoredGoFiles,EmbedPatterns,EmbedFiles,TestEmbedPatterns,TestEmbedFiles,XTestEmbedPatterns,XTestEmbedFiles,Incomplete,Error"}
	if o.importPath == "" {
		if err := o.query(&selected, append(args, ".")...); err != nil {
			return listedPackage{}, err
		}
		if selected.Dir != o.root || selected.ImportPath == "" {
			return listedPackage{}, fmt.Errorf("Go package address is not exact")
		}
		o.importPath = selected.ImportPath
	}
	raw, err = o.run(append(args, "-test", ".")...)
	if err != nil {
		return listedPackage{}, err
	}
	selected, err = selectOriginalPackage(raw, o.importPath)
	if err != nil {
		return listedPackage{}, err
	}
	if selected.Dir != o.root || selected.ImportPath == "" || selected.Incomplete || selected.Error != nil {
		return listedPackage{}, fmt.Errorf("Go package selection is incomplete or addresses a different directory")
	}
	return selected, nil
}

func (o *observer) query(value any, args ...string) error {
	raw, err := o.run(args...)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("decode Go embed query: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("Go embed query has trailing JSON")
	}
	return nil
}

func (o *observer) run(args ...string) ([]byte, error) {
	command := exec.CommandContext(o.ctx, "go", args...)
	command.Dir, command.Env = o.root, o.env
	command.WaitDelay = 2e9
	stdout, stderr := &boundedOutput{limit: 1 << 20}, &boundedOutput{limit: 32 << 10}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("Go embed query: %w: %s", err, stderr.Bytes())
	}
	return stdout.Bytes(), nil
}

func selectOriginalPackage(raw []byte, importPath string) (listedPackage, error) {
	decoder, found := json.NewDecoder(bytes.NewReader(raw)), false
	var selected listedPackage
	for count := 0; ; count++ {
		var item listedPackage
		if err := decoder.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return listedPackage{}, err
		}
		if count >= 8 {
			return listedPackage{}, fmt.Errorf("Go test package selection exceeds variant limit")
		}
		if item.ImportPath != importPath {
			continue
		}
		if found {
			return listedPackage{}, fmt.Errorf("duplicate Go package selection")
		}
		selected, found = item, true
	}
	if !found {
		return listedPackage{}, fmt.Errorf("Go test selection omitted the source package")
	}
	return selected, nil
}

type boundedOutput struct {
	buffer bytes.Buffer
	limit  int
}

func (output *boundedOutput) Bytes() []byte { return output.buffer.Bytes() }

func (output *boundedOutput) Write(raw []byte) (int, error) {
	if len(raw) > output.limit-output.buffer.Len() {
		return 0, fmt.Errorf("Go embed query output exceeds byte limit")
	}
	return output.buffer.Write(raw)
}

func (o *observer) address(name string) (string, error) {
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("nonlocal Go source address: %q", name)
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(filepath.Join(o.root, name)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(o.root, parent)
	if err != nil || !filepath.IsLocal(relative) {
		return "", fmt.Errorf("Go source parent leaves package directory")
	}
	return filepath.Join(parent, filepath.Base(name)), nil
}
