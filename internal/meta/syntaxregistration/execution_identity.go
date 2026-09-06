package syntaxregistration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ObserveExecutionIdentity reads the current generator, Go driver and the
// compiler selected by that driver. It runs no build, test, fix or generator.
func ObserveExecutionIdentity() (ExecutionIdentity, error) {
	executable, err := os.Executable()
	if err != nil {
		return ExecutionIdentity{}, unavailableExecutionIdentity("locate-generator")
	}
	driver, err := exec.LookPath("go")
	if err != nil {
		return ExecutionIdentity{}, unavailableExecutionIdentity("locate-go-driver")
	}
	compiler, err := selectedCompiler(driver)
	if err != nil {
		return ExecutionIdentity{}, err
	}
	identity := ExecutionIdentity{GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	paths := []string{executable, driver, compiler}
	destinations := []*string{&identity.ExecutableDigest, &identity.GoCommandDigest, &identity.CompilerDigest}
	for index, path := range paths {
		value, err := hashExecutionFile(path)
		if err != nil {
			return ExecutionIdentity{}, unavailableExecutionIdentity("hash-execution-binary")
		}
		*destinations[index] = value
	}
	return identity, nil
}

func selectedCompiler(driver string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, driver, "env", "-json", "GOVERSION", "GOTOOLDIR")
	// Match the native candidate evaluator. Do not download/select another Go.
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "GOTOOLCHAIN=") && !strings.HasPrefix(value, "GOWORK=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, "GOTOOLCHAIN=local", "GOWORK=off")
	raw, err := command.Output()
	if ctx.Err() != nil {
		return "", failure("UNKNOWN", "observe-go-driver", "REGISTRATION_TOOLCHAIN_OBSERVATION_TIMEOUT",
			"UNBOUNDED", "restore-observable-local-toolchain")
	}
	if err != nil {
		return "", unavailableExecutionIdentity("observe-go-driver")
	}
	var environment struct {
		Version string `json:"GOVERSION"`
		ToolDir string `json:"GOTOOLDIR"`
	}
	if json.Unmarshal(raw, &environment) != nil || environment.Version == "" || !filepath.IsAbs(environment.ToolDir) {
		return "", failure("UNKNOWN", "observe-go-driver", "REGISTRATION_TOOLCHAIN_OBSERVATION_AMBIGUOUS",
			"AMBIGUOUS", "restore-observable-local-toolchain")
	}
	if environment.Version != runtime.Version() {
		return "", failure("UNKNOWN", "bind-go-driver", "REGISTRATION_TOOLCHAIN_VERSION_STALE",
			"STALE", "select-matching-local-toolchain")
	}
	name := "compile"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(environment.ToolDir, name), nil
}

func hashExecutionFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", os.ErrInvalid
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func unavailableExecutionIdentity(step string) error {
	return failure("UNKNOWN", step, "REGISTRATION_EXECUTION_IDENTITY_UNAVAILABLE",
		"DIRECT_MISSING", "restore-observable-local-toolchain")
}
