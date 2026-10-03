package bodyexecution

import (
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

// ToolchainObservation keeps the original version process separate from work
// in this call. Reuse is available only for a native cmd/go executable whose
// embedded release and platform agree with the first actual version response.
type ToolchainObservation struct {
	Schema            string             `json:"schema"`
	KeySHA256         string             `json:"key_sha256"`
	GoToolSHA256      string             `json:"go_tool_sha256"`
	GoVersion         string             `json:"go_version"`
	Reused            bool               `json:"reused"`
	BytesVerified     bool               `json:"bytes_verified"`
	BuildInfoVerified bool               `json:"build_info_verified"`
	SourceCheck       ProcessObservation `json:"source_check"`
	SourceOutput      []byte             `json:"source_output"`
	SourceCheckSHA256 string             `json:"source_check_sha256"`
}

type ownedToolchain struct {
	key         string
	observation ToolchainObservation
}

func toolchainKey(goBinary string, r *Observation) string {
	env := childEnvironment()
	sort.Strings(env)
	key, _ := json.Marshal([]any{"gooo/native-cmd-go-version/v1", goBinary, r.GoToolSHA256,
		runtime.GOOS, runtime.GOARCH, runtime.Version(), r.ProducerSourceSHA, env})
	return digest(key)
}

func nativeGoBuildInfo(goBinary string) bool {
	info, err := buildinfo.ReadFile(goBinary)
	if err != nil || info.Path != "cmd/go" || info.GoVersion != "go1.27.1" {
		return false
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return settings["GOOS"] == runtime.GOOS && settings["GOARCH"] == runtime.GOARCH
}

func cloneToolchain(o ToolchainObservation) ToolchainObservation {
	o.SourceCheck = copyProcess(o.SourceCheck)
	o.SourceOutput = append([]byte(nil), o.SourceOutput...)
	return o
}

func bindToolchain(o *ToolchainObservation) {
	raw, _ := json.Marshal([]any{o.KeySHA256, o.GoToolSHA256, o.GoVersion, o.SourceCheck, o.SourceOutput})
	o.SourceCheckSHA256 = digest(raw)
}

func observeToolchain(ctx context.Context, goBinary string, r *Observation, owner *Executor) error {
	key := toolchainKey(goBinary, r)
	eligible := owner != nil && nativeGoBuildInfo(goBinary)
	if eligible && owner.toolchain != nil && owner.toolchain.key == key {
		reference := cloneToolchain(owner.toolchain.observation)
		reference.Reused = true
		r.GoVersion, r.ToolchainReference = reference.GoVersion, &reference
		return nil
	}
	// A changed/non-native tool always gets a new actual version process. A shell
	// wrapper's other dependencies cannot be established from its own byte hash.
	if owner != nil {
		owner.toolchain = nil
	}
	version, observed, err := process(ctx, "", goBinary, nil, "version")
	r.Toolchain = observed
	if err != nil {
		return err
	}
	r.GoVersion = strings.TrimSpace(string(version))
	if r.GoVersion != "go version go1.27.1 "+runtime.GOOS+"/"+runtime.GOARCH {
		return fmt.Errorf("runtime requires the local Go 1.27.1 toolchain")
	}
	if eligible {
		reference := ToolchainObservation{Schema: "gooo/owned-go-toolchain/v1", KeySHA256: key,
			GoToolSHA256: r.GoToolSHA256, GoVersion: r.GoVersion, BytesVerified: true, BuildInfoVerified: true,
			SourceCheck: copyProcess(observed), SourceOutput: append([]byte(nil), version...)}
		bindToolchain(&reference)
		owner.toolchain = &ownedToolchain{key: key, observation: cloneToolchain(reference)}
		r.ToolchainReference = &reference
	}
	return nil
}
