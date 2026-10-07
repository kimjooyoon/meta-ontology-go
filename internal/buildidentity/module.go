// Package buildidentity snapshots metadata embedded in the running compiler.
// These observations do not attest the executable or resolve a Git revision.
package buildidentity

import "runtime/debug"

type ModuleReplacement struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Sum     string `json:"sum,omitempty"`
}

type Module struct {
	Path        string             `json:"path"`
	Version     string             `json:"version"`
	Sum         string             `json:"sum,omitempty"`
	Replacement *ModuleReplacement `json:"replacement,omitempty"`
}

// ModuleFromBuildInfo copies the declared main module and its replacement.
// Empty versions or sums remain unobserved; they are not inferred from paths.
func ModuleFromBuildInfo(info *debug.BuildInfo) *Module {
	if info == nil || info.Main.Path == "" {
		return nil
	}
	m := info.Main
	result := &Module{Path: m.Path, Version: m.Version, Sum: m.Sum}
	if m.Replace != nil {
		result.Replacement = &ModuleReplacement{
			Path: m.Replace.Path, Version: m.Replace.Version, Sum: m.Replace.Sum,
		}
	}
	return result
}

func MainModule() *Module {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return ModuleFromBuildInfo(info)
}
