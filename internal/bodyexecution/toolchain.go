package bodyexecution

import (
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/nativeobservation"
)

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
	file, err := openDigestFile(goBinary)
	if err != nil {
		return false
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() > executableFileLimit {
		return false
	}
	info, err := buildinfo.Read(file)
	if err != nil || info.Path != "cmd/go" || info.GoVersion != "go1.27.3" {
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
	o.SourceCheckSHA256 = nativeobservation.SourceCheckBindingDigest(*o)
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
	expected := "go version go1.27.3 " + runtime.GOOS + "/" + runtime.GOARCH
	if r.GoVersion != expected {
		observed := r.GoVersion
		if len(observed) > 128 {
			observed = observed[:128] + "..."
		}
		return fmt.Errorf("native execution requires %q; selected %q. Match Go 1.27.3 to this Gooo binary's operating system and architecture, then pass that executable with --go-bin (or --go for package execute)", expected, observed)
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
