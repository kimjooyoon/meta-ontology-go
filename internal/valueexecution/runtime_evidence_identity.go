package valueexecution

import (
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
)

// RuntimeEvidenceIdentities reports the Go runtime/toolchain identity and the
// exact clean VCS build identity used by this evaluator. An evaluator digest
// is unavailable for modified or non-VCS builds so evidence cannot claim a
// reproducible replay from an unbound binary.
func RuntimeEvidenceIdentities() (toolchainDigest, evaluatorDigest string) {
	toolchainDigest = digestBytes([]byte(strings.Join([]string{
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	}, "\x00")))

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return toolchainDigest, ""
	}
	settings := make(map[string]string, len(buildInfo.Settings))
	for _, setting := range buildInfo.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["vcs.revision"] == "" || settings["vcs.modified"] != "false" {
		return toolchainDigest, ""
	}
	buildSettings := make([]string, 0, len(buildInfo.Settings))
	for key, value := range settings {
		buildSettings = append(buildSettings, key+"="+value)
	}
	sort.Strings(buildSettings)
	evaluatorIdentity := append([]string{
		"gooo/valueexecution/evaluator/v1",
		settings["vcs.revision"],
	}, buildSettings...)
	return toolchainDigest, digestBytes([]byte(strings.Join(evaluatorIdentity, "\x00")))
}
