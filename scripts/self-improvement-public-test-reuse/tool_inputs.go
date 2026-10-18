package main

import (
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

type goCommand struct {
	Inputs publictestreuse.ExecutionToolInputs
	Env    []string
}

// Observe the selected executable without starting a child. A wrapper's bytes
// cannot establish which tool it will invoke, so this profile requires cmd/go.
func captureGoCommand() (goCommand, error) {
	path, err := exec.LookPath("go")
	if err != nil {
		return goCommand{}, err
	}
	return observeGoCommand(path, os.Environ())
}

func observeGoCommand(path string, inherited []string) (goCommand, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return goCommand{}, err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return goCommand{}, err
	}
	digest, version, err := nativeGoCommandBytes(path)
	if err != nil {
		return goCommand{}, err
	}
	env := goCommandEnvironment(inherited, filepath.Dir(filepath.Dir(path)))
	inputs := publictestreuse.ExecutionToolInputs{Profile: publictestreuse.ExecutionToolProfile,
		Path: path, ExecutableDigest: digest, Version: version, Platform: runtime.GOOS + "/" + runtime.GOARCH,
		EnvironmentDigest: cache.HashBytes([]byte(strings.Join(env, "\x00"))).String()}
	return goCommand{Inputs: inputs, Env: env}, inputs.Validate()
}

func nativeGoCommandBytes(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() <= 0 || stat.Size() > 128<<20 {
		return "", "", fmt.Errorf("selected Go tool is not a bounded regular executable")
	}
	info, err := buildinfo.Read(file)
	if err != nil || info.Path != "cmd/go" || info.GoVersion != "go1.27.2" {
		return "", "", fmt.Errorf("selected Go tool requires native cmd/go built with go1.27.2")
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["GOOS"] != runtime.GOOS || settings["GOARCH"] != runtime.GOARCH {
		return "", "", fmt.Errorf("selected Go tool does not match the host platform")
	}
	hash := sha256.New()
	n, err := io.Copy(hash, io.LimitReader(file, 128<<20+1))
	if err != nil || n != stat.Size() {
		return "", "", fmt.Errorf("selected Go tool bytes changed or could not be read")
	}
	return hex.EncodeToString(hash.Sum(nil)), info.GoVersion, nil
}

func goCommandEnvironment(inherited []string, root string) []string {
	values := map[string]string{}
	for _, entry := range inherited {
		name, value, present := strings.Cut(entry, "=")
		if !present {
			continue
		}
		switch name {
		case "PATH", "HOME", "TMPDIR", "TMP", "TEMP", "SystemRoot", "SYSTEMROOT", "WINDIR", "USERPROFILE", "LOCALAPPDATA", "GOCACHE":
			values[name] = value
		}
	}
	maps.Copy(values, map[string]string{"GOTOOLCHAIN": "local", "GOENV": "off", "GOWORK": "off",
		"GO111MODULE": "on", "GOPROXY": "off", "GOSUMDB": "off", "CGO_ENABLED": "0",
		"GOFLAGS": "-mod=readonly", "GOROOT": root, "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH,
		"LANG": "C", "LC_ALL": "C", "TZ": "UTC"})
	env := make([]string, 0, len(values))
	for name, value := range values {
		env = append(env, name+"="+value)
	}
	sort.Strings(env)
	return env
}

func (command goCommand) verifyBytes() error {
	current, err := observeGoCommand(command.Inputs.Path, command.Env)
	if err != nil {
		return err
	}
	if current.Inputs != command.Inputs || !slices.Equal(current.Env, command.Env) {
		return fmt.Errorf("selected Go command changed after its observation")
	}
	return nil
}
