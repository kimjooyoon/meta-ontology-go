package cliobservations

import (
	"debug/buildinfo"
	"runtime/debug"
)

// Observe the binary that the harness actually built. A transformed verifier
// workspace can retain an outer work origin without producing a clean VCS build.
func cliProducerSource(binary string) (string, error) {
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return "", err
	}
	return cliProducerSourceFromSettings(info.Settings), nil
}

func cliProducerSourceFromSettings(settings []debug.BuildSetting) string {
	revision, clean := "", false
	for _, setting := range settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			clean = setting.Value == "false"
		}
	}
	if !clean || len(revision) != 40 {
		return "UNBOUND_LOCAL_SOURCE"
	}
	return revision
}

// Work admission and compiler production have different origins. A bound build
// must match both; a derived build keeps its unbound identity. Original CI
// artifact consumers independently require the actual producer to equal the head.
func textBoundProducerMatches1906(actual, built, workHead string) bool {
	if built == "" || actual != built {
		return false
	}
	return built == "UNBOUND_LOCAL_SOURCE" || workHead == "" || built == workHead
}
