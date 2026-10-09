package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/buildidentity"
)

const decisionRuntimeModule = "github.com/kimjooyoon/gooo-decision-runtime"

type versionSDKInfo struct {
	ModulePath         string `json:"module_path"`
	Version            string `json:"version"`
	Replaced           bool   `json:"replaced"`
	ReplacementPath    string `json:"replacement_path,omitempty"`
	ReplacementVersion string `json:"replacement_version,omitempty"`
}

// Build metadata describes the running executable. It does not observe the Go
// executable selected by a later native request.
type versionBuildInfo struct {
	Schema            string                `json:"schema"`
	Language          string                `json:"language"`
	Version           string                `json:"version"`
	Status            string                `json:"status"`
	BuildInfoObserved bool                  `json:"build_info_observed"`
	GoVersion         string                `json:"go_version"`
	ModuleVersion     string                `json:"module_version"`
	Module            *buildidentity.Module `json:"module,omitempty"`
	CompilerSourceSHA string                `json:"compiler_source_sha"`
	SourceStatus      string                `json:"source_status"`
	VCSRevision       string                `json:"vcs_revision"`
	VCSModified       string                `json:"vcs_modified"`
	SDK               versionSDKInfo        `json:"decision_runtime"`
	NativeGoRequired  string                `json:"native_go_required"`
}

func versionBuildMetadata(info *debug.BuildInfo, runtimeVersion string) versionBuildInfo {
	result := versionBuildInfo{Schema: "gooo/build-identity/v1", Language: "gooo", Version: goooVersion,
		Status: versionStatus, GoVersion: runtimeVersion, ModuleVersion: "unobserved",
		CompilerSourceSHA: "UNBOUND_LOCAL_SOURCE", SourceStatus: "UNOBSERVED_VCS",
		VCSRevision: "unobserved", VCSModified: "unobserved", NativeGoRequired: "go1.27.2",
		SDK: versionSDKInfo{ModulePath: decisionRuntimeModule, Version: "unobserved"}}
	if info == nil {
		return result
	}
	result.BuildInfoObserved = true
	result.Module = buildidentity.ModuleFromBuildInfo(info)
	if info.GoVersion != "" {
		result.GoVersion = info.GoVersion
	}
	if info.Main.Version != "" {
		result.ModuleVersion = info.Main.Version
	}
	bindVersionVCS(&result, info.Settings)
	result.SDK = versionRuntimeDependency(info.Deps)
	return result
}

func bindVersionVCS(result *versionBuildInfo, settings []debug.BuildSetting) {
	for _, setting := range settings {
		switch setting.Key {
		case "vcs.revision":
			if setting.Value != "" {
				result.VCSRevision = setting.Value
			}
		case "vcs.modified":
			if setting.Value != "" {
				result.VCSModified = setting.Value
			}
		}
	}
	if result.VCSModified == "true" {
		result.SourceStatus = "MODIFIED_VCS"
	}
	if result.VCSModified == "false" && versionBuildCommitSHA(result.VCSRevision) {
		result.SourceStatus, result.CompilerSourceSHA = "CLEAN_VCS", result.VCSRevision
	}
}

func versionRuntimeDependency(deps []*debug.Module) versionSDKInfo {
	result := versionSDKInfo{ModulePath: decisionRuntimeModule, Version: "unobserved"}
	for _, dep := range deps {
		if dep == nil || dep.Path != decisionRuntimeModule {
			continue
		}
		if dep.Version != "" {
			result.Version = dep.Version
		}
		if dep.Replace != nil {
			result.Replaced = true
			result.ReplacementPath = dep.Replace.Path
			result.ReplacementVersion = dep.Replace.Version
			if dep.Replace.Version == "" {
				result.ReplacementVersion = "unobserved"
			}
		}
		break
	}
	return result
}

func versionBuildCommitSHA(revision string) bool {
	if len(revision) != 40 || strings.ToLower(revision) != revision {
		return false
	}
	decoded, err := hex.DecodeString(revision)
	return err == nil && len(decoded) == 20
}

func runBuildVersion(asJSON bool, stdout io.Writer) int {
	info, _ := debug.ReadBuildInfo()
	identity := versionBuildMetadata(info, runtime.Version())
	if asJSON {
		if err := json.NewEncoder(stdout).Encode(identity); err != nil {
			return exitFailure
		}
		return exitOK
	}
	_, err := fmt.Fprintf(stdout, "gooo %s (%s)\ncompiler source: %s (%s)\ncompiler module: %s\nbuilt with: %s\ndecision runtime: %s\n",
		identity.Version, identity.Status, identity.CompilerSourceSHA, identity.SourceStatus,
		identity.ModuleVersion, identity.GoVersion, identity.SDK.Version)
	if err != nil {
		return exitFailure
	}
	if identity.SDK.Replaced {
		if _, err := fmt.Fprintf(stdout, "runtime replacement: %q @ %q\n", identity.SDK.ReplacementPath, identity.SDK.ReplacementVersion); err != nil {
			return exitFailure
		}
	}
	if identity.SourceStatus != "CLEAN_VCS" {
		if _, err := fmt.Fprintln(stdout, "source binding: build from a clean Git checkout to bind source receipts"); err != nil {
			return exitFailure
		}
	}
	if _, err := fmt.Fprintln(stdout, "native execution: Go1.27.2 selected automatically; --go-bin chooses an explicit executable"); err != nil {
		return exitFailure
	}
	return exitOK
}
